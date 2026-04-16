'use client';

import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAuth } from '@/hooks/useAuth';
import { Loader as Loader2, CircleCheck as CheckCircle, User, Mail } from 'lucide-react';

const schema = z.object({
  username: z
    .string()
    .min(2, 'Username must be at least 2 characters')
    .max(30, 'Username must be at most 30 characters')
    .regex(/^[a-zA-Z0-9_]+$/, 'Username may only contain letters, numbers, and underscores'),
});

type FormValues = z.infer<typeof schema>;

export default function UpdateUsernamePage() {
  const { user, updateUsername } = useAuth();
  const [serverError, setServerError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting, isDirty },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { username: user?.username ?? '' },
  });

  // Keep form in sync with user state (e.g. after external update)
  useEffect(() => {
    if (user?.username) {
      reset({ username: user.username });
    }
  }, [user?.username, reset]);

  async function onSubmit(values: FormValues) {
    setServerError(null);
    setSuccess(false);
    const result = await updateUsername(values.username);
    if (result.ok) {
      setSuccess(true);
      reset({ username: result.data.username });
    } else {
      setServerError(result.error);
    }
  }

  return (
    <div className="space-y-6">
      {/* Page heading */}
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-gray-900">Account settings</h1>
        <p className="mt-1 text-sm text-gray-500">Manage your profile information</p>
      </div>

      {/* Profile card */}
      <div className="rounded-2xl border border-gray-200 bg-white overflow-hidden shadow-sm">
        {/* Card header */}
        <div className="border-b border-gray-100 px-6 py-5">
          <h2 className="text-base font-semibold text-gray-900">Profile</h2>
          <p className="mt-0.5 text-sm text-gray-500">Your public display information</p>
        </div>

        {/* Current info */}
        <div className="grid sm:grid-cols-2 gap-4 px-6 py-5 border-b border-gray-100">
          <div className="flex items-start gap-3">
            <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gray-100">
              <User className="h-4 w-4 text-gray-500" />
            </div>
            <div>
              <p className="text-xs font-medium uppercase tracking-wide text-gray-400">
                Current Username
              </p>
              <p className="mt-0.5 text-sm font-medium text-gray-900 break-all">
                {user?.username ?? '—'}
              </p>
            </div>
          </div>
          <div className="flex items-start gap-3">
            <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gray-100">
              <Mail className="h-4 w-4 text-gray-500" />
            </div>
            <div>
              <p className="text-xs font-medium uppercase tracking-wide text-gray-400">
                Email address
              </p>
              <p className="mt-0.5 text-sm font-medium text-gray-900 break-all">
                {user?.email ?? '—'}
              </p>
            </div>
          </div>
        </div>

        {/* Update form */}
        <div className="px-6 py-5">
          <h3 className="text-sm font-semibold text-gray-900 mb-4">Change username</h3>

          {/* Server error */}
          {serverError && (
            <div className="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
              {serverError}
            </div>
          )}

          {/* Success */}
          {success && (
            <div className="mb-4 flex items-center gap-2 rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
              <CheckCircle className="h-4 w-4 shrink-0" />
              Username updated successfully.
            </div>
          )}

          <form onSubmit={handleSubmit(onSubmit)} noValidate className="flex flex-col sm:flex-row gap-3">
            <div className="flex-1">
              <input
                type="text"
                id="username"
                autoComplete="username"
                placeholder="New username"
                {...register('username')}
                disabled={isSubmitting}
                className="w-full rounded-lg border border-gray-300 bg-white px-3.5 py-2.5 text-sm text-gray-900 placeholder-gray-400 outline-none transition-colors focus:border-gray-900 focus:ring-2 focus:ring-gray-900/10 disabled:opacity-50"
              />
              {errors.username && (
                <p className="mt-1.5 text-xs text-red-600">{errors.username.message}</p>
              )}
            </div>

            <button
              type="submit"
              disabled={isSubmitting || !isDirty}
              className="shrink-0 rounded-lg bg-gray-900 px-4 py-2.5 text-sm font-medium text-white hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-gray-900 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60 transition-colors flex items-center justify-center gap-2 sm:w-auto"
            >
              {isSubmitting && <Loader2 className="h-4 w-4 animate-spin" />}
              {isSubmitting ? 'Saving...' : 'Save changes'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
