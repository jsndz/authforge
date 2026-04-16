'use client';

import { useEffect, useRef, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { CircleCheck as CheckCircle, Circle as XCircle, Loader as Loader2, MailCheck } from 'lucide-react';
import Link from 'next/link';

type Status = 'idle' | 'loading' | 'success' | 'error' | 'missing-token';

export default function VerifyEmailPage() {
  const searchParams = useSearchParams();
  const { verifyEmail } = useAuth();
  const router = useRouter();

  const [status, setStatus] = useState<Status>('idle');
  const [errorMessage, setErrorMessage] = useState<string>('');
  const calledRef = useRef(false);

  useEffect(() => {
    if (calledRef.current) return;
    calledRef.current = true;

    const token = searchParams.get('token');
    if (!token) {
      setStatus('missing-token');
      return;
    }

    setStatus('loading');
    verifyEmail(token).then((result) => {
      if (result.ok) {
        setStatus('success');
        setTimeout(() => router.replace('/account/username'), 2000);
      } else {
        setErrorMessage(result.error);
        setStatus('error');
      }
    });
  }, [searchParams, verifyEmail, router]);

  return (
    <div className="w-full max-w-md">
      <div className="rounded-2xl border border-gray-200 bg-white p-8 shadow-sm text-center">
        {/* Loading */}
        {(status === 'idle' || status === 'loading') && (
          <>
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-gray-100">
              <Loader2 className="h-7 w-7 text-gray-600 animate-spin" />
            </div>
            <h1 className="text-xl font-semibold text-gray-900">Verifying your email</h1>
            <p className="mt-2 text-sm text-gray-500">This will just take a moment&hellip;</p>
          </>
        )}

        {/* Success */}
        {status === 'success' && (
          <>
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-emerald-50">
              <CheckCircle className="h-7 w-7 text-emerald-600" />
            </div>
            <h1 className="text-xl font-semibold text-gray-900">Email verified!</h1>
            <p className="mt-2 text-sm text-gray-500">
              Your account is confirmed. Redirecting you now&hellip;
            </p>
          </>
        )}

        {/* Missing token */}
        {status === 'missing-token' && (
          <>
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-amber-50">
              <MailCheck className="h-7 w-7 text-amber-600" />
            </div>
            <h1 className="text-xl font-semibold text-gray-900">No token provided</h1>
            <p className="mt-2 text-sm text-gray-500">
              Use the verification link from your email. The token must be included in the URL.
            </p>
            <Link
              href="/auth/login"
              className="mt-6 inline-block rounded-lg bg-gray-900 px-6 py-2.5 text-sm font-medium text-white hover:bg-gray-800 transition-colors"
            >
              Back to login
            </Link>
          </>
        )}

        {/* Error */}
        {status === 'error' && (
          <>
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-red-50">
              <XCircle className="h-7 w-7 text-red-600" />
            </div>
            <h1 className="text-xl font-semibold text-gray-900">Verification failed</h1>
            <p className="mt-2 text-sm text-gray-500">
              {errorMessage || 'The link may have expired or already been used.'}
            </p>
            <div className="mt-6 flex flex-col gap-2">
              <Link
                href="/auth/signup"
                className="inline-block rounded-lg bg-gray-900 px-6 py-2.5 text-sm font-medium text-white hover:bg-gray-800 transition-colors"
              >
                Sign up again
              </Link>
              <Link
                href="/auth/login"
                className="inline-block rounded-lg border border-gray-300 px-6 py-2.5 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
              >
                Back to login
              </Link>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
