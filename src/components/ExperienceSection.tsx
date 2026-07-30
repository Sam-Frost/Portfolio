import { experiences } from "../data/experience";
import Experience from "./Experience";

function ExperienceSection() {
  return (
    <div className="pb-11">
      <div className="text-xl mb-4.5 font-space font-semibold traking-[-0.4px]">Experience</div>
      {experiences.map(experience =>
        <Experience
        key={experience.company}
        logo = {experience.logo}
        position= {experience.position}
        company= {experience.company}
        description= {experience.description}
        details= {experience.details}
        techStack= {experience.techStack}
        startDate= {experience.startDate}
        endDate= {experience.endDate}
        />)}

    </div>
  );
}

export default ExperienceSection;
