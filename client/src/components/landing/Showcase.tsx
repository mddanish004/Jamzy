import { Reveal } from "../ui/Reveal";
import { ShowcaseArt } from "./ShowcaseArt";

const SHOWCASE_ART = "/album-art/kun-faya-kun.jpg";

const QUEUE_ITEMS = [
  { index: "01", title: "Chaiyya Chaiyya", meta: "Sukhwinder Singh · 05:44", active: true },
  { index: "02", title: "Kesariya", meta: "Arijit Singh · 04:28" },
  { index: "03", title: "Channa Mereya", meta: "Arijit Singh · 04:39" },
];

export function Showcase() {
  return (
    <section className="w-full border-y border-obsidian-border/40 bg-[#0e0e10] px-gutter py-24">
      <div className="mx-auto max-w-7xl">
        <Reveal delay={0}>
          <div className="mx-auto mb-16 max-w-3xl text-center">
            <span className="font-label-caption text-label-caption uppercase tracking-widest text-primary">
              Spatial Session Deck
            </span>
            <h2 className="mt-2 font-headline-lg text-headline-lg text-on-surface">
              All the tools for one shared listen
            </h2>
            <p className="mt-4 font-body-lg text-body-lg text-on-surface-variant">
              A lightweight, distraction-free space with group tools woven into
              every corner.
            </p>
          </div>
        </Reveal>

        <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
          <Reveal delay={100}>
            <div className="flex flex-col justify-between rounded-2xl border border-obsidian-border bg-[#141418] p-6 sm:p-8 lg:col-span-7">
              <div className="mb-6 flex items-center justify-between">
                <div className="flex items-center gap-space-sm">
                  <span className="material-symbols-outlined text-[24px] text-primary">
                    play_circle
                  </span>
                  <span className="font-headline-sm text-headline-sm text-on-surface">
                    Instant-Response Player
                  </span>
                </div>
                <span className="rounded-full border border-obsidian-border bg-surface-bright px-3 py-1 font-label-caption text-label-caption text-on-surface">
                  Core Engine
                </span>
              </div>

              <div className="relative w-full overflow-hidden rounded-xl border border-obsidian-border/50 bg-surface-container p-6">
                <div className="flex flex-col items-center gap-6 sm:flex-row">
                  <ShowcaseArt
                    src={SHOWCASE_ART}
                    alt="Rockstar soundtrack cover — illustrated portrait of a long-haired rocker engulfed in flames"
                    className="h-32 w-32 rounded-xl border border-obsidian-border/60 object-cover shadow-2xl"
                  />
                  <div className="flex-1 min-w-0 text-center sm:text-left">
                    <span className="font-label-caption text-label-caption uppercase tracking-widest text-primary">
                      Group Stream
                    </span>
                    <h4 className="truncate font-headline-sm text-headline-sm text-on-surface">
                      Kun Faya Kun
                    </h4>
                    <p className="font-body-md text-body-md text-on-surface-variant">
                      A.R. Rahman · Rockstar
                    </p>
                    <div className="mt-4 flex h-6 items-center justify-center gap-1 sm:justify-start" aria-hidden="true">
                      {[
                        { color: "bg-primary-container", h: "h-2" },
                        { color: "bg-primary", h: "h-5" },
                        { color: "bg-primary-container", h: "h-6" },
                        { color: "bg-primary", h: "h-3" },
                        { color: "bg-primary-container", h: "h-4" },
                        { color: "bg-primary", h: "h-1" },
                        { color: "bg-primary-container", h: "h-5" },
                        { color: "bg-primary", h: "h-3" },
                      ].map((bar, i) => (
                        <span
                          key={i}
                          className={`w-1 rounded-full ${bar.color} ${bar.h} jamzy-eq-bar`}
                          style={{ animationDelay: `${i * -0.18}s` }}
                        />
                      ))}
                    </div>
                  </div>
                </div>
              </div>

              <div className="mt-6 grid grid-cols-1 gap-4 text-center sm:grid-cols-3">
                {[
                  { label: "Drift Gap", value: "< 2s", tone: "text-tertiary" },
                  { label: "Recovery", value: "Auto", tone: "text-on-surface" },
                  { label: "Signaling", value: "Socket.IO", tone: "text-primary" },
                ].map((tile) => (
                  <div
                    key={tile.label}
                    className="rounded-xl border border-obsidian-border/50 bg-surface-container p-3"
                  >
                    <span className="block font-label-caption text-label-caption text-on-surface-variant">
                      {tile.label}
                    </span>
                    <span
                      className={`mt-1 block font-headline-sm text-headline-sm font-bold ${tile.tone}`}
                    >
                      {tile.value}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </Reveal>

          <Reveal delay={220}>
            <div className="flex flex-col justify-between rounded-2xl border border-obsidian-border bg-[#141418] p-6 sm:p-8 lg:col-span-5">
              <div>
                <div className="mb-4 flex items-center justify-between">
                  <div className="flex items-center gap-space-sm">
                    <span className="material-symbols-outlined text-[24px] text-primary">
                      forum
                    </span>
                    <span className="font-headline-sm text-headline-sm text-on-surface">
                      In-Room Chat
                    </span>
                  </div>
                  <span className="size-2 rounded-full bg-tertiary shadow-[0_0_8px_#10b981]" />
                </div>
                <p className="mb-4 font-body-sm text-body-sm text-on-surface-variant">
                  Quiet inline messaging that keeps the music front and center.
                </p>

                <div className="space-y-3">
                  <div className="flex items-start gap-3 rounded-xl border border-obsidian-border/40 bg-surface-container p-3 text-on-surface">
                    <div className="grid size-6 shrink-0 place-items-center rounded-full border border-primary/30 bg-primary/20 text-[11px] font-bold text-primary">
                      D
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center justify-between">
                        <span className="font-label-ui text-label-ui font-semibold text-on-surface">
                          Danish
                        </span>
                        <span className="font-label-caption text-label-caption text-on-surface-variant">
                          02:14
                        </span>
                      </div>
                      <p className="mt-0.5 font-body-sm text-body-sm text-on-surface-variant">
                        That qawwali build at 01:45 sends shivers every single time…
                      </p>
                    </div>
                  </div>
                  <div className="flex items-start gap-3 rounded-xl border border-obsidian-border/40 bg-surface-container-high p-3 text-on-surface">
                    <div className="grid size-6 shrink-0 place-items-center rounded-full border border-tertiary/30 bg-tertiary/20 text-[11px] font-bold text-tertiary">
                      Z
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center justify-between">
                        <span className="font-label-ui text-label-ui font-semibold text-on-surface">
                          Zoya
                        </span>
                        <span className="font-label-caption text-label-caption text-on-surface-variant">
                          02:22
                        </span>
                      </div>
                      <p className="mt-0.5 font-body-sm text-body-sm text-on-surface-variant">
                        Chaiyya Chaiyya is up next!
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 pt-2">
                    {[
                      { glyph: "♥", count: 8 },
                      { glyph: "±", count: 12 },
                      { glyph: "🔥", count: 5 },
                    ].map((reaction) => (
                      <span
                        key={reaction.glyph}
                        className="flex cursor-pointer items-center gap-1 rounded-full border border-obsidian-border/40 bg-surface-container px-2 py-1 text-xs transition-colors hover:bg-surface-bright"
                      >
                        {reaction.glyph}{" "}
                        <span className="font-label-caption text-label-caption text-on-surface-variant">
                          {reaction.count}
                        </span>
                      </span>
                    ))}
                  </div>
                </div>
              </div>

              <div className="mt-6 flex items-center gap-2">
                <input
                  type="text"
                  readOnly
                  className="w-full rounded-lg border border-obsidian-border bg-surface-container px-4 py-2.5 font-body-sm text-body-sm text-on-surface placeholder:text-on-surface-variant/60 focus:outline-none focus:ring-1 focus:ring-primary/40"
                  placeholder="Drop a reaction for the room..."
                />
                <button
                  type="button"
                  className="grid size-10 shrink-0 place-items-center rounded-lg bg-gradient-to-tr from-primary-container to-primary text-white shadow-sm transition-transform hover:scale-105 active:scale-95"
                  aria-label="Send"
                >
                  <span className="material-symbols-outlined text-[18px]">send</span>
                </button>
              </div>
            </div>
          </Reveal>

          <Reveal delay={340}>
            <div className="rounded-2xl border border-obsidian-border bg-[#141418] p-6 sm:p-8 lg:col-span-12">
              <div className="mb-6 flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
                <div>
                  <h3 className="font-headline-sm text-headline-sm text-on-surface">
                    A Shared ‘Up Next’ Queue
                  </h3>
                  <p className="font-body-sm text-body-sm text-on-surface-variant">
                    Anyone in the room can pitch a track while the host lines the
                    picks up.
                  </p>
                </div>
                <button
                  type="button"
                  className="flex items-center gap-2 self-start rounded-lg border border-obsidian-border bg-surface-container px-4 py-2 font-label-ui text-label-ui text-on-surface transition-colors hover:bg-surface-bright sm:self-auto"
                >
                  <span className="material-symbols-outlined text-[16px] text-primary">
                    add
                  </span>
                  Queue a Track
                </button>
              </div>

              <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
                {QUEUE_ITEMS.map((item) => (
                  <div
                    key={item.index}
                    className="flex items-center justify-between rounded-xl border border-obsidian-border/50 bg-surface-container p-4"
                  >
                    <div className="flex min-w-0 items-center gap-3">
                      <span
                        className={`font-label-code text-label-code ${
                          item.active ? "text-primary" : "text-on-surface-variant"
                        }`}
                      >
                        {item.index}
                      </span>
                      <div className="truncate">
                        <p className="truncate font-body-md text-body-md font-semibold text-on-surface">
                          {item.title}
                        </p>
                        <p className="truncate font-body-sm text-body-sm text-on-surface-variant">
                          {item.meta}
                        </p>
                      </div>
                    </div>
                    <span
                      className={`material-symbols-outlined text-[20px] ${
                        item.active
                          ? "text-primary"
                          : "cursor-pointer text-on-surface-variant hover:text-primary"
                      }`}
                    >
                      thumb_up
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </Reveal>
        </div>
      </div>
    </section>
  );
}
