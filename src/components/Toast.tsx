import { useEffect } from "react";

interface ToastProps {
  message: string;
  show: boolean;
  duration?: number;
  onClose?: () => void;
}

export default function Toast({ message, show, duration = 800, onClose }: ToastProps) {
  useEffect(() => {
    if (!show) return;
    const t = setTimeout(() => onClose?.(), duration);
    return () => clearTimeout(t);
  }, [show, duration, onClose]);

  return (
    <div
      className="fixed top-4 left-1/2 z-50 px-4 py-2 text-sm rounded-lg bg-(--card) text-(--navbar-link) shadow-lg transition-all duration-300 pointer-events-none"
      style={{
        transform: `translateX(-50%) translateY(${show ? "0" : "-4rem"})`,
        opacity: show ? 1 : 0,
      }}
    >
      {message}
    </div>
  );
}
