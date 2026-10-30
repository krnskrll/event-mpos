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

export type ProductSource = "ERP" | "MANUAL";

export interface VariantDto {
  id: string;
  name: string;
  sku: string;
  price: number;
  active: boolean;
}

export interface BarcodeDto {
  barcode: string;
  label: string | null;
}

export interface ProductResponse {
  id: string;
  sku: string;
  name: string;
  brand: string | null;
  price: number;
  imageUrl: string | null;
  source: ProductSource;
  active: boolean;
  barcodes: BarcodeDto[];
  variants: VariantDto[];
  updatedAt: string;
}

export interface EventProductResponse {
  product: ProductResponse;
  available: boolean;
}

export interface ErpSyncStatus {
  lastSyncAt: string | null;
  lastResult: string | null;
  productCount: number;
  source: string | null;
  message: string | null;
}

export type MemberTier = "STANDARD" | "SILVER" | "GOLD" | "PLATINUM";

export interface MemberResponse {
  id: string;
  memberCode: string;
  qrCode: string;
  name: string;
  tier: MemberTier;
  region: string | null;
  active: boolean;
  updatedAt: string;
}

export interface MemberSyncStatus {
  lastSyncAt: string | null;
  lastResult: string | null;
  memberCount: number;
  source: string | null;
  message: string | null;
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

export type OrderStatus = "NEW" | "PACKING" | "READY" | "COLLECTED";

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
}

// ---- EPIC-11: back office admin ----

export type SyncType = "ERP" | "MEMBER";

export interface SyncTypeView {
  type: SyncType;
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
  erp: SyncTypeView;
  member: SyncTypeView;
}

export interface ProductReportRow {
  productId: string;
  sku: string;
  name: string;
  unitsSold: number;
  revenue: number;
  allocated: number;
  sold: number;
  remaining: number;
}

export interface EndOfEventReport {
  eventId: string;
  eventCode: string;
  eventName: string;
  status: EventStatus;
  generatedAt: string;
  totalOrders: number;
  totalItems: number;
  totalRevenue: number;
  totalDiscount: number;
  products: ProductReportRow[];
}

export type AuditAction = "ORDER_CREATED" | "STATUS_CHANGED" | "RECEIPT_PRINTED";
export const AUDIT_ACTIONS: AuditAction[] = ["ORDER_CREATED", "STATUS_CHANGED", "RECEIPT_PRINTED"];

export interface AuditEntry {
  id: string;
  orderId: string;
  orderReference: string;
  action: AuditAction;
  actorName: string;
  fromStatus: string | null;
  toStatus: string | null;
  detail: string | null;
  at: string;
}

/** One page of an event's audit log, newest first (page is 0-based). */
export interface AuditPage {
  items: AuditEntry[];
  page: number;
  size: number;
  total: number;
}

export type PromotionType = "PERCENTAGE" | "FIXED" | "BUNDLE";
export type PromotionScope = "ALL" | "SPECIFIC";

export interface PromotionResponse {
  id: string;
  eventId: string;
  name: string;
  type: PromotionType;
  percentOff: number | null;
  amountOff: number | null;
  scope: PromotionScope;
  productIds: string[];
  minQty: number | null;
  minSubtotal: number | null;
  startsAt: string | null;
  endsAt: string | null;
  active: boolean;
  priority: number;
}
