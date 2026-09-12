import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { DeviceLivePage } from "./index";

const device = {
  mode: "live",
  generated_at: "2026-09-12T10:00:00Z",
  device: {
    device_id: "edge-01",
    freshness: "stale",
    overall_status: "warning",
    last_processed_at: "2026-09-12T09:58:30Z",
    measurements: [{ measurement_type: "temperature", unit: "C", latest_value: 76.2, average: 72.1, minimum: 68.4, maximum: 76.2, window_start: "2026-09-12T09:53:30Z", window_end: "2026-09-12T09:58:30Z", threshold_status: "warning" }],
  },
};

function renderDevice(client = new QueryClient({ defaultOptions: { queries: { retry: false } } })) {
  return render(<QueryClientProvider client={client}><DeviceLivePage deviceID="edge-01" /></QueryClientProvider>);
}

afterEach(() => vi.unstubAllGlobals());

test("keeps the shaped loading facts and charts visible while the live request is pending", () => {
  vi.stubGlobal("fetch", vi.fn().mockImplementation(() => new Promise(() => {})));
  renderDevice();
  expect(screen.getByRole("heading", { name: "edge-01" })).toBeVisible();
  expect(screen.getByLabelText("Device measurement facts pending")).toBeVisible();
  expect(screen.getByLabelText("Device aggregate charts pending")).toBeVisible();
});

test("renders Redis-backed device facts, aggregate text alternative, stale state, and history navigation", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => device }));
  renderDevice();
  expect(await screen.findByText("Latest 76.2 C")).toBeVisible();
  expect(screen.getByRole("heading", { name: "edge-01" })).toBeVisible();
  expect(screen.getAllByText("warning").length).toBeGreaterThan(0);
  expect(screen.getAllByText("stale").length).toBeGreaterThan(0);
  expect(screen.getByText("Stale: this device has not processed telemetry within 60 seconds.")).toBeVisible();
  expect(screen.getByText("Threshold bands: Normal, Warning, Critical. Range: min—max; line: average; point: latest.")).toBeVisible();
  expect(screen.getByText("Average 72.1 · Min 68.4 · Max 76.2 C")).toBeVisible();
  expect(screen.getByRole("link", { name: "View history" })).toHaveAttribute("href", "/devices/edge-01/history");
  expect(fetch).toHaveBeenCalledWith("/v1/live/devices/edge-01");
});

test("recovers explicitly when a device has no live Redis aggregate", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({ ...device, device: { ...device.device, freshness: "unknown", measurements: [] } }) }));
  renderDevice();
  expect(await screen.findByRole("heading", { name: "No live aggregate for this device" })).toBeVisible();
  expect(screen.getByRole("link", { name: "Return to fleet" })).toHaveAttribute("href", "/");
});

test("names Redis degradation and keeps recovery navigation visible", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false, status: 503 }));
  renderDevice();
  expect(await screen.findByRole("heading", { name: "Degraded mode: live aggregates paused" })).toBeVisible();
  expect(screen.getByText("Redis is offline. No historical data was requested.")).toBeVisible();
  expect(screen.getByRole("link", { name: "Return to fleet" })).toHaveAttribute("href", "/");
});

test("retains and labels the last successful snapshot when a refresh is degraded", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false, status: 503 }));
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  client.setQueryData(["device-live", "edge-01"], device);
  renderDevice(client);
  expect(await screen.findByText("Degraded mode: live aggregates paused (Redis offline). Showing the last successful snapshot.")).toBeVisible();
  expect(screen.getByText("Latest 76.2 C")).toBeVisible();
});

test("recovers from the documented not-found response", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false, status: 404 }));
  renderDevice();
  expect(await screen.findByRole("heading", { name: "Device not found" })).toBeVisible();
  expect(screen.getByRole("link", { name: "Return to fleet" })).toHaveAttribute("href", "/");
});
