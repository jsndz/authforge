import axios from 'axios';
import { tokenStore } from '@/lib/auth/token-store';
import {
  LoginResponse,
  MessageResponse,
  OAuthTokenRequest,
  OAuthTokenResponse,
  RefreshResponse,
  SignupResponse,
  UpdateUsernameResponse,
  VerifyEmailResponse,
} from '@/lib/auth/types';

const authApi = axios.create({
  baseURL: '/api/auth',
  withCredentials: true,
});

const refreshApi = axios.create({
  baseURL: '/api/auth',
  withCredentials: true,
});

let refreshPromise: Promise<string | null> | null = null;
let onAuthFailure: (() => void) | null = null;

function getErrorMessage(data: unknown): string {
  if (!data || typeof data !== 'object') return 'An unexpected error occurred';
  const maybeError = (data as { error?: unknown }).error;
  return typeof maybeError === 'string' ? maybeError : 'An unexpected error occurred';
}

async function refreshAccessToken(): Promise<string | null> {
  if (refreshPromise) return refreshPromise;

  refreshPromise = (async () => {
    try {
      const token = tokenStore.getToken();
      const headers = token ? { Authorization: `Bearer ${token}` } : undefined;
      const { data } = await refreshApi.post<RefreshResponse>('/refresh', undefined, { headers });
      tokenStore.setToken(data.access_token);
      return data.access_token;
    } catch {
      tokenStore.clear();
      onAuthFailure?.();
      return null;
    } finally {
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

authApi.interceptors.request.use((config) => {
  const token = tokenStore.getToken();
  if (token) {
    config.headers = config.headers ?? {};
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

authApi.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (!axios.isAxiosError(error)) {
      return Promise.reject(error);
    }

    const original = error.config;
    const isUnauthorized = error.response?.status === 401;
    const isRefreshRequest = original?.url?.includes('/refresh');

    if (!original || !isUnauthorized || isRefreshRequest || (original as { _retry?: boolean })._retry) {
      return Promise.reject(error);
    }

    (original as { _retry?: boolean })._retry = true;

    const newToken = await refreshAccessToken();
    if (!newToken) {
      return Promise.reject(error);
    }

    original.headers = original.headers ?? {};
    original.headers.Authorization = `Bearer ${newToken}`;
    return authApi(original);
  }
);

export function setAuthFailureHandler(handler: (() => void) | null): void {
  onAuthFailure = handler;
}

export function parseApiError(error: unknown): { error: string; status: number } {
  if (!axios.isAxiosError(error)) {
    return { error: 'An unexpected error occurred', status: 500 };
  }

  return {
    error: getErrorMessage(error.response?.data),
    status: error.response?.status ?? 500,
  };
}

export async function signupApi(username: string, email: string, password: string): Promise<SignupResponse> {
  const { data } = await authApi.post<SignupResponse>('/signup', { username, email, password });
  return data;
}

export async function loginApi(email: string, password: string): Promise<LoginResponse> {
  const { data } = await authApi.post<LoginResponse>('/login', { email, password });
  return data;
}

export async function verifyEmailApi(token: string): Promise<VerifyEmailResponse> {
  const { data } = await authApi.get<VerifyEmailResponse>('/email/verify', {
    params: { token },
  });
  return data;
}

export async function forgotPasswordApi(email: string): Promise<MessageResponse> {
  const { data } = await authApi.post<MessageResponse>('/reset/password', { email });
  return data;
}

export async function updateUsernameApi(username: string): Promise<UpdateUsernameResponse> {
  const { data } = await authApi.patch<UpdateUsernameResponse>('/update/username', { username });
  return data;
}

export async function oauthTokenApi(payload: OAuthTokenRequest): Promise<OAuthTokenResponse> {
  const body = new URLSearchParams({
    grant_type: payload.grant_type,
    client_id: payload.client_id,
    code: payload.code,
    redirect_uri: payload.redirect_uri,
    code_verifier: payload.code_verifier,
  });

  const { data } = await authApi.post<OAuthTokenResponse>('/oauth/token', body.toString(), {
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
  });

  return data;
}

export async function logoutApi(): Promise<void> {
  await authApi.get('/logout');
}

export async function logoutAllApi(): Promise<void> {
  await authApi.get('/logout/all');
}
