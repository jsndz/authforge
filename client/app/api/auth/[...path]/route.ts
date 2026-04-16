/**
 * BFF (Backend-For-Frontend) proxy for Go auth backend.
 *
 * All frontend auth calls go through /api/auth/* instead of the Go server
 * directly. This avoids CORS issues and keeps the refresh_token HTTP-only
 * cookie on the same origin as the Next.js app.
 *
 * NOTE: For HTTP-only cookies to be flagged Secure in the browser the Next.js
 * app must be served over HTTPS in production. In development (localhost)
 * Secure is not required.
 *
 * Path mapping:
 *   /api/auth/<segments>  ->  BACKEND/api/v1/auth/<segments>
 *
 * Headers forwarded to backend:
 *   Authorization  (Bearer access token from client)
 *   Cookie         (refresh_token HTTP-only cookie)
 *
 * Headers forwarded back to client:
 *   set-cookie     (rotated refresh_token from backend)
 */

import { NextRequest, NextResponse } from 'next/server';

const BACKEND = process.env.NEXT_PUBLIC_AUTH_API_URL ?? 'http://localhost:8080';

async function handler(
  req: NextRequest,
  { params }: { params: { path: string[] } }
): Promise<NextResponse> {
  const segments = params.path ?? [];
  const qs = req.nextUrl.searchParams.toString();
  const backendUrl = `${BACKEND}/api/v1/auth/${segments.join('/')}${qs ? `?${qs}` : ''}`;

  const forwardHeaders: Record<string, string> = {
    'content-type': 'application/json',
  };

  const auth = req.headers.get('authorization');
  if (auth) forwardHeaders['authorization'] = auth;

  const cookie = req.headers.get('cookie');
  if (cookie) forwardHeaders['cookie'] = cookie;

  let body: string | undefined;
  if (!['GET', 'HEAD'].includes(req.method)) {
    try {
      body = await req.text();
    } catch {
      body = undefined;
    }
  }

  let backendRes: Response;
  try {
    backendRes = await fetch(backendUrl, {
      method: req.method,
      headers: forwardHeaders,
      ...(body !== undefined ? { body } : {}),
    });
  } catch (err) {
    return NextResponse.json(
      { error: 'Backend unreachable' },
      { status: 502 }
    );
  }

  const responseText = await backendRes.text();

  const nextRes = new NextResponse(responseText, {
    status: backendRes.status,
    headers: { 'content-type': 'application/json' },
  });

  // Forward all set-cookie headers so refresh_token lands in browser.
  // node-fetch / undici collapse multiple set-cookie values into one
  // comma-separated string; split and re-append each.
  const rawCookie = backendRes.headers.get('set-cookie');
  if (rawCookie) {
    // Cookies can contain commas in expires= dates, so we split on ", "
    // only when followed by a known cookie attribute or a new cookie name.
    // Safest approach: append the raw header value as-is.
    nextRes.headers.set('set-cookie', rawCookie);
  }

  return nextRes;
}

export const GET = handler;
export const POST = handler;
export const PATCH = handler;
export const PUT = handler;
export const DELETE = handler;
