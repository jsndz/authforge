export interface User {
  username: string;
  email: string;
}

export interface AuthState {
  user: User | null;
  accessToken: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
}

export interface LoginResponse {
  access_token: string;
  username: string;
  email: string;
}

export interface SignupResponse {
  id: number;
  username: string;
  email: string;
}

export interface VerifyEmailResponse {
  access_token: string;
  username: string;
  email: string;
}

export interface RefreshResponse {
  access_token: string;
}

export interface UpdateUsernameResponse {
  id: number;
  username: string;
  email: string;
}

export interface MessageResponse {
  message: string;
}

export interface ApiErrorResponse {
  error: string;
}

export type ApiResult<T> =
  | { ok: true; data: T; status: number }
  | { ok: false; error: string; status: number };
