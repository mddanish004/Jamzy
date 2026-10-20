export function HowItWorks() {
  return (
    <section className="mx-auto w-full max-w-7xl px-gutter py-24">
      <div className="mb-16 flex flex-col justify-between md:flex-row md:items-end">
        <div>
          <span className="font-label-caption text-label-caption uppercase tracking-widest text-primary">
            The Flow
          </span>
          <h2 className="mt-2 font-headline-lg text-headline-lg text-on-surface">
            Sharing a listen is easy
          </h2>
        </div>
        <p className="mt-2 max-w-sm font-body-md text-body-md text-on-surface-variant md:mt-0">
          Spin up a shared listening room in under ten seconds, with nothing in
          your way.
        </p>
      </div>

      <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
        <Step
          number="01"
          title="Start a room"
          copy="Open a room in seconds and you instantly become the host. Find any track and hit play for everyone inside."
          accent="primary"
          icon="radio_button_checked"
          footer="Host Powers, Instantly"
        />
        <Step
          number="02"
          title="Pass the code"
          copy="Drop the code into a chat with your friends or group. They tap it and land straight in the room — no signup on your invite, and they connect without a hitch."
          accent="tertiary"
          icon="group_add"
          footer="No Listener Cap"
          card
        />
        <Step
          number="03"
          title="Enjoy it together"
          copy="Find the music, hit play, and everyone rides the exact same playback, nudged back in place by the room's drift correction."
          accent="primary"
          icon="equalizer"
          footer="Self-Correcting Playback"
        />
      </div>
    </section>
  );
}

function Step({
  number,
  title,
  copy,
  accent,
  icon,
  footer,
  card = false,
}: {
  number: string;
  title: string;
  copy: string;
  accent: "primary" | "tertiary";
  icon: string;
  footer: string;
  card?: boolean;
}) {
  return (
    <div className="group flex flex-col justify-between rounded-2xl border border-obsidian-border bg-[#141418] p-8 transition-colors hover:border-primary/40 hover:bg-[#1c1b1d]">
      <div>
        <div className="mb-6 flex size-12 items-center justify-center rounded-xl border border-obsidian-border bg-surface-container font-headline-sm text-headline-sm text-primary transition-colors group-hover:text-primary-container">
          {number}
        </div>
        <h3 className="font-headline-sm text-headline-sm text-on-surface">{title}</h3>
        <p className="mt-3 font-body-md text-body-md leading-relaxed text-on-surface-variant">
          {copy}
        </p>
        {card && (
          <div className="mt-6 flex items-center justify-between rounded-lg border border-obsidian-border bg-surface-container-lowest p-3">
            <span className="font-label-code text-label-code tracking-widest text-primary">
              ROOM - 71K9
            </span>
            <button
              type="button"
              className="flex items-center gap-1 rounded border border-obsidian-border bg-surface-container-high px-2.5 py-1 font-label-caption text-label-caption text-on-surface transition-colors hover:bg-primary-container hover:text-white"
            >
              <span className="material-symbols-outlined text-[14px]">
                content_copy
              </span>
              Copy
            </button>
          </div>
        )}
      </div>
      <div
        className={`mt-8 flex items-center gap-2 pt-4 font-label-ui text-label-ui ${
          accent === "tertiary" ? "text-tertiary" : "text-primary"
        }`}
      >
        <span className="material-symbols-outlined text-[18px]">{icon}</span>
        {footer}
      </div>
    </div>
  );
}
