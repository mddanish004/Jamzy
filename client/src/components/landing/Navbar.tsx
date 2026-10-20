import { Link } from "react-router-dom";
import { Logo } from "../ui/Logo";

export function Navbar() {
  return (
    <header className="fixed top-0 z-50 w-full border-b border-obsidian-border/50 bg-[#0e0e10]/85 shadow-[0_1px_12px_rgba(0,0,0,0.4)] backdrop-blur-xl">
      <div className="flex h-16 w-full items-center justify-between px-gutter">
        <div className="flex items-center gap-space-md">
          <Link to="/" aria-label="Jamzy home" className="shrink-0">
            <Logo size={32} dot={false} />
          </Link>
          <span className="hidden sm:inline-flex items-center rounded-full border border-primary/20 bg-surface-container px-space-xs py-0.5 font-label-caption text-label-caption text-primary">
            IN STEP
          </span>
        </div>

        <div className="flex items-center gap-space-md">
          <Link
            to="/enter"
            className="hidden items-center rounded-lg border border-obsidian-border/60 bg-surface-container px-space-md py-2 font-label-ui text-label-ui text-on-surface transition-colors hover:bg-surface-container-high sm:inline-flex"
          >
            Log In
          </Link>
          <Link
            to="/enter"
            aria-label="Open Jamzy"
            className="grid size-8 place-items-center rounded-full bg-gradient-to-tr from-primary-container to-primary shadow-[0_0_12px_rgba(251,113,133,0.35)]"
          >
            <span className="material-symbols-outlined text-[18px] text-white msr-fill">
              person
            </span>
          </Link>
        </div>
      </div>
    </header>
  );
}
