import type { EventResponse } from "./types";

/** Same rule as the backend: an event takes allocations while DRAFT, an outlet until it closes. */
export function canAllocate(loc: Pick<EventResponse, "type" | "status">): boolean {
  return loc.type === "outlet" ? loc.status !== "CLOSED" : loc.status === "DRAFT";
}
