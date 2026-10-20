import { AuthShell } from "./ui/AuthShell";
import { Button } from "./ui/Button";

export function AuthNotConfiguredScreen() {
  return (
    <AuthShell>
      <div className="w-full space-y-4 rounded-2xl border border-obsidian-border bg-surface-container-low p-6 shadow-2xl sm:p-8">
        <p className="text-center font-headline-sm text-headline-sm font-semibold text-on-surface">
          Auth is not configured
        </p>
        <div className="flex items-center gap-3">
          <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary">
            <span className="material-symbols-outlined text-[20px]">key_off</span>
          </span>
          <div className="min-w-0">
            <p className="font-label-lg text-label-lg font-semibold text-on-surface">
              VITE_CLERK_PUBLISHABLE_KEY is missing
            </p>
            <p className="font-body-sm text-body-sm text-on-surface-variant">
              Sign-in cannot mount without it. The landing page works; every
              auth surface shows this card until the key is set.
            </p>
          </div>
        </div>

        <pre className="overflow-x-auto rounded-xl border border-obsidian-border bg-surface-container-lowest p-4 font-label-md text-label-md text-tertiary">
          {`# client/.env\nVITE_API_BASE_URL=http://localhost:5000/api\nVITE_CLERK_PUBLISHABLE_KEY=pk_test_…`}
        </pre>

        <ol className="space-y-1.5 pl-5 font-body-sm text-body-sm text-on-surface-variant [counter-reset:step] list-decimal marker:text-primary">
          <li>Copy the publishable key from the Clerk dashboard.</li>
          <li>Paste it into <code className="font-label-md text-label-md text-primary">client/.env</code> as shown above.</li>
          <li>Refresh — Vite picks the new value up on reload.</li>
        </ol>

        <Button onClick={() => window.location.reload()} className="w-full">
          Reload after adding the key
        </Button>
      </div>
    </AuthShell>
  );
}

interface AuthErrorScreenProps {
  message: string;
  onRetry: () => void;
}

export function AuthLoadingScreen() {
  return (
    <AuthShell>
      <div className="flex w-full items-center justify-center gap-3 rounded-2xl border border-obsidian-border bg-surface-container-low p-6 shadow-2xl sm:p-8">
        <span className="h-2 w-2 animate-pulse rounded-full bg-primary" />
        <p className="text-body-md font-body-md text-on-surface-variant">
          Checking your session…
        </p>
      </div>
    </AuthShell>
  );
}

export function AuthErrorScreen({ message, onRetry }: AuthErrorScreenProps) {
  return (
    <AuthShell>
      <div className="w-full space-y-4 rounded-2xl border border-obsidian-border bg-surface-container-low p-6 text-center shadow-2xl sm:p-8">
        <p className="text-body-md font-body-md text-error">{message}</p>
        <p className="text-body-sm font-body-sm text-outline">
          You are still signed in — this is a connection problem, not a missing
          account.
        </p>
        <Button onClick={onRetry} className="w-full">
          Try again
        </Button>
      </div>
    </AuthShell>
  );
}
