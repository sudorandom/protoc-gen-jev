package parser_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/sudorandom/protoc-gen-jev/internal/parser"
)

func compileProto(t *testing.T, protoSrc string) *protogen.Plugin {
	t.Helper()

	tmpDir := t.TempDir()
	protoFile := filepath.Join(tmpDir, "test.proto")
	descFile := filepath.Join(tmpDir, "test.bin")

	err := os.WriteFile(protoFile, []byte(protoSrc), 0600)
	require.NoError(t, err)

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)

	// Copy jev/v1/options.proto so buf can resolve it
	jevDir := filepath.Join(tmpDir, "jev", "v1")
	err = os.MkdirAll(jevDir, 0755)
	require.NoError(t, err)

	optionsBytes, err := os.ReadFile(filepath.Join(repoRoot, "proto", "jev", "v1", "options.proto"))
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(jevDir, "options.proto"), optionsBytes, 0600)
	require.NoError(t, err)

	bufYaml := `version: v2
modules:
  - path: .
`
	err = os.WriteFile(filepath.Join(tmpDir, "buf.yaml"), []byte(bufYaml), 0600)
	require.NoError(t, err)

	cmd := exec.Command("buf", "build", tmpDir, "-o", descFile)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "buf build failed: %s", string(out))

	data, err := os.ReadFile(descFile)
	require.NoError(t, err)

	var set descriptorpb.FileDescriptorSet
	err = proto.Unmarshal(data, &set)
	require.NoError(t, err)

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"test.proto"},
		ProtoFile:      set.File,
	}

	gen, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	return gen
}

func TestSanity_EmptyRequest(t *testing.T) {
	src := `syntax = "proto3";
package test.v1;
option go_package = "test/v1;testv1";
import "jev/v1/options.proto";

message EmptyReq {}
message TriageResp {
  bool flag = 1 [(jev.v1.field).noul = {}];
}
service TestService {
  rpc DoTest(EmptyReq) returns (TriageResp);
}`

	gen := compileProto(t, src)
	file := gen.FilesByPath["test.proto"]
	require.NotNil(t, file)
	require.Len(t, file.Services, 1)

	_, err := parser.ProcessService(file.Services[0])
	require.Error(t, err)
	require.Contains(t, err.Error(), "must have at least one input field")
}

func TestSanity_EmptyResponseDecisions(t *testing.T) {
	src := `syntax = "proto3";
package test.v1;
option go_package = "test/v1;testv1";
import "jev/v1/options.proto";

message Req {
  string id = 1;
}
message EmptyResp {
  string text_only = 1; // Not a decision primitive
}
service TestService {
  option (jev.v1.service).enabled = true;
  rpc DoTest(Req) returns (EmptyResp);
}`

	gen := compileProto(t, src)
	file := gen.FilesByPath["test.proto"]
	require.NotNil(t, file)
	require.Len(t, file.Services, 1)

	_, err := parser.ProcessService(file.Services[0])
	require.Error(t, err)
	require.Contains(t, err.Error(), "must contain at least one decision field")
}

func TestSanity_MultiplePrimitivesError(t *testing.T) {
	src := `syntax = "proto3";
package test.v1;
option go_package = "test/v1;testv1";
import "jev/v1/options.proto";

message BadField {
  int32 mixed = 1 [
    (jev.v1.field).score = { min: 1, max: 5 },
    (jev.v1.field).choice = { choices: ["A", "B"] }
  ];
}`

	gen := compileProto(t, src)
	file := gen.FilesByPath["test.proto"]
	require.NotNil(t, file)
	require.Len(t, file.Messages, 1)

	_, err := parser.ProcessMessage(file.Messages[0])
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot specify more than one Jev primitive")
}

func TestSanity_ScoreMinGreaterThanMax(t *testing.T) {
	src := `syntax = "proto3";
package test.v1;
option go_package = "test/v1;testv1";
import "jev/v1/options.proto";

message BadScore {
  int32 val = 1 [
    (jev.v1.field).score = { min: 10, max: 2 }
  ];
}`

	gen := compileProto(t, src)
	file := gen.FilesByPath["test.proto"]
	require.NotNil(t, file)
	require.Len(t, file.Messages, 1)

	_, err := parser.ProcessMessage(file.Messages[0])
	require.Error(t, err)
	require.Contains(t, err.Error(), "score min (10) cannot be greater than max (2)")
}

func TestSanity_ScoreScaleTooSmall(t *testing.T) {
	src := `syntax = "proto3";
package test.v1;
option go_package = "test/v1;testv1";
import "jev/v1/options.proto";

message BadScale {
  float val = 1 [
    (jev.v1.field).score = { scale: [1.0] }
  ];
}`

	gen := compileProto(t, src)
	file := gen.FilesByPath["test.proto"]
	require.NotNil(t, file)
	require.Len(t, file.Messages, 1)

	_, err := parser.ProcessMessage(file.Messages[0])
	require.Error(t, err)
	require.Contains(t, err.Error(), "score scale must contain at least 2 distinct values")
}

func TestSanity_ChoiceNoChoices(t *testing.T) {
	src := `syntax = "proto3";
package test.v1;
option go_package = "test/v1;testv1";
import "jev/v1/options.proto";

message BadChoice {
  string val = 1 [
    (jev.v1.field).choice = {}
  ];
}`

	gen := compileProto(t, src)
	file := gen.FilesByPath["test.proto"]
	require.NotNil(t, file)
	require.Len(t, file.Messages, 1)

	_, err := parser.ProcessMessage(file.Messages[0])
	require.Error(t, err)
	require.Contains(t, err.Error(), "choice question has no valid choices")
}

func TestSanity_NonJevServiceSkipped(t *testing.T) {
	src := `syntax = "proto3";
package test.v1;
option go_package = "test/v1;testv1";

message LoginReq {
  string username = 1;
}
message LoginResp {
  string token = 1;
}
service AuthService {
  rpc Login(LoginReq) returns (LoginResp);
}`

	gen := compileProto(t, src)
	file := gen.FilesByPath["test.proto"]
	require.NotNil(t, file)
	require.Len(t, file.Services, 1)

	spec, err := parser.ProcessService(file.Services[0])
	require.NoError(t, err)
	require.Empty(t, spec.Methods, "Non-Jev service should have 0 methods processed")
}

func TestSanity_ExplicitDisableOverridesAutoDiscovery(t *testing.T) {
	src := `syntax = "proto3";
package test.v1;
option go_package = "test/v1;testv1";
import "jev/v1/options.proto";

message Req {
  string id = 1;
}
message Resp {
  bool flag = 1 [(jev.v1.field).noul = {}];
}
service MixedService {
  rpc JevMethod(Req) returns (Resp);
  rpc NonJevMethod(Req) returns (Resp) {
    option (jev.v1.method).enabled = false;
  }
}`

	gen := compileProto(t, src)
	file := gen.FilesByPath["test.proto"]
	require.NotNil(t, file)
	require.Len(t, file.Services, 1)

	spec, err := parser.ProcessService(file.Services[0])
	require.NoError(t, err)
	require.Len(t, spec.Methods, 1)
	require.Equal(t, "JevMethod", spec.Methods[0].Name)
}

func TestSanity_ServiceExplicitDisable(t *testing.T) {
	src := `syntax = "proto3";
package test.v1;
option go_package = "test/v1;testv1";
import "jev/v1/options.proto";

message Req {
  string id = 1;
}
message Resp {
  bool flag = 1 [(jev.v1.field).noul = {}];
}
service DisabledService {
  option (jev.v1.service).enabled = false;
  rpc Method1(Req) returns (Resp);
}`

	gen := compileProto(t, src)
	file := gen.FilesByPath["test.proto"]
	require.NotNil(t, file)
	require.Len(t, file.Services, 1)

	spec, err := parser.ProcessService(file.Services[0])
	require.NoError(t, err)
	require.Empty(t, spec.Methods)
}
