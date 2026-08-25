

interface ButtonProps {
  styling: string //Tailwind class or custom value
  text: string
  colorTheme?: string
  href?: string
  disabled?: boolean
}

function Button(props: ButtonProps) {
    const colorTheme = props.colorTheme ?? "bg-(--fg) text-(--bg)";
  return (
    <button
      disabled={props.disabled}
      className={`${colorTheme} ${props.styling} rounded-lg text-[length:var(--text-caption)] ${props.disabled ? "opacity-60 cursor-not-allowed" : "cursor-pointer"}`}
    >
      {props.text}
    </button>
    )
}

export default Button
