import { useContent } from "../content/ContentContext";
import { visibleSorted } from "../content/types";
import Project from "./Project";

function ProjectSection() {
  const projects = visibleSorted(useContent().projects);
  return (
    <div className="pb-11">
      <h2 className="text-xl font-space font-semibold tracking-[-0.4px] mb-4.5">Projects</h2>
      <div className="grid sm:grid-cols-2 gap-3">
        {projects.map((project) => (
          <Project
            key={project.id}
            title={project.title}
            description={project.description}
            stack={project.stack}
            github={project.github}
            liveLink={project.liveLink}
          />
        ))}
      </div>
    </div>
  );
}

export default ProjectSection;
