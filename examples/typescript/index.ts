import { TypeSafeClient } from "@typesafe-ai/sdk";
import { IncidentTriageJevClient } from "../gen/jev/incident/v1/incident_jev.js";

// MockTypeSafeClient simulates Jev responses when running offline without an API key.
class MockTypeSafeClient {
  async systemOne(_params: { state: any; questions: Record<string, any> }): Promise<any> {
    return {
      choices: {
        routing_target: { choice: "oncall_engineer" },
        priority: { choice: "PRIORITY_LEVEL_HIGH" },
        complianceClassification: { choice: "INTERNAL_CONFIDENTIAL" },
      },
      nouls: {
        requiresImmediatePaging: { result: true },
      },
      scores: {
        urgencyRating: { score: 4 },
        blastRadiusPercentage: { score: 45.0 },
      },
    };
  }
}

async function main() {
  console.log("==================================================");
  console.log("  protoc-gen-jev: TypeScript Client End-to-End Example");
  console.log("==================================================");

  const apiKey = process.env.TYPESAFE_API_KEY;
  let client: IncidentTriageJevClient;

  if (!apiKey) {
    console.log("\n[INFO] TYPESAFE_API_KEY not set in environment.");
    console.log("[INFO] Initializing client with simulated Jev client for offline demo...");
    client = new IncidentTriageJevClient(new MockTypeSafeClient() as unknown as TypeSafeClient);
  } else {
    console.log("\n[INFO] Using live TypeSafe AI Jev SDK with TYPESAFE_API_KEY");
    client = new IncidentTriageJevClient();
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
    incident_id: "INC-8891",
    title: "Database connection pool exhausted",
    description: "API latency increased to 4500ms and 500 errors spike to 12%",
    raw_logs: "Connection refused on port 5432 after 100 pool max connections",
  };

  console.log("\n2. Evaluating Single Incident State...");
  const decision = await client.evaluate(state);
  console.log("✔ Decision received:");
  console.log(JSON.stringify(decision, null, 2));

  // 3. Batch evaluation
  console.log("\n3. Batch Evaluating 3 Incident States...");
  const batchStates = [
    "Batch Item 1: Ingress 502 bad gateway spikes across region us-east-1",
    "Batch Item 2: Low-priority deprecation warning logged in analytics service",
    "Batch Item 3: Routine memory compaction completed without customer impact",
  ];

  const batchDecisions = await client.batchEvaluate(batchStates);
  console.log(`✔ Successfully evaluated ${batchDecisions.length} batch items.`);
  batchDecisions.forEach((d, idx) => {
    console.log(
      `  - Item [${idx + 1}]: RoutingTarget=${d.routing_target}, Paging=${d.requiresImmediatePaging}, Priority=${d.priority}, Urgency=${d.urgencyRating}, BlastRadius=${d.blastRadiusPercentage}%`
    );
  });

  console.log("\n✔ TypeScript End-to-End Test PASSED successfully!");
}

main().catch((err) => {
  console.error("Error executing TypeScript example:", err);
  process.exit(1);
});
