export interface GoogleLoginContext {
  clientId: string;
  redirectUri: string;
  initialState: string;
  pkceState: string;
  pkceVerifier: string;
  pkceChallenge: string;
  createdAt: number;
}

const STORAGE_KEY = 'auth_google_login_ctx';
const MAX_CONTEXT_AGE_MS = 10 * 60 * 1000;

function toBase64Url(bytes: Uint8Array): string {
  let binary = '';
  bytes.forEach((b) => {
    binary += String.fromCharCode(b);
  });

  return btoa(binary)
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/g, '');
}

function randomBytes(length: number): Uint8Array {
  const bytes = new Uint8Array(length);
  crypto.getRandomValues(bytes);
  return bytes;
}

export function createRandomState(length = 32): string {
  return toBase64Url(randomBytes(length));
}

export async function createPkcePair(): Promise<{ verifier: string; challenge: string; method: 'S256' }> {
  const verifier = toBase64Url(randomBytes(64));
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(verifier));
  const challenge = toBase64Url(new Uint8Array(digest));
  return { verifier, challenge, method: 'S256' };
}

export function saveGoogleLoginContext(ctx: GoogleLoginContext): void {
  if (typeof window === 'undefined') return;
  window.sessionStorage.setItem(STORAGE_KEY, JSON.stringify(ctx));
}

export function loadGoogleLoginContext(): GoogleLoginContext | null {
  if (typeof window === 'undefined') return null;

  const raw = window.sessionStorage.getItem(STORAGE_KEY);
  if (!raw) return null;

  try {
    const parsed = JSON.parse(raw) as GoogleLoginContext;
    if (!parsed?.createdAt || Date.now() - parsed.createdAt > MAX_CONTEXT_AGE_MS) {
      clearGoogleLoginContext();
      return null;
    }
    return parsed;
  } catch {
    clearGoogleLoginContext();
    return null;
  }
}

export function clearGoogleLoginContext(): void {
  if (typeof window === 'undefined') return;
  window.sessionStorage.removeItem(STORAGE_KEY);
}
