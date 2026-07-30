import BlogSection from "../components/BlogSection";
import ExperienceSection from "../components/ExperienceSection";
import HeroSection from "../components/HeroSection";
import ProjectSection from "../components/ProjectSection";
import SkillSection from "../components/SkillSection";


/*
  Skill section commeted out, might add it later on to the page
*/
export function HomePage() {
  return (
    <>
      <HeroSection />
      <ProjectSection />
      <ExperienceSection />
      {/*<SkillSection />*/}
      <BlogSection />
    </>
  )
}
