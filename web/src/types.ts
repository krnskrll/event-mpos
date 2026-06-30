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
