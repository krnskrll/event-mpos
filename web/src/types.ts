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

export interface EventResponse {
  id: string;
  code: string;
  name: string;
  startDate: string | null;
  endDate: string | null;
  active: boolean;
}

export interface UserResponse extends UserSummary {
  active: boolean;
  createdAt: string;
}
