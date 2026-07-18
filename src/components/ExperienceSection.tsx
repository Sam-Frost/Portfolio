import { experiences } from "../data/experience";
import Experience from "./Experience";

function ExperienceSection() {
  return (
    <div className="pb-11">
      <div className="text-xl mb-4.5 font-space font-semibold traking-[-0.4px]">Experience</div>
      {experiences.map(experience =>
        <Experience
        logo = {experience.logo}
        position= {experience.position}
        company= {experience.company}
        description= {experience.description}
        startDate= {experience.startDate}
        endDate= {experience.endDate}
        />)}

    </div>
  );
}

export default ExperienceSection;
