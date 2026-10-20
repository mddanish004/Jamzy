import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { Logo } from "./Logo";

interface AuthShellProps {
  children: ReactNode;
}

export function AuthShell({ children }: AuthShellProps) {
  return (
    <div className="relative flex min-h-dvh w-full flex-col items-center justify-center px-5 py-16">
      <div className="pointer-events-none absolute inset-0 overflow-hidden" aria-hidden="true">
        <div className="absolute -top-32 left-1/2 h-96 w-96 -translate-x-1/2 rounded-full bg-[#fb7185]/10 blur-[130px]" />
        <div className="absolute -bottom-24 left-1/2 h-80 w-80 -translate-x-1/2 rounded-full bg-[#f43f5e]/5 blur-[110px]" />
      </div>

      <div className="relative z-10 flex w-full max-w-md flex-col items-center">
        <Link
          to="/"
          aria-label="Jamzy home"
          className="group mb-space-lg flex cursor-pointer items-center gap-space-sm transition-transform duration-300 hover:scale-105"
        >
          <Logo size={40} animated />
        </Link>

        {children}

        <div className="mt-space-lg flex w-full flex-col items-center justify-between gap-space-sm px-space-xs font-label-sm text-label-sm text-outline sm:flex-row">
          <div className="flex items-center gap-1.5 opacity-80 transition-opacity hover:opacity-100">
            <span className="material-symbols-outlined text-[14px]">lock</span>
            <span>
              Secured by <span className="font-medium text-on-surface-variant">Clerk</span>
            </span>
          </div>
          {import.meta.env.DEV && (
            <div className="flex items-center gap-2 rounded-full border border-obsidian-border bg-surface-container-low px-2.5 py-1 shadow-inner">
              <span className="relative flex h-2 w-2">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-tertiary opacity-75" />
                <span className="relative inline-flex h-2 w-2 rounded-full bg-tertiary" />
              </span>
              <span className="font-medium tracking-wide text-on-surface-variant">
                Development mode
              </span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
