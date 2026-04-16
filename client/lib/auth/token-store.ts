const TOKEN_KEY = 'auth_access_token';
const USER_KEY = 'auth_user';

let inMemoryToken: string | null = null;

function safeSessionStorage(): Storage | null {
  try {
    return typeof window !== 'undefined' ? window.sessionStorage : null;
  } catch {
    return null;
  }
}

export const tokenStore = {
  getToken(): string | null {
    if (inMemoryToken) return inMemoryToken;
    const ss = safeSessionStorage();
    if (!ss) return null;
    return ss.getItem(TOKEN_KEY);
  },

  setToken(token: string | null): void {
    inMemoryToken = token;
    const ss = safeSessionStorage();
    if (!ss) return;
    if (token) {
      ss.setItem(TOKEN_KEY, token);
    } else {
      ss.removeItem(TOKEN_KEY);
    }
  },

  getUserFromStorage(): { username: string; email: string } | null {
    const ss = safeSessionStorage();
    if (!ss) return null;
    const raw = ss.getItem(USER_KEY);
    if (!raw) return null;
    try {
      return JSON.parse(raw);
    } catch {
      return null;
    }
  },

  setUser(user: { username: string; email: string } | null): void {
    const ss = safeSessionStorage();
    if (!ss) return;
    if (user) {
      ss.setItem(USER_KEY, JSON.stringify(user));
    } else {
      ss.removeItem(USER_KEY);
    }
  },

  clear(): void {
    inMemoryToken = null;
    const ss = safeSessionStorage();
    if (!ss) return;
    ss.removeItem(TOKEN_KEY);
    ss.removeItem(USER_KEY);
  },

  initFromStorage(): { token: string | null; user: { username: string; email: string } | null } {
    const ss = safeSessionStorage();
    if (!ss) return { token: null, user: null };
    const token = ss.getItem(TOKEN_KEY);
    inMemoryToken = token;
    const user = this.getUserFromStorage();
    return { token, user };
  },
};
