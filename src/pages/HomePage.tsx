import BlogSection from "../components/BlogSection";
import ExperienceSection from "../components/ExperienceSection";
import HeroSection from "../components/HeroSection";
import ProjectSection from "../components/ProjectSection";
import SkillSection from "../components/SkillSection";

export function HomePage() {
  return (
    <>
      <HeroSection />
      <SkillSection />
      <ProjectSection />
      <ExperienceSection />
      <BlogSection />
    </>
  )
}
