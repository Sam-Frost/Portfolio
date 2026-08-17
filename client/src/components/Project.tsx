interface Project {
  title: string
  description: string
  stack: string[]
  github: string
  liveLink: string
}


function Project(props: Project) {
  return (
    <div className="px-5 py-4.5  bg-(--card) rounded-xl border-(--line) border-[0.5px] border-solid">
      <div className="text-sm font-semibold font-space mb-1.5 text-(--fg)">{props.title}</div>
      <div className="text-sm text-(--text-muted) mb-3 max-w-110">{props.description}</div>
      <div className="flex flex-row justify-between items-center mt-2">
        <div className="flex flex-row gap-1.5 flex-wrap">
          {props.stack.map((skill, index) => {
            return (<ProjectSkill key={index} skill={skill} className={index >= 2 ? "hidden sm:block" : ""} />)
        })}
        </div>
        <div className="flex flex-row gap-2">
          <ProjectLinks text="Live" link={props.liveLink} />
          <ProjectLinks text="Github" link={props.github} />
        </div>
      </div>
    </div>
  );
}

function ProjectSkill({ skill, className } : {skill : string, className?: string}) {
  return (<div className={`text-[length:var(--text-pill)] bg-(--line-soft) py-0.75 px-2.5 rounded-xl text-(--text-muted) ${className ?? ""}`}>
    {skill}
  </div>)
}

function ProjectLinks({text, link} : {text: string, link: string}) {
  return (
    <a
      className="font-medium text-[length:var(--text-caption)] text-(--gold) hover:text-(--gold-deep) cursor-pointer"
      href={link}
      target="_blank"
      rel="noopener noreferrer"
    >
      {text}
    </a>
  )
}

export default Project;
