import {
  cloneElement,
  useEffect,
  useRef,
  useState,
  type CSSProperties,
  type ReactElement,
} from "react";

interface RevealChildProps {
  className?: string;
  style?: CSSProperties;
  ref?: (node: HTMLElement | null) => void;
}

interface RevealProps {
  children: ReactElement<RevealChildProps>;
  delay?: number;
}

export function Reveal({ children, delay = 0 }: RevealProps) {
  const nodeRef = useRef<HTMLElement | null>(null);
  const [shown, setShown] = useState(false);

  useEffect(() => {
    const node = nodeRef.current;
    if (!node) return;
    if (
      typeof IntersectionObserver === "undefined" ||
      window.matchMedia("(prefers-reduced-motion: reduce)").matches
    ) {
      setShown(true);
      return;
    }
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          setShown(true);
          observer.disconnect();
        }
      },
      { threshold: 0.1 },
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, []);

  return cloneElement(children, {
    ref: (node: HTMLElement | null) => {
      nodeRef.current = node;
    },
    className: [children.props.className, "jamzy-reveal", shown ? "jamzy-reveal-in" : ""]
      .filter(Boolean)
      .join(" "),
    style: {
      ...children.props.style,
      animationDelay: shown ? `${delay}ms` : undefined,
    },
  });
}
