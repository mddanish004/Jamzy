export function StoryBanner() {
  return (
    <section className="w-full border-y border-obsidian-border/40 bg-[#0e0e10] px-gutter py-20">
      <div className="mx-auto flex max-w-4xl flex-col items-center text-center">
        <span className="font-label-caption text-label-caption uppercase tracking-widest text-primary">
          Effortless By Design
        </span>
        <h2 className="mt-3 font-headline-lg text-headline-lg text-on-surface">
          One room. One track. Everybody on the same page.
        </h2>
        <p className="mt-4 max-w-2xl font-body-lg text-body-lg leading-relaxed text-on-surface-variant">
          Nothing to configure. No video calls. Open a room, hand out the code,
          and hit play.
        </p>
        <div className="mt-8 flex h-6 items-center gap-1" aria-hidden="true">
          <span className="h-2 w-1 animate-pulse rounded-full bg-primary/40" />
          <span className="h-4 w-1 rounded-full bg-primary/70" />
          <span className="h-6 w-1 animate-bounce rounded-full bg-primary-container" />
          <span className="h-3 w-1 rounded-full bg-primary" />
          <span className="h-5 w-1 rounded-full bg-primary/40" />
          <span className="h-2 w-1 rounded-full bg-primary-container" />
          <span className="h-4 w-1 rounded-full bg-primary/60" />
        </div>
      </div>
    </section>
  );
}
