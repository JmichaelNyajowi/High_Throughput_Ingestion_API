import { expect, test } from "vitest";
import { bootstrapTitle } from "./App";

test("exposes the approved foundation status", () => {
  expect(bootstrapTitle).toBe("Application foundation ready");
});
