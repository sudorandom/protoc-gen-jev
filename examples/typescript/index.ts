import { TypeSafeClient } from "@typesafe-ai/sdk";
import { GenericContainer, Wait } from "testcontainers";
import path from "path";
import { IncidentTriageJevClient } from "../gen/jev/incident/v1/incident_jev.js";

async function main() {
  console.log("==================================================");
  console.log("  protoc-gen-jev: TypeScript Client End-to-End Example");
  console.log("==================================================");

  const apiKey = process.env.TYPESAFE_API_KEY;
  const customEndpoint = process.env.JEV_ENDPOINT;
  let client: IncidentTriageJevClient;
  let container: any;

  if (customEndpoint) {
    const baseURL = customEndpoint.replace(/\/v1\/systemone\/?$/, "").replace(/\/$/, "");
    console.log(`\n[INFO] Using custom Jev/Laya endpoint: ${baseURL}`);
    const sdkClient = new TypeSafeClient({
      apiKey: apiKey ?? "local",
      baseURL,
    });
    client = new IncidentTriageJevClient(sdkClient);
  } else if (apiKey) {
    console.log("\n[INFO] Using live TypeSafe AI Jev SDK with TYPESAFE_API_KEY");
    client = new IncidentTriageJevClient();
  } else {
    const absSchemaDir = path.resolve(__dirname, "../../testdata/openapi");
    console.log("\n[INFO] Starting FauxRPC Testcontainer with Jev OpenAPI spec...");
    container = await new GenericContainer("docker.io/sudorandom/fauxrpc:v0.29.1")
      .withBindMounts([{ source: absSchemaDir, target: "/openapi" }])
      .withCommand(["run", "--schema=/openapi/typesafe-jev.yaml", "--stubs=/openapi/stubs.jev.yaml", "--addr=0.0.0.0:6660"])
      .withExposedPorts(6660)
      .withWaitStrategy(Wait.forHttp("/fauxrpc/openapi-docs/", 6660))
      .start();

    const port = container.getMappedPort(6660);
    console.log(`[INFO] FauxRPC mock server running on port ${port}`);

    const sdkClient = new TypeSafeClient({
      apiKey: "mock-api-key",
      baseURL: `http://localhost:${port}`,
    });
    client = new IncidentTriageJevClient(sdkClient);
  }

  try {
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
    batchDecisions.forEach((d, i) => {
      console.log(
        `  - Item [${i + 1}]: RoutingTarget=${d.routing_target}, ` +
          `Paging=${d.requiresImmediatePaging}, Priority=${d.priority}, ` +
          `Urgency=${d.urgencyRating}, BlastRadius=${d.blastRadiusPercentage}%`,
      );
    });

    console.log("\n✔ TypeScript End-to-End Test PASSED successfully!");
  } finally {
    if (container) {
      await container.stop();
    }
  }
}

main().catch((err) => {
  console.error("Fatal error running TypeScript example:", err);
  process.exit(1);
});
