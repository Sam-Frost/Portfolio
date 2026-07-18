import type { SVGProps } from "react";

export interface IconProps extends Omit<SVGProps<SVGSVGElement>, "name"> {
  size?: number | string;
  className?: string;
  filled?: boolean;
  href?: string;
}

interface BaseIconProps extends IconProps {
  name: string;
  children: React.ReactNode;
}

export const Icon = ({
  size = 20,
  className = "",
  filled = false,
  name,
  children,
  ...props
}: BaseIconProps) => (
  <a className="hover:text-(--button)" href={props.href} target="_blank" rel="noopener noreferrer">
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill={filled ? "currentColor" : "none"}
    stroke={filled ? "none" : "currentColor"}
    strokeWidth={2}
    strokeLinecap="round"
    strokeLinejoin="round"
    className={`lucide lucide-${name} ${className}`}
    aria-hidden="true"
    {...props}
  >
    {children}
    </svg>
  </a>
);
