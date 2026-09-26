// Package jev provides the transport and validated response mapping used by generated Go clients.
package jev

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Level associates a score position with a domain value and a rubric description.
type Level struct {
	Value       float64 `json:"value"`
	Description string  `json:"description"`
}

// Question contains the wire question and its Protobuf response mapping.
type Question struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Instructions string            `json:"instructions"`
	Field        string            `json:"field"`
	Kind         string            `json:"kind"`
	Choices      map[string]string `json:"choices,omitempty"`
	Levels       []Level           `json:"levels,omitempty"`
	Threshold    float64           `json:"threshold"`
	Oneof        map[string]string `json:"oneof,omitempty"`
}

// Questions builds a fresh API payload. Mapping information is never sent to the model.
func Questions(rules []Question) map[string]any {
	out := make(map[string]any, len(rules))
	for _, q := range rules {
		item := map[string]any{"type": q.Type, "instructions": q.Instructions}
		switch q.Type {
		case "choice":
			choices := make(map[string]any, len(q.Choices))
			for k, v := range q.Choices {
				choices[k] = v
			}
			item["criteria"] = choices
		case "score":
			levels := make([]string, len(q.Levels))
			for i, l := range q.Levels {
				levels[i] = l.Description
			}
			item["criteria"] = levels
		}
		out[q.Name] = item
	}
	return out
}

// Response retains the complete provider response, including probabilities, confidence,
// model identity, usage and backend-specific metadata. Raw is the original JSON body.
type Response struct {
	Model   string                     `json:"model"`
	Usage   map[string]any             `json:"usage"`
	Answers map[string]json.RawMessage `json:"answers"`
	Raw     json.RawMessage            `json:"-"`
}

// Evaluation contains a typed decision and its provider response.
type Evaluation[T proto.Message] struct {
	Value    T
	Response *Response
}

// Client is embedded in generated clients. Set Endpoint, Model and HTTPClient to
// configure a backend. Requests are not retried automatically; callers own retry policy.
type Client struct {
	APIKey     string
	Endpoint   string
	Model      string
	HTTPClient *http.Client
}

// Evaluate sends one request and validates every requested decision before populating out.
func (c *Client) Evaluate(ctx context.Context, state any, rules []Question, out proto.Message) (*Response, error) {
	if pm, ok := state.(proto.Message); ok {
		if !pm.ProtoReflect().IsValid() {
			return nil, fmt.Errorf("nil protobuf request")
		}
		b, err := protojson.Marshal(pm)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal proto state: %w", err)
		}
		state = json.RawMessage(b)
	}
	payload, err := json.Marshal(map[string]any{"model": c.Model, "state": state, "questions": Questions(rules)})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Jev payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	if c.HTTPClient == nil {
		return nil, fmt.Errorf("jev HTTPClient is nil")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jev request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024*1024+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read Jev response: %w", err)
	}
	if len(body) > 16*1024*1024 {
		return nil, fmt.Errorf("jev response exceeds 16 MiB")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jev API returned error status %d: %s", resp.StatusCode, body)
	}
	return Decode(body, rules, out)
}

// Decode validates canonical answers or legacy grouped answers, then uses the
// native Protobuf JSON decoder to preserve presence and language-specific types.
func Decode(body []byte, rules []Question, out proto.Message) (*Response, error) {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(body, &envelope); err != nil || envelope == nil {
		return nil, fmt.Errorf("failed to decode Jev response: expected a JSON object")
	}
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode Jev response: %w", err)
	}
	response.Raw = append(json.RawMessage(nil), body...)
	values := map[string]any{}
	for _, q := range rules {
		item, err := answer(envelope, q)
		if err != nil {
			return nil, err
		}
		invalid := func() error { return fmt.Errorf("invalid %s answer for question %q", q.Type, q.Name) }
		switch q.Type {
		case "choice":
			var label string
			if err := json.Unmarshal(item["choice"], &label); err != nil {
				return nil, invalid()
			}
			if _, ok := q.Choices[label]; !ok {
				return nil, fmt.Errorf("question %q: unknown choice %q", q.Name, label)
			}
			if len(q.Oneof) > 0 {
				switch q.Oneof[label] {
				case "bytes":
					values[label] = base64.StdEncoding.EncodeToString([]byte(label))
				case "string":
					values[label] = label
				default:
					return nil, invalid()
				}
			} else {
				values[q.Field] = label
			}
		case "noul":
			raw, ok := item["noul"]
			if !ok {
				raw = item["result"]
			}
			var value any
			if err := json.Unmarshal(raw, &value); err != nil {
				return nil, invalid()
			}
			switch v := value.(type) {
			case bool:
				values[q.Field] = v
			case float64:
				if v < 0 || v > 1 || math.IsNaN(v) || math.IsInf(v, 0) {
					return nil, invalid()
				}
				values[q.Field] = v >= q.Threshold
			default:
				return nil, invalid()
			}
		case "score":
			var value any
			if err := json.Unmarshal(item["score"], &value); err != nil {
				return nil, invalid()
			}
			pos, ok := value.(float64)
			if !ok || math.IsNaN(pos) || math.IsInf(pos, 0) || len(q.Levels) < 2 || pos < 0 || pos > float64(len(q.Levels)-1) {
				return nil, invalid()
			}
			idx := int(pos)
			val := q.Levels[idx].Value
			if idx+1 < len(q.Levels) {
				frac := pos - float64(idx)
				val = (1-frac)*val + frac*q.Levels[idx+1].Value
			}
			if strings.HasPrefix(q.Kind, "int") || strings.HasPrefix(q.Kind, "uint") {
				val = math.Round(val) // Half away from zero, identical in all generated languages.
			}
			if math.IsNaN(val) || math.IsInf(val, 0) {
				return nil, invalid()
			}
			if q.Kind == "int64" || q.Kind == "uint64" {
				values[q.Field] = strconv.FormatFloat(val, 'f', 0, 64)
			} else {
				values[q.Field] = val
			}
		default:
			return nil, invalid()
		}
	}
	data, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	if err := protojson.Unmarshal(data, out); err != nil {
		return nil, fmt.Errorf("invalid protobuf decision: %w", err)
	}
	return &response, nil
}

func answer(envelope map[string]json.RawMessage, q Question) (map[string]json.RawMessage, error) {
	raw, canonical := envelope["answers"]
	if !canonical {
		raw = envelope[map[string]string{"choice": "choices", "noul": "nouls", "score": "scores"}[q.Type]]
	}
	var group map[string]json.RawMessage
	if err := json.Unmarshal(raw, &group); err != nil {
		return nil, fmt.Errorf("missing or invalid answers for question %q", q.Name)
	}
	var item map[string]json.RawMessage
	if err := json.Unmarshal(group[q.Name], &item); err != nil || item == nil {
		return nil, fmt.Errorf("missing or invalid answer for question %q", q.Name)
	}
	if rawType, ok := item["type"]; ok || canonical {
		var kind string
		if err := json.Unmarshal(rawType, &kind); err != nil || kind != q.Type {
			return nil, fmt.Errorf("question %q: expected answer type %q", q.Name, q.Type)
		}
	}
	return item, nil
}
