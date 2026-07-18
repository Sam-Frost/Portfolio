import Button from "./Button";
import NavLink from "./NavLink";

import gojoImage from "../assets/images/gojo.jpg";

function Navbar() {
  return (
    <div className="min-w-full border-b-[0.5px] border-(--line-soft)">
    <div className="bg-(--navbar-bg) px-6 py-4 max-w-(--maxw) mx-auto ">
        <div className="flex flex-row items-center justify-between">
          <div className="flex flex-row items-center gap-2">
              <div className="h-6.5 w-6.5 overflow-hidden rounded-lg bg-(--button) text-">
              <img className="h-full w-full object-cover object-center" src={gojoImage}></img>
            </div>
            <div>sat0ru.dev</div>
          </div>
        <div className="flex flex-row flex-nowrap items-center gap-5 ">
          <NavLink value="Work" link="#" />
          <NavLink value="Projects" link="#" />
          <NavLink value="Blog" link="#" />
          <Button styling="py-1.5 px-3.5" text="Contact"></Button>
        </div>
      </div>
      </div>
    </div>
  );
}

export default Navbar;
