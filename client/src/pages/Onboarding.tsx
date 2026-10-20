import { useState } from "react";
import { useAuth } from "@clerk/react";
import { useNavigate } from "react-router-dom";
import { createProfile } from "../services/api";
import { useCurrentUserContext } from "../hooks/useCurrentUser";
import { AuthShell } from "../components/ui/AuthShell";

export function Onboarding() {
  const { getToken } = useAuth();
  const navigate = useNavigate();
  const { applyProfile } = useCurrentUserContext();

  const [username, setUsername] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);

    const trimmed = username.trim();
    if (trimmed.length < 3) {
      setError("Username must be at least 3 characters");
      return;
    }
    if (trimmed.length > 20) {
      setError("Username must be at most 20 characters");
      return;
    }
    if (!/^[a-zA-Z0-9_]+$/.test(trimmed)) {
      setError("Username can only contain letters, numbers, and underscores");
      return;
    }

    setSubmitting(true);
    try {
      const { user } = await createProfile(getToken, trimmed);
      applyProfile(user);
      navigate("/home", { replace: true });
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Something went wrong";
      setError(message);
    } finally {
      setSubmitting(false);
    }
  }

  const ready =
    /^[a-zA-Z0-9_]{3,20}$/.test(username.trim());

  return (
    <AuthShell>
      <div className="relative w-full overflow-hidden rounded-xl border border-obsidian-border bg-[#1c1b1d] p-space-lg shadow-2xl backdrop-blur-xl sm:p-space-xl">
        <div className="absolute inset-x-0 top-0" aria-hidden="true">
          <div className="h-[2px] w-full rounded-t-xl bg-gradient-to-r from-transparent via-[#fb7185]/40 to-transparent" />
        </div>

        <div className="mb-space-lg text-center">
          <h1 className="mb-space-xs font-headline-md text-headline-md tracking-tight text-on-surface">
            Choose your username
          </h1>
          <p className="font-body-md text-body-md text-outline">
            This is how others will find you on Jamzy
          </p>
        </div>

        <form className="flex flex-col gap-space-md" onSubmit={handleSubmit}>
          <div className="flex flex-col gap-space-xs text-left">
            <label
              htmlFor="username"
              className="flex items-center justify-between font-label-md text-label-md text-on-surface-variant"
            >
              <span>Username</span>
              <span
                className={`font-label-sm text-label-sm ${
                  ready ? "text-tertiary" : "text-outline"
                }`}
              >
                {username.trim().length > 0
                  ? ready
                    ? "Ready"
                    : "3–20 letters, numbers, _"
                  : ""}
              </span>
            </label>
            <div className="relative flex items-center">
              <input
                id="username"
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="afrobeatfan"
                autoComplete="off"
                autoFocus
                maxLength={20}
                disabled={submitting}
                className="h-12 w-full rounded-lg border border-[#2C2B36]/30 bg-surface-container-lowest px-4 pr-11 text-body-md font-body-md text-on-surface shadow-inner placeholder:text-outline/60 outline-none transition-all duration-200 focus:border-primary focus:ring-1 focus:ring-primary disabled:opacity-50"
              />
              <span className="pointer-events-none absolute right-3 text-[18px] text-outline transition-colors">
                <span
                  className={`material-symbols-outlined text-[18px] ${
                    ready ? "text-tertiary" : "text-outline"
                  }`}
                >
                  {ready ? "check_circle" : "badge"}
                </span>
              </span>
            </div>
          </div>

          {error && <p className="text-body-sm font-body-sm text-error">{error}</p>}

          <button
            type="submit"
            disabled={submitting || !ready}
            className="group flex cursor-pointer items-center justify-center gap-space-xs rounded-lg bg-gradient-to-r from-[#fb7185] via-[#f43f5e] to-[#fb7185] px-space-md py-3 font-semibold font-headline-sm text-headline-sm text-[#400010] shadow-[0_4px_24px_-4px_rgba(251,113,133,0.35)] transition-all duration-200 hover:shadow-[0_4px_32px_0px_rgba(251,113,133,0.5)] hover:brightness-110 active:scale-[0.99] disabled:opacity-60 disabled:pointer-events-none"
          >
            {submitting ? (
              <>
                <span className="material-symbols-outlined animate-spin text-base">
                  progress_activity
                </span>
                <span>Creating…</span>
              </>
            ) : (
              <>
                <span>Continue</span>
                <span className="material-symbols-outlined text-base transition-transform duration-200 group-hover:translate-x-1">
                  arrow_forward
                </span>
              </>
            )}
          </button>
        </form>
      </div>
    </AuthShell>
  );
}
