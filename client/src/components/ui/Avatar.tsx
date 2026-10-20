interface AvatarProps {
  name: string;
  size?: number;
  online?: boolean;
  className?: string;
}

const GRADIENTS = [
  "from-[#f43f5e] to-[#be123c]",
  "from-[#fb7185] to-[#f43f5e]",
  "from-[#10b981] to-[#047857]",
  "from-[#fda4af] to-[#fb7185]",
  "from-[#00b47d] to-[#005236]",
  "from-[#fb923c] to-[#ea580c]",
  "from-[#a78bfa] to-[#7c3aed]",
  "from-[#38bdf8] to-[#0369a1]",
];

function gradientFor(name: string): number {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = (hash * 31 + name.charCodeAt(i)) >>> 0;
  }
  return hash % GRADIENTS.length;
}

export function Avatar({ name, size = 40, online, className = "" }: AvatarProps) {
  const trimmed = name.trim();
  const initial = trimmed.charAt(0).toUpperCase() || "?";
  return (
    <div
      className={`relative shrink-0 rounded-full ${className}`}
      style={{ width: size, height: size }}
    >
      <div
        className={`grid h-full w-full place-items-center rounded-full bg-gradient-to-br ${
          GRADIENTS[gradientFor(trimmed)]
        } font-semibold text-white`}
        style={{ fontSize: Math.max(10, size * 0.38) }}
      >
        {initial}
      </div>
      {online !== undefined && (
        <span
          className={`absolute bottom-0 right-0 size-2.5 rounded-full border border-[#1c1b1d] ${
            online ? "bg-tertiary shadow-[0_0_8px_rgba(16,185,129,0.8)]" : "bg-surface-bright"
          }`}
          aria-hidden="true"
        />
      )}
    </div>
  );
}
