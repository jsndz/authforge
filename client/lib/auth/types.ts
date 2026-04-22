export interface User {
  username: string;
  email: string;
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

export type CodeChallengeMethod = 'S256' | 'plain';

export interface OAuthAuthorizeRequest {
  client_id: string;
  redirect_uri: string;
  state: string;
  code_challenge: string;
  code_challenge_method: CodeChallengeMethod;
  scopes?: string;
}

export interface OAuthTokenRequest {
  grant_type: 'authorization_code';
  client_id: string;
  code: string;
  redirect_uri: string;
  code_verifier: string;
}

export interface OAuthTokenResponse {
  access_token: string;
  id_token?: string;
  token_type: 'Bearer';
  expires_in: number;
}

export interface UpdateUsernameResponse {
  id: number;
  username: string;
  email: string;
}

export interface MessageResponse {
  message: string;
}

export type ApiResult<T> =
  | { ok: true; data: T; status: number }
  | { ok: false; error: string; status: number };
