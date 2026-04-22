'use client';

import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
} from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import {
  ApiResult,
  LoginResponse,
  MessageResponse,
  OAuthAuthorizeRequest,
  OAuthTokenRequest,
  OAuthTokenResponse,
  SignupResponse,
  UpdateUsernameResponse,
  User,
  VerifyEmailResponse,
} from '@/lib/auth/types';
import { tokenStore } from '@/lib/auth/token-store';
import {
  forgotPasswordApi,
  loginApi,
  logoutAllApi,
  logoutApi,
  oauthTokenApi,
  parseApiError,
  setAuthFailureHandler,
  signupApi,
  updateUsernameApi,
  verifyEmailApi,
} from '@/lib/auth/api';

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
  oauthAuthorizeUrl(params: OAuthAuthorizeRequest): string;
  oauthToken(payload: OAuthTokenRequest): Promise<ApiResult<OAuthTokenResponse>>;
  logout(): Promise<void>;
  logoutAll(): Promise<void>;
}

// ─── Context ──────────────────────────────────────────────────────────────────

const AuthContext = createContext<AuthContextValue | null>(null);

export function useAuthContext(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuthContext must be used within <AuthProvider>');
  return ctx;
}

// ─── Provider ─────────────────────────────────────────────────────────────────

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const queryClient = useQueryClient();

  const sessionQuery = useQuery({
    queryKey: ['auth', 'session'],
    queryFn: async () => {
      const { token, user } = tokenStore.initFromStorage();
      return { token, user };
    },
    staleTime: Infinity,
    gcTime: Infinity,
  });

  const signupMutation = useMutation({
    mutationFn: ({ username, email, password }: { username: string; email: string; password: string }) =>
      signupApi(username, email, password),
  });

  const loginMutation = useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) => loginApi(email, password),
  });

  const verifyEmailMutation = useMutation({
    mutationFn: (token: string) => verifyEmailApi(token),
  });

  const forgotPasswordMutation = useMutation({
    mutationFn: (email: string) => forgotPasswordApi(email),
  });

  const updateUsernameMutation = useMutation({
    mutationFn: (username: string) => updateUsernameApi(username),
  });

  const oauthTokenMutation = useMutation({
    mutationFn: (payload: OAuthTokenRequest) => oauthTokenApi(payload),
  });

  const logoutMutation = useMutation({
    mutationFn: () => logoutApi(),
  });

  const logoutAllMutation = useMutation({
    mutationFn: () => logoutAllApi(),
  });

  // ── Persist helpers ──────────────────────────────────────────────────────

  const applySession = useCallback((token: string, u: User) => {
    tokenStore.setToken(token);
    tokenStore.setUser(u);
    queryClient.setQueryData(['auth', 'session'], { token, user: u });
  }, [queryClient]);

  const clearSession = useCallback(() => {
    tokenStore.clear();
    queryClient.setQueryData(['auth', 'session'], { token: null, user: null });
  }, [queryClient]);

  useEffect(() => {
    setAuthFailureHandler(clearSession);
    return () => setAuthFailureHandler(null);
  }, [clearSession]);

  const toResult = useCallback(<T,>(fn: () => Promise<T>): Promise<ApiResult<T>> => {
    return fn()
      .then((data) => ({ ok: true, data, status: 200 } as const))
      .catch((error) => {
        const parsed = parseApiError(error);
        return { ok: false, error: parsed.error, status: parsed.status } as const;
      });
  }, []);

  // ── Auth actions ─────────────────────────────────────────────────────────

  const signup = useCallback(
    async (username: string, email: string, password: string): Promise<ApiResult<SignupResponse>> => {
      return toResult(() => signupMutation.mutateAsync({ username, email, password }));
    },
    [signupMutation, toResult]
  );

  const login = useCallback(
    async (email: string, password: string): Promise<ApiResult<LoginResponse>> => {
      const result = await toResult(() => loginMutation.mutateAsync({ email, password }));
      if (result.ok) {
        applySession(result.data.access_token, {
          username: result.data.username,
          email: result.data.email,
        });
      }
      return result;
    },
    [applySession, loginMutation, toResult]
  );

  const verifyEmail = useCallback(
    async (token: string): Promise<ApiResult<VerifyEmailResponse>> => {
      const result = await toResult(() => verifyEmailMutation.mutateAsync(token));
      if (result.ok) {
        applySession(result.data.access_token, {
          username: result.data.username,
          email: result.data.email,
        });
      }
      return result;
    },
    [applySession, toResult, verifyEmailMutation]
  );

  const forgotPassword = useCallback(
    async (email: string): Promise<ApiResult<MessageResponse>> => {
      return toResult(() => forgotPasswordMutation.mutateAsync(email));
    },
    [forgotPasswordMutation, toResult]
  );

  const updateUsername = useCallback(
    async (username: string): Promise<ApiResult<UpdateUsernameResponse>> => {
      const result = await toResult(() => updateUsernameMutation.mutateAsync(username));
      if (result.ok) {
        const currentUser = sessionQuery.data?.user;
        if (currentUser) {
          const updated = { ...currentUser, username: result.data.username };
          tokenStore.setUser(updated);
          queryClient.setQueryData(['auth', 'session'], {
            token: sessionQuery.data?.token ?? null,
            user: updated,
          });
        }
      }
      return result;
    },
    [queryClient, sessionQuery.data, toResult, updateUsernameMutation]
  );

  const oauthAuthorizeUrl = useCallback((params: OAuthAuthorizeRequest): string => {
    const search = new URLSearchParams({
      client_id: params.client_id,
      redirect_uri: params.redirect_uri,
      state: params.state,
      code_challenge: params.code_challenge,
      code_challenge_method: params.code_challenge_method,
      ...(params.scopes ? { scopes: params.scopes } : {}),
    });
    return `/api/auth/oauth/authorize?${search.toString()}`;
  }, []);

  const oauthToken = useCallback(
    async (payload: OAuthTokenRequest): Promise<ApiResult<OAuthTokenResponse>> => {
      return toResult(() => oauthTokenMutation.mutateAsync(payload));
    },
    [oauthTokenMutation, toResult]
  );

  const logout = useCallback(async () => {
    try {
      await logoutMutation.mutateAsync();
    } catch {}
    clearSession();
    router.replace('/auth/login');
  }, [clearSession, logoutMutation, router]);

  const logoutAll = useCallback(async () => {
    try {
      await logoutAllMutation.mutateAsync();
    } catch {}
    clearSession();
    router.replace('/auth/login');
  }, [clearSession, logoutAllMutation, router]);

  // ─────────────────────────────────────────────────────────────────────────

  const user = sessionQuery.data?.user ?? null;
  const accessToken = sessionQuery.data?.token ?? null;

  const value: AuthContextValue = {
    user,
    isLoading: sessionQuery.isLoading,
    isAuthenticated: !!accessToken && !!user,
    signup,
    login,
    verifyEmail,
    forgotPassword,
    updateUsername,
    oauthAuthorizeUrl,
    oauthToken,
    logout,
    logoutAll,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
