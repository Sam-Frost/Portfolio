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
      <div className="text-sm font-semibold font-space mb-1.5 text-(--button)">{props.title}</div>
      <div className="text-sm text-(--navbar-link) mb-3 max-w-110">{props.description}</div>
      <div className="flex flex-row justify-between items-center mt-2">
        <div className="flex flex-row gap-1.5 flex-wrap">
          {props.stack.map(skill => {
            return (<ProjectSkill skill={skill} />)
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

function ProjectSkill({ skill } : {skill : string}) {
  return (<div className="text-[11px] bg-(--line-soft) py-0.75 px-2.5 rounded-xl text-(--navbar-link)">
    {skill}
  </div>)
}

function ProjectLinks({text, link} : {text: string, link: string}) {
  return (
    <div className="font-medium text-[13px] text-(--gold) hover:text-(--gold-deep) cursor-pointer " onClick={() => {
      console.log(`Redirect to ${link}`)
    }}>{text}</div>
  )
}

export default Project;
