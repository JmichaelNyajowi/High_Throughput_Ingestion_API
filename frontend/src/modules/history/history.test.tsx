import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { HistoryPreQueryPage } from "./index";

afterEach(() => vi.unstubAllGlobals());

test("does not query history until a valid range is explicitly submitted", async () => {
  const fetchMock = vi.fn().mockResolvedValue({ ok: true, json: async () => ({ request_id: "request-01", events: [], limit: 100 }) });
  vi.stubGlobal("fetch", fetchMock);
  render(<HistoryPreQueryPage deviceID="edge-01" />);

  expect(fetchMock).not.toHaveBeenCalled();
  fireEvent.change(screen.getByLabelText("Start time"), { target: { value: "2026-09-12T09:00" } });
  fireEvent.change(screen.getByLabelText("End time"), { target: { value: "2026-09-12T10:00" } });
  fireEvent.click(screen.getByRole("button", { name: "Run query" }));

  expect(await screen.findByText("No events in this range")).toBeVisible();
  const requestURL = new URL(fetchMock.mock.calls[0][0], "https://telemetry.test");
  expect(requestURL.pathname).toBe("/v1/history/devices/edge-01");
  expect(requestURL.searchParams.get("from")).toBe(new Date("2026-09-12T09:00").toISOString());
  expect(requestURL.searchParams.get("to")).toBe(new Date("2026-09-12T10:00").toISOString());
  expect(requestURL.searchParams.get("limit")).toBe("100");
});
