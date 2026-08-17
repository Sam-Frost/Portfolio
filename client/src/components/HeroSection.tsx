import me from "../assets/images/me.jpeg"
import resume from "../assets/resume.pdf"
import Button from "./Button";
import { homeData } from "../data/home";

function HeroSection() {
  return (
    <div className="pt-14 pb-10">
      <div className="flex flex-col sm:flex-row gap-6 sm:gap-10 justify-between">
        <BioSection />
        <AvatarSection />
      </div>
    </div>
  );
}

function BioSection() {
  return (
    <div className="">
      <div className="inline-flex flex-row items-center gap-1.75 bg-(--green-bg) mb-4 font-medium rounded-2xl px-3 py-1 text-xs text-(--green-fg)">
        <span className="inline-block w-1.5 h-1.5 bg-(--green) rounded-full"></span>
        <div>
          {homeData.heroHighlightText}
        </div>
      </div>
      <div className="font-semibold leading-[1.1] mb-1.5 text-4xl font-space">{homeData.heroName}</div>
        <p className="mb-4 font-medium text-(--gold)">{homeData.heroSubText}</p>
      <div className="mb-5.5 max-w-105 text-(--text-muted) leading-relaxed">{homeData.heroDetails}</div>
      <div className="flex flex-row gap-3">
        <a href={resume} download="Samarth_Negi_Resume.pdf">
        <Button styling="py-2.5 px-4.5" colorTheme="text-(--fg) bg-transparent border-[0.5px] border-(--line-strong)" text="↓ Resume" />
        </a>
      </div>

    </div>
  )
}

function AvatarSection() {
  return (
    <div className="flex flex-col items-center text-center sm:items-start sm:text-left">
      <div className="h-39.5 w-39.5 border-t-amber-700">
        <img className="h-full w-full rounded-full object-cover object-center" src={me}></img>

      </div>
      <p className="text-xs mt-3 text-(--text-faint)" >{homeData.imageSubText}</p>
    </div>
  )
}

export default HeroSection;
