

interface ButtonProps {
  styling: string //Tailwind class or custom value
  text: string
  colorTheme?: string
  href?: string
}

function Button(props: ButtonProps) {
    const colorTheme = props.colorTheme ?? "bg-(--fg) text-(--bg)";
  return (
    <button className={`${colorTheme} ${props.styling} cursor-pointer rounded-lg text-[length:var(--text-caption)]`}>
      {props.text}
    </button>
    )
}

export default Button
