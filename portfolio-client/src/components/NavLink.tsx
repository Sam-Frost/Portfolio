interface NavLinkProps {
  value: string
  link: string
}

function NavLink(props: NavLinkProps) {
  return (
    <div className="text-(--text-muted) text-[length:var(--text-caption)] cursor-pointer hover:text-(--fg)">
      {props.value}
    </div>
  )
}

export default NavLink;
