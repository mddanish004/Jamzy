import { useState } from "react";

interface ShowcaseArtProps {
  src: string;
  alt: string;
  className?: string;
}

export function ShowcaseArt({ src, alt, className = "" }: ShowcaseArtProps) {
  const [failed, setFailed] = useState(false);

  if (failed) {
    return (
      <div
        className={`${className} grid place-items-center bg-gradient-to-br from-[#73323d] via-[#f43f5e]/40 to-[#0e0e10]`}
        role="img"
        aria-label={alt}
      >
        <span className="material-symbols-outlined text-[40%] text-primary/70 msr-fill">
          music_cast
        </span>
      </div>
    );
  }

  return (
    <img
      src={src}
      alt={alt}
      className={className}
      onError={() => setFailed(true)}
      loading="lazy"
    />
  );
}
