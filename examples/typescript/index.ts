import { TypeSafeClient } from "@typesafe-ai/sdk";
import {
  IncidentTriageServiceClient,
  TriageRequest,
  TriageResponse,
} from "../gen/jev/incident/v1/incident_jev.js";

async function main() {
  console.log("==================================================");
  console.log("  protoc-gen-jev: TypeScript Client End-to-End Example");
  console.log("==================================================");

  const apiKey = process.env.TYPESAFE_API_KEY;
  const customEndpoint = process.env.JEV_ENDPOINT;
  let client: IncidentTriageServiceClient;

  if (customEndpoint) {
    const baseURL = customEndpoint.replace(/\/v1\/systemone\/?$/, "").replace(/\/$/, "");
    console.log(`\n[INFO] Using custom Jev/Laya endpoint: ${baseURL}`);
    const sdkClient = new TypeSafeClient({
      apiKey: apiKey ?? "local",
      baseURL,
    });
    client = new IncidentTriageServiceClient(sdkClient);
  } else if (apiKey) {
    console.log("\n[INFO] Using live TypeSafe AI Jev SDK with TYPESAFE_API_KEY");
    client = new IncidentTriageServiceClient();
  } else {
    console.log("\n[INFO] Using default TypeSafe AI Jev SDK client");
    client = new IncidentTriageServiceClient();
  }

  // 1. Inspect generated questions
  const questions = client.buildTriageQuestions();
  console.log(`\n1. Built Jev Questions (${Object.keys(questions).length} total):`);
  for (const [key, q] of Object.entries(questions)) {
    const crit = (q as any).criteria ? JSON.stringify((q as any).criteria) : undefined;
    console.log(`  - ${key}: ${(q as any).instructions}`);
    if (crit) {
      console.log(`      criteria: ${crit}`);
    }
  }

  // 2. Evaluate single input state
  const req: TriageRequest = {
    incident_id: "INC-8891",
    title: "Database connection pool exhausted",
    description: "API latency increased to 4500ms and 500 errors spike to 12%",
    raw_logs: "Connection refused on port 5432 after 100 pool max connections",
  };

  console.log("\n2. Evaluating Single Incident State...");
  const decision: TriageResponse = await client.triage(req);
  console.log("✔ Decision received:");
  console.log(JSON.stringify(decision, null, 2));

  // 3. Batch evaluation
  console.log("\n3. Batch Evaluating 3 Incident States...");
  const batchReqs: TriageRequest[] = [
    {
      incident_id: "INC-8892",
      title: "Ingress 502 bad gateway spikes across region us-east-1",
      description: "Edge proxy reports connection reset by peer from upstream cluster",
      raw_logs: "HTTP 502 Bad Gateway - upstream connect error or disconnect/reset before headers",
    },
    {
      incident_id: "INC-8893",
      title: "Low-priority deprecation warning logged in analytics service",
      description: "Client library using deprecated v1 query endpoint; scheduled for removal in Q3",
      raw_logs: "WARN [analytics-worker] Endpoint /v1/query is deprecated, migrate to /v2/query",
    },
    {
      incident_id: "INC-8894",
      title: "Routine memory compaction completed without customer impact",
      description: "Background compaction cycle reclaimed 4.2GB memory; latency within SLO",
      raw_logs: "INFO [compactor] Compaction cycle finished in 45s, 0 errors, 4200MB reclaimed",
    },
  ];

  const batchDecisions: TriageResponse[] = await client.batchTriage(batchReqs);
  console.log(`✔ Successfully evaluated ${batchDecisions.length} batch items.`);
  batchDecisions.forEach((d, i) => {
    console.log(
      `  - Item [${i + 1}]: RoutingTarget=${d.routing_target}, ` +
        `Paging=${d.requiresImmediatePaging}, Priority=${d.priority}, ` +
        `Urgency=${d.urgencyRating}, BlastRadius=${d.blastRadiusPercentage}%`,
    );
  });

  console.log("\n✔ TypeScript End-to-End Test PASSED successfully!");
}

main().catch((err) => {
  console.error("Fatal error running TypeScript example:", err);
  process.exit(1);
});
