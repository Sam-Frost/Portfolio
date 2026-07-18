

interface ButtonProps {
  styling: string //Tailwind class or custom value
  text: string
  colorTheme?: string
  href?: string
}

function Button(props: ButtonProps) {
    const colorTheme = props.colorTheme ?? "bg-(--button) text-(--home-bg)";
  return (
    <button className={`${colorTheme} ${props.styling} cursor-pointer rounded-lg text-[13px]`}>
      {props.text}
    </button>

    )
}

export default Button
