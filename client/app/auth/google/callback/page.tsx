'use client';

import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Loader as Loader2 } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import {
  clearGoogleLoginContext,
  loadGoogleLoginContext,
} from '@/lib/auth/google-login';

export default function GoogleCallbackPage() {
  const router = useRouter();
  const params = useSearchParams();
  const { oauthAuthorizeUrl, oauthToken } = useAuth();

  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function run() {
      const code = params.get('code');
      const state = params.get('state');

      if (!code || !state) {
        if (!cancelled) setError('Missing OAuth callback parameters.');
        return;
      }

      const context = loadGoogleLoginContext();
      if (!context) {
        if (!cancelled) setError('Google login session expired. Please try again.');
        return;
      }

      if (state === context.initialState) {
        const authorizeHref = oauthAuthorizeUrl({
          client_id: context.clientId,
          redirect_uri: context.redirectUri,
          state: context.pkceState,
          code_challenge: context.pkceChallenge,
          code_challenge_method: 'S256',
          scopes: 'openid email profile',
        });

        window.location.replace(authorizeHref);
        return;
      }

      if (state !== context.pkceState) {
        clearGoogleLoginContext();
        if (!cancelled) setError('Invalid OAuth state. Please try again.');
        return;
      }

      const tokenResult = await oauthToken({
        grant_type: 'authorization_code',
        client_id: context.clientId,
        code,
        redirect_uri: context.redirectUri,
        code_verifier: context.pkceVerifier,
      });

      if (!tokenResult.ok) {
        clearGoogleLoginContext();
        if (!cancelled) setError(tokenResult.error || 'Failed to complete Google login.');
        return;
      }

      clearGoogleLoginContext();
      router.replace('/account/username');
    }

    run();

    return () => {
      cancelled = true;
    };
  }, [oauthAuthorizeUrl, oauthToken, params, router]);

  return (
    <div className="w-full max-w-md">
      <div className="rounded-2xl border border-gray-200 bg-white p-8 shadow-sm">
        {!error ? (
          <div className="flex flex-col items-center text-center">
            <Loader2 className="h-7 w-7 animate-spin text-gray-500" />
            <h1 className="mt-4 text-xl font-semibold text-gray-900">Completing sign-in</h1>
            <p className="mt-2 text-sm text-gray-500">Finalizing your Google login session…</p>
          </div>
        ) : (
          <div className="text-center">
            <h1 className="text-xl font-semibold text-gray-900">Google sign-in failed</h1>
            <p className="mt-3 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
              {error}
            </p>
            <button
              type="button"
              onClick={() => router.replace('/auth/login')}
              className="mt-5 rounded-lg bg-gray-900 px-4 py-2.5 text-sm font-medium text-white hover:bg-gray-800"
            >
              Back to login
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
