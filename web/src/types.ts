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
