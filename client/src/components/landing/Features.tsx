import { Reveal } from "../ui/Reveal";

const FEATURES = [
  {
    icon: "library_music",
    title: "One Library",
    copy: "Search YouTube together and everyone queues from the same shelf — no matching streaming plans or local files required.",
  },
  {
    icon: "sync_saved_locally",
    title: "Beat-Matched Sync",
    copy: "Drift correction holds every playhead in place, so hits, breaks, and drops all land at once.",
  },
  {
    icon: "meeting_room",
    title: "Private Rooms",
    copy: "Rooms locked behind a short code, with DJ rights held by the host and live presence for every member.",
  },
  {
    icon: "chat_bubble",
    title: "In-Moment Chat",
    copy: "Post timestamped notes and emoji pulses right as the chorus lands — everything stays tied to the moment.",
  },
  {
    icon: "search",
    title: "Track Finder",
    copy: "Type a song or paste a YouTube link and the room picks it up instantly — playlists open up into queueable tracks.",
  },
  {
    icon: "sensors",
    title: "Who's Here",
    copy: "Beacons light up when friends are listening, adding tracks, or settling in with headphones.",
  },
];

export function Features() {
  return (
    <section className="mx-auto w-full max-w-7xl px-gutter py-24">
      <Reveal delay={0}>
        <div className="mx-auto mb-16 max-w-3xl text-center">
          <span className="font-label-caption text-label-caption uppercase tracking-widest text-primary">
            Designed Around Sharing
          </span>
          <h2 className="mt-2 font-headline-lg text-headline-lg text-on-surface">
            Made for listening side by side
          </h2>
          <p className="mt-4 font-body-lg text-body-lg text-on-surface-variant">
            Every feature exists to shrink the distance between listeners around
            the globe.
          </p>
        </div>
      </Reveal>

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
        {FEATURES.map((feature, index) => (
          <Reveal key={feature.title} delay={index * 70}>
            <div className="rounded-2xl border border-obsidian-border bg-[#141418] p-6 transition-all hover:border-primary/40 hover:bg-[#1c1b1d] sm:p-8">
              <div className="mb-6 grid size-12 place-items-center rounded-xl border border-obsidian-border bg-surface-container text-primary">
                <span className="material-symbols-outlined text-[26px]">
                  {feature.icon}
                </span>
              </div>
              <h3 className="font-headline-sm text-headline-sm text-on-surface">
                {feature.title}
              </h3>
              <p className="mt-2 font-body-md text-body-md leading-relaxed text-on-surface-variant">
                {feature.copy}
              </p>
            </div>
          </Reveal>
        ))}
      </div>
    </section>
  );
}
