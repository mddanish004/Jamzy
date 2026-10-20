interface LogoProps {
  className?: string;
  markOnly?: boolean;
  animated?: boolean;
  size?: number;
  dot?: boolean;
}

export function Logo({
  className = "",
  markOnly = false,
  animated = false,
  size = 40,
  dot = true,
}: LogoProps) {
  const bars = [
    { h: 12, delay: "0ms" },
    { h: 20, delay: "150ms" },
    { h: 8, delay: "300ms" },
    { h: 16, delay: "90ms" },
  ];

  return (
    <span
      className={`inline-flex items-center gap-space-sm ${className}`}
      style={{
        transition: "transform 300ms ease",
      }}
    >
      <span
        className="relative flex items-center justify-center overflow-hidden rounded-xl border border-obsidian-border bg-surface-container-high shadow-lg"
        style={{ width: size, height: size }}
        aria-hidden="true"
      >
        <span className="absolute inset-0 bg-gradient-to-tr from-primary/20 to-transparent" />
        <span className="relative flex items-end gap-[3px]" style={{ height: size * 0.4 }}>
          {bars.map((bar, i) => (
            <span
              key={i}
              className={`w-[3px] rounded-full bg-primary ${animated ? "jamzy-eq-bar" : ""}`}
              style={{ height: bar.h * (size / 40), animationDelay: bar.delay }}
            />
          ))}
        </span>
      </span>
      {!markOnly && (
        <span className="text-headline-md font-headline-md flex items-center tracking-tight text-on-surface">
          Jamzy
          {dot && (
            <span
              className="ml-1 inline-block size-1.5 rounded-full bg-primary-container"
              style={{ boxShadow: "0 0 8px #fb7185" }}
            />
          )}
        </span>
      )}
    </span>
  );
}
