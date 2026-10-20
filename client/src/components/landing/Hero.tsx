import { Link } from "react-router-dom";
import { ShowcaseArt } from "./ShowcaseArt";

const HERO_ART = "/album-art/tum-hi-ho.jpg";

const LISTENERS = [
  { initial: "D", name: "Danish", role: "Host · DJ", host: true },
  { initial: "S", name: "Sujal", role: "±2ms Drift" },
  { initial: "A", name: "Akshat", role: "±5ms Drift" },
  { initial: "Z", name: "Zoya", role: "±1ms Drift" },
];

export function Hero() {
  return (
    <div className="relative w-full">
      <div className="pointer-events-none absolute -top-32 left-1/2 h-[360px] w-[720px] -translate-x-1/2 rounded-full bg-primary-container/15 blur-[140px]" />
      <div className="pointer-events-none absolute top-[480px] -left-20 h-[420px] w-[420px] rounded-full bg-secondary-container/20 blur-[130px]" />

      <section className="relative z-10 mx-auto flex w-full max-w-7xl flex-col items-center px-gutter pb-16 pt-28 text-center sm:pb-20 sm:pt-32">
        <div className="inline-flex items-center gap-space-sm rounded-full border border-obsidian-border bg-surface-container-low px-4 py-1.5 shadow-sm">
          <span className="h-2 w-2 animate-pulse rounded-full bg-primary-container shadow-[0_0_10px_#fb7185]" />
          <span className="font-label-ui text-label-ui uppercase tracking-wider text-primary">
            Music is better with company
          </span>
        </div>

        <h1 className="mt-8 max-w-4xl font-display-hero-mobile text-display-hero-mobile tracking-tight text-on-surface sm:font-display-hero sm:text-display-hero">
          Play it together. Stay in{" "}
          <span className="bg-gradient-to-r from-[#fda4af] via-[#fb7185] to-[#f43f5e] bg-clip-text text-transparent drop-shadow-[0_2px_15px_rgba(251,113,133,0.3)]">
            step.
          </span>
        </h1>

        <p className="mt-6 max-w-2xl font-body-lg text-body-lg leading-relaxed text-on-surface-variant">
          Spin up a room, pass the code around, and hear every track together in
          real time — no matter where your people are.
        </p>

        <div className="mt-10 flex flex-col items-center gap-space-md sm:flex-row">
          <Link
            to="/enter"
            className="inline-flex w-full items-center justify-center gap-space-sm rounded-xl border border-primary/30 bg-gradient-to-r from-[#f43f5e] to-[#fb7185] px-8 py-3.5 shadow-lg shadow-[#fb7185]/25 transition-all hover:scale-[1.02] hover:opacity-95 active:scale-[0.98] sm:w-auto"
          >
            <span className="material-symbols-outlined text-[20px]">graphic_eq</span>
            <span className="font-headline-sm text-headline-sm text-white">
              Start a room
            </span>
          </Link>
          <Link
            to="/home"
            className="inline-flex w-full items-center justify-center gap-space-sm rounded-xl border border-obsidian-border bg-[#1c1b1d] px-8 py-3.5 shadow-sm transition-all hover:bg-surface-bright sm:w-auto"
          >
            <span className="material-symbols-outlined text-[20px] text-primary">key</span>
            <span className="font-headline-sm text-headline-sm text-on-surface">
              Enter a room
            </span>
          </Link>
        </div>

        <p className="mt-3 font-body-sm text-body-sm text-on-surface-variant/80">
          No ghost-room clutter · Free forever
        </p>

        <div className="mt-10 flex flex-wrap items-center justify-center gap-8 rounded-2xl border border-obsidian-border/60 bg-surface-container-lowest/80 px-6 py-3 shadow-lg backdrop-blur-md">
          <div className="flex items-center gap-space-sm">
            <div className="flex -space-x-2">
              <span className="grid size-6 place-items-center rounded-full border border-primary/30 bg-primary/25 text-[10px] font-bold text-primary">
                D
              </span>
              <span className="grid size-6 place-items-center rounded-full border border-tertiary/30 bg-tertiary/25 text-[10px] font-bold text-tertiary">
                S
              </span>
              <span className="grid size-6 place-items-center rounded-full border border-secondary/30 bg-secondary/25 text-[10px] font-bold text-secondary">
                Z
              </span>
            </div>
            <span className="font-label-ui text-label-ui text-on-surface">
              <strong className="font-bold text-primary">Rooms</strong> tuned in
              right now
            </span>
          </div>
          <div className="hidden h-4 w-px bg-outline-variant/30 sm:block" />
          <div className="flex items-center gap-1.5 font-label-ui text-label-ui text-on-surface">
            <div className="flex text-primary">
              {[0, 1, 2, 3, 4].map((i) => (
                <span
                  key={i}
                  className="material-symbols-outlined text-[16px] msr-fill"
                >
                  star
                </span>
              ))}
            </div>
            <span>
              <strong className={"font-bold text-on-surface"}>In rhythm</strong>{" "}
              — Because music was meant to be shared
            </span>
          </div>
        </div>

        <div className="mt-14 w-full max-w-5xl rounded-2xl border border-obsidian-border bg-[#141418] p-4 shadow-[0_20px_50px_rgba(0,0,0,0.6)] shadow-2xl sm:p-6">
          <div className="flex flex-col items-stretch gap-6 lg:flex-row">
            <div className="relative flex flex-1 flex-col justify-between overflow-hidden rounded-xl border border-obsidian-border/60 bg-[#1c1b1d] p-5 sm:p-6">
              <div className="pointer-events-none absolute -right-16 -top-16 h-56 w-56 rounded-full bg-primary-container/15 blur-3xl" />
              <div className="z-10 flex items-start justify-between">
                <div className="flex items-center gap-space-sm min-w-0">
                  <span className="flex items-center gap-1.5 rounded-full border border-tertiary/20 bg-tertiary/15 px-2.5 py-1 font-label-caption text-label-caption text-tertiary">
                    <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-tertiary" />
                    LIVE · LOCKED
                  </span>
                  <span className="truncate rounded border border-obsidian-border/40 bg-surface-container-lowest px-2 py-0.5 font-label-code text-label-code tracking-widest text-on-surface-variant/90">
                    #AMBR-990
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="material-symbols-outlined cursor-pointer text-[20px] text-on-surface-variant hover:text-on-surface">
                    airplay
                  </span>
                  <span className="material-symbols-outlined text-[20px] text-primary">
                    equalizer
                  </span>
                </div>
              </div>

              <div className="z-10 my-8 flex flex-col items-center gap-6 sm:flex-row">
                <ShowcaseArt
                  src={HERO_ART}
                  alt="Tum Hi Ho single cover — Arijit Singh singing against a warm gold backdrop with red music notes"
                  className="h-28 w-28 rounded-xl border border-obsidian-border/60 object-cover shadow-xl sm:h-36 sm:w-36"
                />
                <div className="min-w-0 text-center sm:text-left">
                  <span className="font-label-caption text-label-caption uppercase tracking-wider text-primary">
                    Now Playing For All
                  </span>
                  <h3 className="mt-1 truncate font-headline-md text-headline-md text-on-surface">
                    Tum Hi Ho
                  </h3>
                  <p className="font-body-md text-body-md text-on-surface-variant">
                    Arijit Singh · Aashiqui 2
                  </p>
                  <div className="mt-3 flex items-center justify-center gap-2 sm:justify-start">
                    <span className="rounded border border-obsidian-border bg-surface-bright px-2 py-0.5 font-label-caption text-label-caption text-on-surface">
                      FLAC 24-bit / 96kHz
                    </span>
                    <span className="flex items-center gap-1 font-label-caption text-label-caption text-tertiary">
                      <span className="material-symbols-outlined text-[14px]">
                        bolt
                      </span>
                      14ms Round Trip
                    </span>
                  </div>
                </div>
              </div>

              <div className="z-10 space-y-2">
                <div className="flex w-full items-center justify-between font-label-caption text-label-caption text-on-surface-variant">
                  <span>02:44</span>
                  <div className="flex items-center gap-1 text-primary">
                    <span className="material-symbols-outlined text-[14px]">
                      sync
                    </span>
                    <span>Following the Host</span>
                  </div>
                  <span>04:22</span>
                </div>

                <div className="relative flex h-3 w-full items-center overflow-hidden rounded-full border border-obsidian-border/50 bg-surface-container-lowest px-1">
                  <div className="relative h-1.5 w-[64%] rounded-full bg-gradient-to-r from-primary-container to-primary shadow-[0_0_10px_rgba(251,113,133,0.4)]">
                    <span className="absolute right-0 top-1/2 h-3 w-3 -translate-y-1/2 rounded-full bg-white shadow-[0_0_8px_#fb7185]" />
                  </div>
                </div>

                <div className="flex items-center justify-center gap-6 pt-3">
                  <button
                    type="button"
                    className="pointer-coarse:p-2 text-on-surface-variant transition-colors hover:text-on-surface"
                    aria-label="Shuffle"
                  >
                    <span className="material-symbols-outlined text-[24px]">
                      shuffle
                    </span>
                  </button>
                  <button
                    type="button"
                    className="pointer-coarse:p-2 text-on-surface-variant transition-colors hover:text-on-surface"
                    aria-label="Skip previous"
                  >
                    <span className="material-symbols-outlined text-[28px]">
                      skip_previous
                    </span>
                  </button>
                  <button
                    type="button"
                    className="grid size-12 place-items-center rounded-full border border-primary/40 bg-gradient-to-tr from-primary-container to-primary text-white shadow-md shadow-primary-container/30 transition-transform hover:scale-105"
                    aria-label="Pause"
                  >
                    <span className="material-symbols-outlined text-[28px] msr-fill">
                      pause
                    </span>
                  </button>
                  <button
                    type="button"
                    className="pointer-coarse:p-2 text-on-surface-variant transition-colors hover:text-on-surface"
                    aria-label="Skip next"
                  >
                    <span className="material-symbols-outlined text-[28px]">
                      skip_next
                    </span>
                  </button>
                  <button
                    type="button"
                    className="pointer-coarse:p-2 text-on-surface-variant transition-colors hover:text-on-surface"
                    aria-label="Repeat"
                  >
                    <span className="material-symbols-outlined text-[24px]">
                      repeat
                    </span>
                  </button>
                </div>
              </div>
            </div>

            <div className="flex w-full flex-col justify-between rounded-xl border border-obsidian-border/60 bg-[#1c1b1d] p-5 lg:w-72">
              <div>
                <div className="flex items-center justify-between pb-3">
                  <span className="font-headline-sm text-headline-sm text-on-surface">
                    Here Right Now
                  </span>
                  <span className="rounded-full border border-primary/25 bg-primary/20 px-2 py-0.5 font-label-caption text-label-caption text-primary">
                    4 Listening
                  </span>
                </div>

                <div className="mt-3 space-y-3">
                  {LISTENERS.map((listener) => (
                    <div
                      key={listener.name}
                      className={`flex items-center justify-between rounded-lg border p-2 ${
                        listener.host
                          ? "border-obsidian-border/40 bg-surface-container"
                          : "border-obsidian-border/30 bg-surface-container/60"
                      }`}
                    >
                      <div className="flex min-w-0 items-center gap-space-sm">
                        <div className="relative">
                          <div
                            className={`grid size-8 place-items-center rounded-full text-xs font-bold shadow-sm ${
                              listener.host
                                ? "bg-primary-container text-white"
                                : "bg-surface-bright text-on-surface"
                            }`}
                          >
                            {listener.initial}
                          </div>
                          <span className="absolute -bottom-0.5 -right-0.5 size-2.5 rounded-full bg-tertiary" />
                        </div>
                        <div className="truncate">
                          <p className="font-body-md text-body-md font-semibold leading-none text-on-surface">
                            {listener.name}
                          </p>
                          <span
                            className={`font-label-caption text-label-caption ${
                              listener.host ? "text-primary" : "text-on-surface-variant"
                            }`}
                          >
                            {listener.role}
                          </span>
                        </div>
                      </div>
                      {listener.host ? (
                        <span
                          className="material-symbols-outlined text-[18px] text-primary"
                          aria-hidden="true"
                        >
                          volume_up
                        </span>
                      ) : (
                        <span
                          className="material-symbols-outlined text-[16px] text-tertiary"
                          aria-hidden="true"
                        >
                          check_circle
                        </span>
                      )}
                    </div>
                  ))}
                </div>
              </div>

              <div className="mt-4 rounded-xl border border-obsidian-border/40 bg-surface-container-lowest p-3">
                <span className="font-label-caption text-label-caption uppercase tracking-wider text-on-surface-variant">
                  Up Next
                </span>
                <p className="mt-0.5 truncate font-body-sm text-body-sm font-semibold text-on-surface">
                  Kal Ho Naa Ho — Sonu Nigam
                </p>
                <div className="mt-1 flex items-center justify-between font-label-caption text-label-caption text-on-surface-variant">
                  <span>Queued by Sujal</span>
                  <span>05:39</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
