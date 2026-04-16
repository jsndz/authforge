'use client';

import { useState, useRef, useEffect } from 'react';
import Link from 'next/link';
import { LogOut, Settings, ChevronDown, User } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';

export function AccountMenu() {
  const { user, logout, logoutAll } = useAuth();
  const [open, setOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  if (!user) return null;

  return (
    <div className="relative" ref={menuRef}>
      <button
        onClick={() => setOpen((o) => !o)}
        className="flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 transition-colors"
      >
        <span className="flex h-7 w-7 items-center justify-center rounded-full bg-gray-200 text-xs font-semibold uppercase text-gray-600">
          {user.username.charAt(0)}
        </span>
        <span className="hidden sm:block max-w-[120px] truncate">{user.username}</span>
        <ChevronDown className="h-4 w-4 text-gray-400" />
      </button>

      {open && (
        <div className="absolute right-0 mt-1 w-56 rounded-xl border border-gray-200 bg-white shadow-lg z-50 overflow-hidden">
          <div className="border-b border-gray-100 px-4 py-3">
            <p className="text-sm font-semibold text-gray-900 truncate">{user.username}</p>
            <p className="text-xs text-gray-500 truncate">{user.email}</p>
          </div>

          <div className="py-1">
            <Link
              href="/account/username"
              onClick={() => setOpen(false)}
              className="flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
            >
              <Settings className="h-4 w-4 text-gray-400" />
              Update Username
            </Link>
          </div>

          <div className="border-t border-gray-100 py-1">
            <button
              onClick={() => { setOpen(false); logout(); }}
              className="flex w-full items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
            >
              <LogOut className="h-4 w-4 text-gray-400" />
              Sign out
            </button>
            <button
              onClick={() => { setOpen(false); logoutAll(); }}
              className="flex w-full items-center gap-3 px-4 py-2.5 text-sm text-red-600 hover:bg-red-50 transition-colors"
            >
              <LogOut className="h-4 w-4 text-red-400" />
              Sign out all sessions
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
