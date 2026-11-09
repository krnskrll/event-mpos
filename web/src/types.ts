export type Role = "ADMIN" | "BA" | "WAREHOUSE";

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

export interface UserResponse extends UserSummary {
  active: boolean;
  createdAt: string;
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

export interface StockLevelRow {
  productId: string;
  sku: string;
  name: string;
  allocatedQty: number;
  soldQty: number;
  remaining: number;
  lowStock: boolean;
}

export type OrderStatus = "NEW" | "PACKING" | "READY" | "COLLECTED" | "CANCELLED";

export interface WarehouseItem {
  name: string;
  sku: string;
  qty: number;
}

export interface WarehouseOrderView {
  id: string;
  orderNumber: number;
  orderNumberDisplay: string;
  referenceId: string;
  customerName: string;
  baName: string;
  items: WarehouseItem[];
  status: OrderStatus;
  total: number;
  createdAt: string;
  stockConflict: boolean;
  lateSynced: boolean;
}

// ---- Member sync schedule ----

export interface SyncTypeView {
  type: "MEMBER";
  enabled: boolean;
  cron: string;
  updatedAt: string;
  lastSyncAt: string | null;
  lastResult: string | null;
  count: number;
  source: string | null;
  message: string | null;
}

export interface SyncOverview {
  member: SyncTypeView;
}
