import { describe, expect, it } from "vitest";
import { rupiah } from "./format";

describe("rupiah", () => {
  it("groups thousands with dots and has no decimals", () => {
    expect(rupiah(49500)).toBe("Rp 49.500");
    expect(rupiah(1234567)).toBe("Rp 1.234.567");
  });

  it("rounds to the nearest rupiah", () => {
    expect(rupiah(99000.4)).toBe("Rp 99.000");
    expect(rupiah(99000.5)).toBe("Rp 99.001");
  });

  it("handles zero", () => {
    expect(rupiah(0)).toBe("Rp 0");
  });
});
