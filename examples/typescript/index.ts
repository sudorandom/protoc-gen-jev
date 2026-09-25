import { TypeSafeClient } from "@typesafe-ai/sdk";
import { RuleTestRecordJevClient } from "../../gen/jev/ai/rules/v1/rules_jev";

// MockTypeSafeClient simulates Jev responses when running offline without an API key.
class MockTypeSafeClient extends TypeSafeClient {
  constructor() {
    super({ apiKey: "mock-api-key" });
  }

  async systemOne(params: { state: any; questions: Record<string, any> }): Promise<any> {
    return {
      choices: {
        delivery_method: { choice: "push_notification" },
        executionMode: { choice: "MODE_FAST" },
        filteredMode: { choice: "MODE_BALANCED" },
        securityClearance: { choice: "TOP_SECRET" },
        decisionFlag: { choice: "APPROVE" },
      },
      nouls: {},
      scores: {
        ratingSmall: { score: 5 },
        ratingStrict: { score: 3 },
        discreteCode: { score: 50 },
        largeScale: { score: 500 },
        temperature: { score: 25.0 },
        discreteRatio: { score: 0.8 },
        customBoundedScore: { score: 30.0 },
      },
    };
  }
}

async function main() {
  console.log("==================================================");
  console.log("  protoc-gen-jev: TypeScript Client End-to-End Example");
  console.log("==================================================");

  const apiKey = process.env.TYPESAFE_API_KEY;
  let client: RuleTestRecordJevClient;

  if (!apiKey) {
    console.log("\n[INFO] TYPESAFE_API_KEY not set in environment.");
    console.log("[INFO] Initializing client with simulated Jev client for offline demo...");
    client = new RuleTestRecordJevClient(new MockTypeSafeClient());
  } else {
    console.log("\n[INFO] Using live TypeSafe AI Jev SDK with TYPESAFE_API_KEY");
    client = new RuleTestRecordJevClient();
  }

  // 1. Inspect generated questions
  const questions = client.buildQuestions();
  console.log(`\n1. Built Jev Questions (${Object.keys(questions).length} total):`);
  for (const [key, q] of Object.entries(questions)) {
    const crit = (q as any).criteria ? JSON.stringify((q as any).criteria) : undefined;
    console.log(`  - ${key}: ${(q as any).instructions}`);
    if (crit) {
      console.log(`      criteria: ${crit}`);
    }
  }

  // 2. Evaluate single input state
  const state = {
    userId: "usr_99881",
    requestType: "security_dispatch",
    priority: "critical",
    text: "Automated alert: unauthorized container execution detected in production cluster.",
  };

  console.log("\n2. Evaluating Single State...");
  const decision = await client.evaluate(state);
  console.log("✔ Decision received:");
  console.log(JSON.stringify(decision, null, 2));

  // 3. Batch evaluation
  console.log("\n3. Batch Evaluating 3 States...");
  const batchStates = [
    "Batch Item 1: Ingress latency spike detected",
    "Batch Item 2: Critical vulnerability signature matched",
    "Batch Item 3: Routine memory compaction completed",
  ];

  const batchDecisions = await client.batchEvaluate(batchStates);
  console.log(`✔ Successfully evaluated ${batchDecisions.length} batch items.`);
  batchDecisions.forEach((d, idx) => {
    console.log(
      `  - Item [${idx + 1}]: DeliveryMethod=${d.delivery_method}, RatingSmall=${d.ratingSmall}, Clearance=${d.securityClearance}`
    );
  });

  console.log("\n✔ TypeScript End-to-End Test PASSED successfully!");
}

main().catch((err) => {
  console.error("Error executing TypeScript example:", err);
  process.exit(1);
});
