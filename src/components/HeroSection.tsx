import me from "../assets/images/me.jpeg"
import resume from "../assets/resume.pdf"
import Button from "./Button";

function HeroSection() {
  return (
    <div className="pt-14 pb-10">
      <div className="flex flex-row gap-10 justify-between">
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
          Open to senior backend roles
        </div>
      </div>
      <div className="font-semibold leading-[1.1] mb-1.5 text-4xl font-space">Samarth Negi</div>
        <p className="mb-4 font-medium text-(--gold)">Backend engineer {"\u00b7"} distributed systems</p>
      <div className="mb-5.5 max-w-105 text-(--navbar-link) leading-relaxed">I build high-throughput backend systems in Go and Python. Most recently cut API p99 latency by 60% for a platform serving 2M daily users at Finlay.</div>
      <div className="flex flex-row gap-3">
        <Button styling="py-2.5 px-4.5" text="View my work" />
        <a href={resume} download="Samarth_Negi_Resume.pdf">
        <Button styling="py-2.5 px-4.5" colorTheme="text-(--button) bg-transparent border-[0.5px] border-[#3A3936]" text="↓ Resume" />
</a>
      </div>

    </div>
  )
}

function AvatarSection() {
  return (
    <div>
      <div className="h-39.5 w-39.5 border-t-amber-700">
        <img className="h-full w-full rounded-full object-cover object-center" src={me}></img>

      </div>
      <p className="text-xs mt-3 text-(--avatar-text)" >Delhi India {"\u00b7"} Remote-friendly</p>
    </div>
  )
}

export default HeroSection;
