import { projects } from "../data/project";
import Project from "./Project";

function ProjectSection() {
  return (
    <div className="pb-11">
      <h2 className="text-xl font-space font-semibold tracking-[-0.4px] mb-4.5">Projects</h2>
      <div className="grid grid-cols-2 gap-3">
        {projects.map(project =><Project
        title = {project.title}
        description = {project.description}
        stack = {project.stack}
        github = {project.github}
        liveLink = {project.liveLink} /> )}
      </div>
    </div>
  );
}

export default ProjectSection;
