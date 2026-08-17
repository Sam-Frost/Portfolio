import { useState } from "react";

interface Experience {
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
      <div className="grow shrink">
        <div
          className="flex flex-col sm:flex-row justify-between items-baseline gap-3 cursor-pointer"
          onClick={() => setIsOpen(!isOpen)}
        >
          <div className="text-sm font-medium flex flex-row items-center gap-1.5">

            {props.position} · {props.company}
          </div>
          <div className= "flex flex-row justify-center items-center gap-2">
          <div className="text-xs text-(--text-faint) whitespace-nowrap">
            {props.startDate} - {props.endDate}
          </div>
          <svg
            className={`transition-transform text-(--text-faint) mb-0.5 duration-300 ${isOpen ? "rotate-90" : "-rotate-90"}`}
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
        <div className="text-[length:var(--text-caption)] mt-1 text-(--text-muted) cursor-pointer" onClick={() => setIsOpen(!isOpen)}>
          {props.description}
        </div>
        <div
          className={`grid transition-all duration-300 ease-in-out ${isOpen ? "grid-rows-[1fr] opacity-100" : "grid-rows-[0fr] opacity-0"}`}
        >
          <div className="overflow-hidden">
            <div className="mt-3">
              <ul className="list-disc pl-4 flex flex-col gap-1.5">
                {props.details.map(detail => (
                  <li key={detail} className="text-[length:var(--text-caption)] text-(--text-muted)">
                    {detail}
                  </li>
                ))}
              </ul>
              <div className="flex flex-row gap-1.5 flex-wrap mt-3">
                {props.techStack.map(skill => (
                  <div
                    key={skill}
                    className="text-[length:var(--text-pill)] bg-(--line-soft) py-0.75 px-2.5 rounded-xl text-(--text-muted)"
                  >
                    {skill}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>

        <hr className="mt-4 text-(--text-muted) opacity-20"/>
      </div>
    </div>
  );
}

export default Experience;
