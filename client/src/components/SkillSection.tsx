import { skills } from "../data/skill";
import Skill from "./Skill";

function SkillSection() {
  return (
    <div className=" pb-11">
      <div className="mb-4.5 text-(--fg) text-xl font-space font-semibold tracking-tight">
        Skills
      </div>
      <div className="flex flex-row flex-wrap gap-4">
        {skills.map(skill => <Skill key={skill.skill} skill={skill.skill} value={skill.value} />)}
      </div>
    </div>
  );
}

export default SkillSection;
