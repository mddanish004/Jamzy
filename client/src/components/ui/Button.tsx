import type { ComponentPropsWithoutRef, ReactNode } from "react";
import { Link } from "react-router-dom";

type Variant = "primary" | "surface" | "secondary" | "ghost";
type Size = "sm" | "md" | "lg";

interface CommonProps {
  variant?: Variant;
  size?: Size;
  className?: string;
  children: ReactNode;
}

const BASE =
  "inline-flex items-center justify-center gap-space-sm font-headline-sm text-headline-sm whitespace-nowrap " +
  "transition-all duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-primary/60 " +
  "disabled:opacity-40 disabled:pointer-events-none";

const VARIANTS: Record<Variant, string> = {
  primary:
    "bg-gradient-to-r from-[#f43f5e] to-[#fb7185] text-white border border-primary/30 shadow-lg shadow-primary-container/25 hover:opacity-95 hover:scale-[1.02] active:scale-[0.98]",
  surface:
    "bg-surface-container-low border border-obsidian-border text-on-surface hover:bg-surface-bright shadow-sm",
  secondary:
    "bg-surface-container-high hover:bg-surface-bright border border-obsidian-border/60 text-on-surface active:scale-[0.98]",
  ghost:
    "text-on-surface-variant hover:text-on-surface hover:bg-surface-container",
};

const SIZES: Record<Size, string> = {
  sm: "rounded-lg px-3 py-1.5 text-label-md font-label-md font-semibold",
  md: "rounded-xl px-space-lg py-3",
  lg: "rounded-xl px-8 py-3.5",
};

function classes({ variant = "primary", size = "md", className = "" }: CommonProps) {
  return [BASE, VARIANTS[variant], SIZES[size], className].filter(Boolean).join(" ");
}

type ButtonProps = CommonProps &
  Omit<ComponentPropsWithoutRef<"button">, "className" | "children"> & {
    to?: string;
    href?: string;
  };

export function Button({
  variant,
  size,
  className,
  children,
  to,
  href,
  ...rest
}: ButtonProps) {
  const cls = classes({ variant, size, className, children });

  if (to) {
    return (
      <Link to={to} className={cls} {...(rest as Record<string, unknown>)}>
        {children}
      </Link>
    );
  }
  if (href) {
    return (
      <a href={href} className={cls} {...(rest as Record<string, unknown>)}>
        {children}
      </a>
    );
  }
  return (
    <button className={cls} {...rest}>
      {children}
    </button>
  );
}
