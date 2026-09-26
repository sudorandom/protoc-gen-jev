interface Rule {
  name: string;
  type: "choice" | "noul" | "score";
  instructions: string;
  field: string;
  kind: string;
  choices?: Record<string, string>;
  levels?: { value: number; description: string }[];
  threshold: number;
  oneof?: Record<string, string>;
}

export interface Evaluation<T> {
  value: T;
  /** Complete SDK response, including answers, probabilities, confidence, model and usage. */
  response: Record<string, unknown>;
}

function object(value: unknown, label: string): Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label}: expected an object`);
  }
  return value as Record<string, unknown>;
}

function scoreQuestion(q: Rule) {
  const descriptions = (q.levels ?? []).map(l => l.description);
  if (descriptions.length < 2) throw new Error(`${q.name}: at least two score levels are required`);
  return score(q.instructions, [descriptions[0], descriptions[1], ...descriptions.slice(2)]);
}

function questions(rules: readonly Rule[]) {
  return Object.fromEntries(rules.map(q => [q.name,
    q.type === "choice" ? choice(q.instructions, q.choices ?? {}) :
    q.type === "score" ? scoreQuestion(q) :
    noul(q.instructions)
  ]));
}

function mapResponse(response: Record<string, unknown>, rules: readonly Rule[]): JsonObject {
  const values: JsonObject = Object.create(null) as JsonObject;
  const canonical = Object.hasOwn(response, "answers");
  for (const q of rules) {
    const group = object(response[canonical ? "answers" : {choice:"choices", noul:"nouls", score:"scores"}[q.type]], q.name);
    if (!Object.hasOwn(group, q.name)) throw new Error(`Missing answer for ${q.name}`);
    const item = object(group[q.name], q.name);
    if ((canonical || Object.hasOwn(item, "type")) && item.type !== q.type) {
      throw new Error(`${q.name}: expected answer type ${q.type}`);
    }
    if (q.type === "choice") {
      const label = item.choice;
      if (typeof label !== "string" || !Object.hasOwn(q.choices ?? {}, label)) {
        throw new Error(`${q.name}: invalid choice`);
      }
      if (q.oneof) {
        if (q.oneof[label] === "bytes") {
          // Protobuf field identifiers are ASCII, so the selected label needs no Unicode conversion.
          values[label] = btoa(label);
        } else { values[label] = label; }
      } else { values[q.field] = label; }
    } else if (q.type === "noul") {
      const value = Object.hasOwn(item, "noul") ? item.noul : item.result;
      if (typeof value === "boolean") { values[q.field] = value; }
      else if (typeof value === "number" && Number.isFinite(value) && value >= 0 && value <= 1) {
        values[q.field] = value >= q.threshold;
      } else { throw new Error(`${q.name}: invalid noul`); }
    } else {
      const pos = item.score;
      const levels = q.levels ?? [];
      if (typeof pos !== "number" || !Number.isFinite(pos) || levels.length < 2 || pos < 0 || pos > levels.length - 1) {
        throw new Error(`${q.name}: invalid score`);
      }
      const i = Math.floor(pos);
      let value = levels[i].value;
      if (i + 1 < levels.length) value = (1 - (pos - i)) * value + (pos - i) * levels[i + 1].value;
      if (q.kind.startsWith("int") || q.kind.startsWith("uint")) {
        const magnitude = Math.abs(value);
        const whole = Math.floor(magnitude);
        value = Math.sign(value) * (whole + (magnitude - whole >= 0.5 ? 1 : 0));
      }
      if (!Number.isFinite(value)) throw new Error(`${q.name}: non-finite domain value`);
      values[q.field] = q.kind === "int64" || q.kind === "uint64" ? value.toFixed(0) : q.kind === "float32" ? Math.fround(value) : value;
    }
  }
  return values;
}
