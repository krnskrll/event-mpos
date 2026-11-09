export type Role = "ADMIN";

export interface UserSummary {
  id: string;
  email: string;
  fullName: string;
  role: Role;
}

export interface TokenResponse {
  accessToken: string;
  refreshToken: string;
  expiresInSeconds: number;
  user: UserSummary;
}

export type EventStatus = "DRAFT" | "ACTIVE" | "CLOSED";

export interface EventResponse {
  id: string;
  type: "event" | "outlet";
  code: string;
  name: string;
  startDate: string | null;
  endDate: string | null;
  status: EventStatus;
}

export interface AllocationRow {
  productId: string;
  sku: string;
  name: string;
  warehouseStock: number;
  allocatedQty: number;
  soldQty: number;
  remaining: number;
}
