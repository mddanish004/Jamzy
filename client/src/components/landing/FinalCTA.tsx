import { Link } from "react-router-dom";

export function FinalCTA() {
  return (
    <section className="mx-auto w-full max-w-5xl px-gutter py-20">
      <div className="relative w-full overflow-hidden rounded-3xl border border-obsidian-border bg-[#141418] p-8 text-center shadow-2xl sm:p-14">
        <div className="pointer-events-none absolute -top-24 -left-24 h-72 w-72 rounded-full bg-primary-container/20 blur-3xl" />
        <div className="pointer-events-none absolute -bottom-24 -right-24 h-72 w-72 rounded-full bg-secondary-container/20 blur-3xl" />

        <div className="relative z-10 mx-auto flex max-w-2xl flex-col items-center">
          <span className="font-label-code text-label-code uppercase tracking-widest text-primary">
            Got your headphones on?
          </span>
          <h2 className="mt-4 font-headline-lg text-headline-lg text-on-surface">
            Open a room, bring your people, hit play.
          </h2>
          <p className="mt-4 font-body-md text-body-md text-on-surface-variant">
            Your shared listening space is one click away. Sign in and you're
            all set.
          </p>
          <div className="mt-8 flex w-full flex-col items-center gap-space-md sm:w-auto sm:flex-row">
            <Link
              to="/enter"
              className="flex w-full items-center justify-center gap-2 rounded-xl border border-primary/30 bg-gradient-to-r from-[#f43f5e] to-[#fb7185] px-8 py-3.5 shadow-lg shadow-[#fb7185]/25 transition-all hover:scale-[1.02] hover:opacity-95 active:scale-[0.98] sm:w-auto"
            >
              <span className="material-symbols-outlined text-[20px] text-white">
                play_arrow
              </span>
              <span className="font-headline-sm text-headline-sm text-white">
                Start your room
              </span>
            </Link>
            <Link
              to="/home"
              className="flex w-full items-center justify-center gap-2 rounded-xl border border-obsidian-border bg-[#1c1b1d] px-8 py-3.5 font-headline-sm text-headline-sm text-on-surface transition-all hover:bg-surface-bright sm:w-auto"
            >
              <span className="material-symbols-outlined text-[20px] text-primary">
                pin
              </span>
              Enter with a code
            </Link>
          </div>
          <p className="mt-4 font-label-caption text-label-caption text-on-surface-variant/70">
            Runs smoothly on Chrome, Safari, Firefox, and mobile browsers.
          </p>
        </div>
      </div>
    </section>
  );
}
