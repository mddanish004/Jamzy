export function Footer() {
  return (
    <footer className="mt-auto w-full border-t border-[#2C2B36]/50 bg-[#0e0e10] py-space-xl shadow-[0_1px_8px_rgba(0,0,0,0.04)]">
      <div className="flex w-full flex-col items-center justify-between gap-space-md px-gutter md:flex-row">
        <div className="flex items-center gap-space-sm">
          <span className="font-headline-sm text-headline-sm tracking-tight text-primary">
            Jamzy
          </span>
          <span className="font-body-sm text-body-sm text-on-surface-variant">
            — Music Heard Together, in Real Time
          </span>
        </div>
        <a
          href="https://www.mddanish.dev/"
          target="_blank"
          rel="noreferrer"
          className="font-label-ui text-label-ui text-on-surface-variant transition-colors hover:text-on-surface"
        >
          Created by Md Danish Ansari
        </a>
      </div>
    </footer>
  );
}
