import gojoImage from "../assets/images/gojo.jpg";
import { useNavigate } from "react-router-dom";
import { TwitterIcon } from "../icons/TwitterIcon";
import { GithubIcon } from "../icons/GithubIcon";
import { LinkedinIcon } from "../icons/LinkedinIcon";
import { MailIcon } from "../icons/MailIcon";
import { social } from "../data/socials";
import Toast from "./Toast";
import { useState } from "react";

function Navbar() {

  const navigate = useNavigate();

  const [emailCopied, setEmailCopied] = useState(false);

  return (
    <div className="sticky top-0 z-50 bg-(--navbar-bg) min-w-full border-b-[0.5px] border-(--line-soft)">
    <div className="bg-(--navbar-bg) px-6 py-4 max-w-(--maxw) mx-auto ">
        <div className="flex flex-row items-center justify-between">
          <div className="flex flex-row items-center gap-2 cursor-pointer" onClick={() => { navigate("/")}}>
              <div className="h-6.5 w-6.5 overflow-hidden rounded-lg bg-(--fg) text-">
              <img className="h-full w-full object-cover object-center" src={gojoImage}></img>
            </div>
            <div>sat0ru.dev</div>
          </div>
          <div className="flex flex-row gap-4 text-(--text-muted)">
            <TwitterIcon href={social.x} />
            <GithubIcon href={social.github} />
            <LinkedinIcon href={social.linkedin} />
            <MailIcon className="cursor-pointer" onClick={() => {
              navigator.clipboard.writeText(social.email)
              setEmailCopied(true);
            }} />
            <Toast message="Email copied to clipboard!" show={emailCopied} onClose={() => setEmailCopied(false)} />

          </div>
      </div>
      </div>
    </div>
  );
}

export default Navbar;
