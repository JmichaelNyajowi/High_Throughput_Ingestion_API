import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { FleetLivePage } from "./index";

const fleet = { mode:"live", generated_at:"2026-09-12T10:00:00Z", summary:{critical:1,warning:0,stale:0,active:1}, devices:[{device_id:"edge-02",freshness:"active",overall_status:"critical",last_processed_at:"2026-09-12T10:00:00Z",measurements:[]},{device_id:"edge-01",freshness:"stale",overall_status:"normal",last_processed_at:"2026-09-12T09:00:00Z",measurements:[]}] };
function renderFleet() { return render(<QueryClientProvider client={new QueryClient({defaultOptions:{queries:{retry:false}}})}><FleetLivePage/></QueryClientProvider>); }
afterEach(()=>vi.unstubAllGlobals());
test("filters, sorts, and links Redis fleet snapshots", async()=>{ vi.stubGlobal("fetch",vi.fn().mockResolvedValue({ok:true,json:async()=>fleet})); renderFleet(); await screen.findByText("edge-02"); fireEvent.change(screen.getByLabelText("Search fleet"),{target:{value:"edge-01"}}); expect(screen.getByText("edge-01")).toBeVisible(); expect(screen.queryByText("edge-02")).toBeNull(); fireEvent.click(screen.getByRole("button",{name:"Clear filters"})); await waitFor(()=>expect(screen.getByText("edge-02")).toBeVisible()); expect(screen.getByRole("link",{name:"edge-02"})).toHaveAttribute("href","/devices/edge-02"); expect(screen.getByRole("button",{name:"Last processed"})).toBeVisible(); });
test("shows an explicit unavailable state without history fetch", async()=>{ vi.stubGlobal("fetch",vi.fn().mockResolvedValue({ok:false})); renderFleet(); expect(await screen.findByText("Live data unavailable")).toBeVisible(); expect(fetch).toHaveBeenCalledWith("/v1/live/fleet"); });
