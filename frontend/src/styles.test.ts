import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { expect, test } from "vitest";

test("component styles consume named dimension tokens", () => {
  const styles = readFileSync(resolve(process.cwd(), "src/styles.css"), "utf8");
  const rootEnd = styles.indexOf("\n}\n\n");
  const componentRules = styles
    .slice(rootEnd + 4)
    .replace(/@media\s*\([^)]*\)/g, "")
    .replace(/@keyframes\s+\w+\s*\{\s*\d+%\s*\{[^}]*\}\s*\}\n/g, "");

  expect(componentRules).not.toMatch(/-?\d+(?:\.\d+)?(?:rem|px|ms|em|%)/);
});

test("mobile styles preserve a visible product name", () => {
  const styles = readFileSync(resolve(process.cwd(), "src/styles.css"), "utf8");

  expect(styles).toMatch(/@media \(max-width: 47\.9375rem\)[\s\S]*\.brand__name \{ display: inline; \}/);
  expect(styles).not.toMatch(/\.brand__name \{ display: none; \}/);
});

test("reduced motion and responsive table safeguards are defined", () => {
  const styles = readFileSync(resolve(process.cwd(), "src/styles.css"), "utf8");
  expect(styles).toContain("@media (prefers-reduced-motion: reduce)");
  expect(styles).toContain(".table-scroll { overflow-x: auto; }");
  expect(styles).toContain("@media (max-width: 47.9375rem)");
});
