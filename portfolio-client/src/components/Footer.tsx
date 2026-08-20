import { useState } from "react";
import Button from "./Button";

function Footer() {

  const [footNote, setFootNote] = useState(false);

  return (
    <div className="px-6 py-9 text-center border-t-[0.5px] border-(--line-soft) bg-(--card)">
      <h2 className="mb-1.5 text-xl sapce-font font-semibold spacing tracking-[-0.4px]" >Let's build something together</h2>
      <p className="text-[length:var(--text-caption)] mb-4.5 text-(--text-muted)">samarthnegi2002@gmail.com · usually replies within a day</p>
      <div className="">
        <div onMouseEnter={() => setFootNote(true)}
             onMouseLeave={() => setFootNote(false)}
        >
          <Button styling="px-4.5 py-2" colorTheme="bg-transparent  border-[0.5px] border-(--line-strong)" text="📅 Book a call" />
        </div>
        {footNote &&
          <div className="mt-2 text-[length:var(--text-caption)] mb-4.5 text-(--text-muted)">Inegration with calendly under development 😅😅  Please email me</div>
        }
      </div>
    </div>
  );
}

export default Footer;
