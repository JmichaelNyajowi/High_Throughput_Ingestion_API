import { fireEvent, render, screen, within } from "@testing-library/react";
import { expect, test } from "vitest";
import { App } from "./App";

test("Fleet provides the first keyboard skip link and an accessible pending table", () => {
  render(<App initialPath="/" />);

  const skipLink = screen.getByRole("link", { name: "Skip to fleet content" });
  const main = screen.getByRole("main");
  expect(skipLink).toHaveAttribute("href", "#main-content");
  expect(document.querySelector("a")).toBe(skipLink);
  expect(main).toHaveAttribute("id", "main-content");
  fireEvent.click(skipLink);
  expect(main).toHaveFocus();
  expect(screen.getByRole("heading", { level: 1, name: "Fleet telemetry" })).toBeVisible();
  expect(screen.getByRole("region", { name: "Fleet telemetry table" })).toHaveAccessibleDescription("Scroll horizontally to review all fleet table columns.");
  expect(screen.getByRole("table", { name: "Fleet telemetry pending table" })).toBeVisible();
  expect(within(screen.getByRole("navigation", { name: "Primary navigation" })).getByRole("link", { name: "Fleet" })).toHaveAttribute("aria-current", "page");
});

test("detail, history, and recovery destinations retain one clear heading and navigation", () => {
  const detail = render(<App initialPath="/devices/gateway-17" />);
  expect(screen.getByRole("heading", { level: 1, name: "gateway-17" })).toBeVisible();
  expect(within(screen.getByRole("navigation", { name: "Primary navigation" })).getByRole("link", { name: "Fleet" })).not.toHaveAttribute("aria-current");

  detail.unmount();
  const history = render(<App initialPath="/devices/gateway-17/history" />);
  expect(screen.getByRole("heading", { level: 1, name: "History for gateway-17" })).toBeVisible();
  expect(screen.getByRole("button", { name: "Run query" })).toBeDisabled();

  history.unmount();
  render(<App initialPath="/missing" />);
  expect(screen.getByRole("heading", { level: 1, name: "Page not found" })).toBeVisible();
  expect(screen.getByRole("link", { name: "Return to fleet" })).toHaveAttribute("href", "/");
});
