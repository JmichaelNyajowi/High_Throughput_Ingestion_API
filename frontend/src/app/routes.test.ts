import { expect, test } from "vitest";
import { resolveRoute } from "./routes";

test("resolves approved monitoring destinations", () => {
  expect(resolveRoute("/")).toEqual({ kind: "fleet" });
  expect(resolveRoute("/devices/gateway-17")).toEqual({ kind: "device", deviceID: "gateway-17" });
  expect(resolveRoute("/devices/gateway-17/history")).toEqual({ kind: "history", deviceID: "gateway-17" });
});

test("returns a recovery state for unsupported destinations", () => {
  expect(resolveRoute("/settings")).toEqual({ kind: "not-found" });
});
