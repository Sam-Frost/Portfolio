import ExperienceSection from './components/ExperienceSection'
import Footer from './components/Footer'
import Navbar from './components/Navbar'
import ProjectSection from './components/ProjectSection'
import SkillSection from './components/SkillSection'
import BlogSection from './components/BlogSection'
import HeroSection from './components/HeroSection'

function App() {

  return (
    <div className="min-h-screen min-w-screen bg-(--home-bg)">
      <Navbar />
      <div className="max-w-(--maxw) mx-auto px-6">
        <HeroSection />
        <SkillSection />
        <ProjectSection />
        <ExperienceSection />
        <BlogSection />
      </div>
      <Footer />
    </div>
  )
}

export default App
