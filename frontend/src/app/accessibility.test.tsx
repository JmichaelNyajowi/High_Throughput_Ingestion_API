import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { App } from "./App";

test("every operator destination retains the skip link and one page heading", () => {
  vi.stubGlobal("fetch", vi.fn().mockImplementation(() => new Promise(() => {})));
  for (const path of ["/", "/devices/edge-01", "/devices/edge-01/history", "/missing"]) {
    const view = render(<QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}><App initialPath={path} /></QueryClientProvider>);
    expect(screen.getByRole("link", { name: "Skip to fleet content" })).toBeVisible();
    expect(screen.getAllByRole("heading", { level: 1 })).toHaveLength(1);
    view.unmount();
  }
});
