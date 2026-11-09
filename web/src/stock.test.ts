import { describe, expect, it } from "vitest";
import { canAllocate } from "./stock";

describe("canAllocate", () => {
  it("allows an event only while DRAFT", () => {
    expect(canAllocate({ type: "event", status: "DRAFT" })).toBe(true);
    expect(canAllocate({ type: "event", status: "ACTIVE" })).toBe(false);
    expect(canAllocate({ type: "event", status: "CLOSED" })).toBe(false);
  });

  it("allows an outlet until it closes", () => {
    expect(canAllocate({ type: "outlet", status: "DRAFT" })).toBe(true);
    expect(canAllocate({ type: "outlet", status: "ACTIVE" })).toBe(true);
    expect(canAllocate({ type: "outlet", status: "CLOSED" })).toBe(false);
  });
});
