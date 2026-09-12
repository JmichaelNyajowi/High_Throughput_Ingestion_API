import http from "k6/http";
import { check, sleep } from "k6";

const target = __ENV.TARGET_URL || "http://localhost";
const scenario = __ENV.SCENARIO || "smoke";
const key = __ENV.API_KEY;
const device = scenario === "single-device-rate-limit" ? "load-device-1" : `load-device-${__VU}`;
const stages = scenario === "baseline" ? [{ duration: "30s", target: 2000 }, { duration: "60s", target: 2000 }] : [{ duration: "15s", target: 10 }];

export const options = { scenarios: { ingestion: { executor: "ramping-arrival-rate", startRate: 1, timeUnit: "1s", preAllocatedVUs: scenario === "baseline" ? 2500 : 20, maxVUs: scenario === "baseline" ? 3000 : 100, stages } }, thresholds: { http_req_duration: ["p(95)<20"] } };

export default function () {
  const events = Array.from({ length: 10 }, (_, index) => ({ event_id: `${device}-${__ITER}-${index}`, device_id: device, timestamp: new Date().toISOString(), measurement_type: "temperature", value: 22.5 }));
  const response = http.post(`${target}/v1/telemetry/batches`, JSON.stringify({ events }), { headers: { "Content-Type": "application/json", "X-API-Key": key } });
  check(response, { "admitted or deliberate overload response": (r) => [202, 429, 503].includes(r.status) });
  if (scenario !== "baseline") sleep(1);
}
