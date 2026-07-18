interface Experience {
  logo: string
  position: string
  company: string
  description: string
  startDate: string
  endDate: string
}

function Experience(props : Experience) {
  return (
    <div className="flex flex-row gap-4 pb-5.5">
      <div className="flex flex-col items-center">
        <div className="w-9 h-9 rounded-[9px] text-(--gold) bg-(--gold-bg) flex flex-col items-center justify-center">
          {props.logo}
        </div>
      </div>
      <div className="grow shrink">
        <div className="flex justify-between items-baseline gap-3">
          <div className="text-sm font-medium">
            {props.position} · {props.company}
          </div>
          <div className="text-xs text-(--avatar-text) whitespace-nowrap">
            {props.startDate} - {props.endDate}
          </div>
        </div>
        <div className="text-[13px] mt-1 text-(--navbar-link) ">
          {props.description}
        </div>
      </div>
    </div>
  );
}

export default Experience;
