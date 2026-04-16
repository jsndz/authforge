'use client';

import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from 'react';
import { useRouter } from 'next/navigation';
import {
  ApiResult,
  LoginResponse,
  MessageResponse,
  RefreshResponse,
  SignupResponse,
  UpdateUsernameResponse,
  User,
  VerifyEmailResponse,
} from '@/lib/auth/types';
import { tokenStore } from '@/lib/auth/token-store';

// ─── Context shape ────────────────────────────────────────────────────────────

interface AuthContextValue {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;

  signup(username: string, email: string, password: string): Promise<ApiResult<SignupResponse>>;
  login(email: string, password: string): Promise<ApiResult<LoginResponse>>;
  verifyEmail(token: string): Promise<ApiResult<VerifyEmailResponse>>;
  forgotPassword(email: string): Promise<ApiResult<MessageResponse>>;
  updateUsername(username: string): Promise<ApiResult<UpdateUsernameResponse>>;
  logout(): Promise<void>;
  logoutAll(): Promise<void>;

  /** Authenticated fetch that auto-refreshes on 401. */
  authFetch(url: string, options?: RequestInit): Promise<Response>;
}

// ─── Context ──────────────────────────────────────────────────────────────────

const AuthContext = createContext<AuthContextValue | null>(null);

export function useAuthContext(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuthContext must be used within <AuthProvider>');
  return ctx;
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

async function parseResult<T>(res: Response): Promise<ApiResult<T>> {
  const text = await res.text();
  let parsed: unknown;
  try {
    parsed = JSON.parse(text);
  } catch {
    return { ok: false, error: 'Invalid response from server', status: res.status };
  }

  if (!res.ok) {
    const err = (parsed as { error?: string })?.error ?? 'An unexpected error occurred';
    return { ok: false, error: err, status: res.status };
  }

  return { ok: true, data: parsed as T, status: res.status };
}

// ─── Provider ─────────────────────────────────────────────────────────────────

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();

  const [user, setUser] = useState<User | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Ref to current access token for use inside closures without stale state.
  const tokenRef = useRef<string | null>(null);
  tokenRef.current = accessToken;

  // Deduplicates in-flight refresh calls.
  const refreshPromiseRef = useRef<Promise<string | null> | null>(null);

  // ── Persist helpers ──────────────────────────────────────────────────────

  const applySession = useCallback((token: string, u: User) => {
    tokenStore.setToken(token);
    tokenStore.setUser(u);
    setAccessToken(token);
    setUser(u);
  }, []);

  const clearSession = useCallback(() => {
    tokenStore.clear();
    setAccessToken(null);
    setUser(null);
  }, []);

  // ── Bootstrap from sessionStorage ────────────────────────────────────────

  useEffect(() => {
    const { token, user: storedUser } = tokenStore.initFromStorage();
    if (token && storedUser) {
      setAccessToken(token);
      setUser(storedUser);
    }
    setIsLoading(false);
  }, []);

  // ── Refresh token ────────────────────────────────────────────────────────

  const doRefresh = useCallback(async (): Promise<string | null> => {
    if (refreshPromiseRef.current) return refreshPromiseRef.current;

    const promise = (async () => {
      try {
        const res = await fetch('/api/auth/refresh', {
          method: 'POST',
          credentials: 'include',
          headers: tokenRef.current
            ? { Authorization: `Bearer ${tokenRef.current}` }
            : {},
        });

        if (!res.ok) return null;

        const data: RefreshResponse = await res.json();
        tokenStore.setToken(data.access_token);
        setAccessToken(data.access_token);
        return data.access_token;
      } catch {
        return null;
      } finally {
        refreshPromiseRef.current = null;
      }
    })();

    refreshPromiseRef.current = promise;
    return promise;
  }, []);

  // ── authFetch ────────────────────────────────────────────────────────────

  const authFetch = useCallback(
    async (url: string, options: RequestInit = {}): Promise<Response> => {
      const token = tokenRef.current;

      const buildHeaders = (t: string | null): Headers => {
        const h = new Headers(options.headers);
        if (t) h.set('Authorization', `Bearer ${t}`);
        if (!h.has('Content-Type')) h.set('Content-Type', 'application/json');
        return h;
      };

      const firstRes = await fetch(url, {
        ...options,
        headers: buildHeaders(token),
        credentials: 'include',
      });

      if (firstRes.status !== 401) return firstRes;

      // Attempt silent refresh then retry once.
      const newToken = await doRefresh();
      if (!newToken) {
        clearSession();
        router.replace('/auth/login');
        return firstRes;
      }

      return fetch(url, {
        ...options,
        headers: buildHeaders(newToken),
        credentials: 'include',
      });
    },
    [doRefresh, clearSession, router]
  );

  // ── Auth actions ─────────────────────────────────────────────────────────

  const signup = useCallback(
    async (username: string, email: string, password: string): Promise<ApiResult<SignupResponse>> => {
      const res = await fetch('/api/auth/signup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, email, password }),
      });
      return parseResult<SignupResponse>(res);
    },
    []
  );

  const login = useCallback(
    async (email: string, password: string): Promise<ApiResult<LoginResponse>> => {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });
      const result = await parseResult<LoginResponse>(res);
      if (result.ok) {
        applySession(result.data.access_token, {
          username: result.data.username,
          email: result.data.email,
        });
      }
      return result;
    },
    [applySession]
  );

  const verifyEmail = useCallback(
    async (token: string): Promise<ApiResult<VerifyEmailResponse>> => {
      const res = await fetch(`/api/auth/email/verify?token=${encodeURIComponent(token)}`, {
        credentials: 'include',
      });
      const result = await parseResult<VerifyEmailResponse>(res);
      if (result.ok) {
        applySession(result.data.access_token, {
          username: result.data.username,
          email: result.data.email,
        });
      }
      return result;
    },
    [applySession]
  );

  const forgotPassword = useCallback(
    async (email: string): Promise<ApiResult<MessageResponse>> => {
      const res = await fetch('/api/auth/reset/password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      });
      return parseResult<MessageResponse>(res);
    },
    []
  );

  const updateUsername = useCallback(
    async (username: string): Promise<ApiResult<UpdateUsernameResponse>> => {
      const res = await authFetch('/api/auth/update/username', {
        method: 'PATCH',
        body: JSON.stringify({ username }),
      });
      const result = await parseResult<UpdateUsernameResponse>(res);
      if (result.ok && user) {
        const updated = { ...user, username: result.data.username };
        tokenStore.setUser(updated);
        setUser(updated);
      }
      return result;
    },
    [authFetch, user]
  );

  const logout = useCallback(async () => {
    try {
      await authFetch('/api/auth/logout', { credentials: 'include' });
    } catch {}
    clearSession();
    router.replace('/auth/login');
  }, [authFetch, clearSession, router]);

  const logoutAll = useCallback(async () => {
    try {
      await authFetch('/api/auth/logout/all');
    } catch {}
    clearSession();
    router.replace('/auth/login');
  }, [authFetch, clearSession, router]);

  // ─────────────────────────────────────────────────────────────────────────

  const value: AuthContextValue = {
    user,
    isLoading,
    isAuthenticated: !!accessToken && !!user,
    signup,
    login,
    verifyEmail,
    forgotPassword,
    updateUsername,
    logout,
    logoutAll,
    authFetch,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
