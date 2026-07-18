interface NavLinkProps {
  value: string
  link: string
}

function NavLink(props: NavLinkProps) {
  return (
    <div className="text-(--navbar-link) text-[13px] cursor-pointer hover:text-(--button)">
      {props.value}
    </div>
  )
}

export default NavLink;
