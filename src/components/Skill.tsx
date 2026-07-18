interface SkillInterface {
  skill: string,
  value: number
}

function Skill(props: SkillInterface) {
  return (
    <div className="bg-(--card-2) px-4.5 py-4 min-w-26 rounded-xl text-center">
      <SkillRing value={props.value}/>
      <div className="mt-2 font-medium text-sm">{props.skill}</div>
    </div>
  );
}


function SkillRing({value}: {value: number}) {
  return (
    <svg
      width="76"
      height="76"
      viewBox="0 0 76 76"
      role="img"
      aria-label="Go, 6 years, expert"
    >
      <circle
        cx="38"
        cy="38"
        r="31"
        fill="none"
        stroke="#2A2A28"
        stroke-width="5"
      ></circle>
      <circle
        cx="38"
        cy="38"
        r="31"
        fill="none"
        stroke="#FAC775"
        stroke-width="5"
        stroke-dasharray="185 195"
        stroke-linecap="round"
        transform="rotate(-90 38 38)"
      ></circle>
      <text
        x="38"
        y="45"
        text-anchor="middle"
        font-size="15"
        font-weight="600"
        font-family="Space Grotesk,sans-serif"
        fill="#F5F5F2"
      >
        {value}/10
      </text>
      {/*<text x="38" y="50" text-anchor="middle" font-size="10" fill="#6E6D68">
        expert
      </text>*/}
    </svg>
  );
}

export default Skill;
