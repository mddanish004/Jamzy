export function EmotionalBanner() {
  return (
    <section className="w-full border-y border-obsidian-border/40 bg-gradient-to-b from-[#0e0e10] to-[#141418] px-gutter py-24">
      <div className="mx-auto flex max-w-4xl flex-col items-center text-center">
        <div className="mb-6 grid size-16 place-items-center rounded-full border border-primary/30 bg-primary/20 text-primary shadow-xl shadow-primary/20">
          <span className="material-symbols-outlined text-[32px] msr-fill">
            favorite
          </span>
        </div>
        <h2 className="max-w-xl font-headline-lg text-headline-lg text-on-surface">
          Songs hit harder when shared.
        </h2>
        <p className="mt-4 max-w-2xl font-body-lg text-body-lg leading-relaxed text-on-surface-variant">
          Being present isn’t staring at faces through a webcam; it’s the exact
          same bassline hitting everyone at the very same instant.
        </p>
        <div className="mt-8 flex flex-wrap items-center justify-center gap-x-6 gap-y-2 font-label-ui text-label-ui text-on-surface-variant">
          <span className="flex items-center gap-2">
            <span className="h-2 w-2 rounded-full bg-tertiary" />Free, No Strings
          </span>
          <span className="flex items-center gap-2">
            <span className="h-2 w-2 rounded-full bg-primary" />Nothing Tracked
          </span>
          <span className="flex items-center gap-2">
            <span className="h-2 w-2 rounded-full bg-primary-container" />
            One-Tap Invites
          </span>
        </div>
      </div>
    </section>
  );
}
