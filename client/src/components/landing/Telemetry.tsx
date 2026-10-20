import { Reveal } from "../ui/Reveal";

export function Telemetry() {
  return (
    <section className="w-full border-y border-obsidian-border/40 bg-[#0e0e10] px-gutter py-20">
      <div className="mx-auto max-w-5xl rounded-3xl border border-obsidian-border bg-[#141418] p-8 shadow-2xl sm:p-12">
        <div className="flex flex-col items-center justify-between gap-10 lg:flex-row">
          <Reveal delay={0}>
            <div className="max-w-md">
              <div className="mb-4 inline-flex items-center gap-2 rounded-full border border-tertiary/20 bg-tertiary/10 px-3 py-1 font-label-caption text-label-caption text-tertiary">
                <span className="h-1.5 w-1.5 rounded-full bg-tertiary" />
                THE SERVER HOLDS THE CLOCK
              </div>
              <h2 className="font-headline-lg text-headline-lg text-on-surface">
                Nobody misses the beat.
              </h2>
              <p className="mt-4 font-body-md text-body-md leading-relaxed text-on-surface-variant">
                The room's playhead is read straight from the server's own
                timestamps, so network lag can never leave joiners behind. A
                quiet background check pulls anyone who drifts back onto the
                shared pulse.
              </p>
            </div>
          </Reveal>

          <Reveal delay={140}>
            <div className="flex w-full flex-col gap-4 rounded-2xl border border-obsidian-border bg-surface-container p-6 lg:w-96">
              <div className="flex items-center justify-between">
                <span className="font-label-ui text-label-ui text-on-surface">
                  Drift Monitor
                </span>
                <span className="flex items-center gap-1 font-label-code text-label-code text-tertiary">
                  <span className="h-2 w-2 rounded-full bg-tertiary" /> round check
                  every 5s
                </span>
              </div>

              <div className="py-2">
                <div className="mb-1 flex justify-between font-label-caption text-label-caption text-on-surface-variant">
                  <span>Host Clock</span>
                  <span className="font-bold text-primary">Aligned</span>
                  <span>Guest Node</span>
                </div>
                <div className="flex h-2 w-full overflow-hidden rounded-full border border-obsidian-border/40 bg-surface-container-lowest">
                  <div className="w-1/2 bg-gradient-to-r from-primary to-primary-container" />
                  <div className="w-1/2 bg-tertiary" />
                </div>
              </div>

              <div className="space-y-2 pt-2">
                {[
                  { label: "Roomwide Accuracy", value: "Timestamp-based", tone: "text-on-surface" },
                  { label: "Control Holder", value: "Host alone", tone: "text-on-surface" },
                  { label: "Buffer Catch-Up", value: "Automatic", tone: "text-tertiary" },
                ].map((row) => (
                  <div
                    key={row.label}
                    className="flex items-center justify-between font-body-sm text-body-sm"
                  >
                    <span className="text-on-surface-variant">{row.label}</span>
                    <span className={`font-semibold ${row.tone}`}>{row.value}</span>
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
