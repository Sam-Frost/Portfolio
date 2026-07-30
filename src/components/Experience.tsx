import { useState } from "react";

interface Experience {
  logo: string
  position: string
  company: string
  description: string
  details: string[]
  techStack: string[]
  startDate: string
  endDate: string
}

function Experience(props : Experience) {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className="flex flex-row gap-4 pb-5.5">
      <div className="flex flex-col items-center">
        <div className="w-9 h-9 rounded-[9px] text-(--gold) bg-(--gold-bg) flex flex-col items-center justify-center">
          {props.logo}
        </div>
      </div>
      <div className="grow shrink">
        <div
          className="flex justify-between items-baseline gap-3 cursor-pointer"
          onClick={() => setIsOpen(!isOpen)}
        >
          <div className="text-sm font-medium flex flex-row items-center gap-1.5">

            {props.position} · {props.company}
          </div>
          <div className= "flex flex-row justify-center items-center gap-2">
          <div className="text-xs text-(--avatar-text) whitespace-nowrap">
            {props.startDate} - {props.endDate}
          </div>
          <svg
            className={`transition-transform text-(--avatar-text) mb-0.5 duration-300 ${isOpen ? "rotate-90" : "-rotate-90"}`}
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth={2}
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M9 6l6 6-6 6" />
            </svg>
          </div>
        </div>
        <div className="text-[13px] mt-1 text-(--navbar-link) cursor-pointer" onClick={() => setIsOpen(!isOpen)}>
          {props.description}
        </div>
        <div
          className={`grid transition-all duration-300 ease-in-out ${isOpen ? "grid-rows-[1fr] opacity-100" : "grid-rows-[0fr] opacity-0"}`}
        >
          <div className="overflow-hidden">
            <div className="mt-3">
              <ul className="list-disc pl-4 flex flex-col gap-1.5">
                {props.details.map(detail => (
                  <li key={detail} className="text-[13px] text-(--navbar-link)">
                    {detail}
                  </li>
                ))}
              </ul>
              <div className="flex flex-row gap-1.5 flex-wrap mt-3">
                {props.techStack.map(skill => (
                  <div
                    key={skill}
                    className="text-[11px] bg-(--line-soft) py-0.75 px-2.5 rounded-xl text-(--navbar-link)"
                  >
                    {skill}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default Experience;
