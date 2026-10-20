import { Reveal } from "../ui/Reveal";

const AUDIENCES = [
  {
    icon: "diversity_1",
    tone: "primary",
    title: "Friends",
    copy: "Far-apart friends unwinding on weeknights, paging through new albums together without a camera call.",
    footer: "Share the songs",
  },
  {
    icon: "school",
    tone: "tertiary",
    title: "Students",
    copy: "Deep-work blocks, lofi study circles, and exam crunches where shared ambient sound keeps everybody flowing.",
    footer: "Focus, together",
  },
  {
    icon: "hub",
    tone: "secondary",
    title: "Communities",
    copy: "Discord servers, audio clubs, and creators throwing live release parties and group listening nights.",
    footer: "Room-wide parties",
  },
  {
    icon: "headphones",
    tone: "primary-container",
    title: "Music Lovers",
    copy: "Album devotees who savor full records end to end while commentary pours into the sidebar.",
    footer: "Shared queues",
  },
] as const;

const TONES = {
  primary: {
    chip: "border-primary/25 bg-primary/20 text-primary",
    label: "text-primary",
    hover: "hover:border-primary/40",
  },
  tertiary: {
    chip: "border-tertiary/25 bg-tertiary/20 text-tertiary",
    label: "text-tertiary",
    hover: "hover:border-tertiary/40",
  },
  secondary: {
    chip: "border-secondary/25 bg-secondary/20 text-secondary",
    label: "text-secondary",
    hover: "hover:border-secondary/40",
  },
  "primary-container": {
    chip: "border-primary-container/25 bg-primary-container/20 text-primary-container",
    label: "text-primary",
    hover: "hover:border-primary-container/40",
  },
} as const;

export function Audience() {
  return (
    <section className="mx-auto w-full max-w-7xl px-gutter py-24">
      <Reveal delay={0}>
        <div className="mx-auto mb-16 max-w-2xl text-center">
          <span className="font-label-caption text-label-caption uppercase tracking-widest text-primary">
            Built For Every Circle
          </span>
          <h2 className="mt-2 font-headline-lg text-headline-lg text-on-surface">
            Who's Jamzy made for?
          </h2>
          <p className="mt-4 font-body-lg text-body-lg text-on-surface-variant">
            From quiet study nights to friendships across oceans.
          </p>
        </div>
      </Reveal>

      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        {AUDIENCES.map((audience, index) => {
          const tone = TONES[audience.tone];
          return (
            <Reveal key={audience.title} delay={index * 90}>
              <div
                className={`flex flex-col justify-between rounded-2xl border border-obsidian-border bg-[#141418] p-6 transition-colors hover:bg-[#1c1b1d] ${tone.hover}`}
              >
                <div>
                  <div
                    className={`mb-4 grid size-10 place-items-center rounded-lg border ${tone.chip}`}
                  >
                    <span className="material-symbols-outlined text-[22px]">
                      {audience.icon}
                    </span>
                  </div>
                  <h3 className="font-headline-sm text-headline-sm text-on-surface">
                    {audience.title}
                  </h3>
                  <p className="mt-2 font-body-sm text-body-sm leading-relaxed text-on-surface-variant">
                    {audience.copy}
                  </p>
                </div>
                <div
                  className={`mt-6 flex items-center gap-1 font-label-caption text-label-caption uppercase tracking-wider ${tone.label}`}
                >
                  {audience.footer}
                  <span className="material-symbols-outlined text-[14px]">
                    arrow_forward
                  </span>
                </div>
              </div>
            </Reveal>
          );
        })}
      </div>
    </section>
  );
}
