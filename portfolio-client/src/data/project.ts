type Project = {
  title: string
  description: string
  stack: string[]
  github: string
  liveLink: string
}

export const projects: Project[] = [{
  title: "Brocode Crypto Exchange",
  description: "Centralized Cryptocoin Exchange supporting spots & perpetuals.",
  stack: ["Golang", "Kafka", "kubernetes", "Postgres"],
  github: "https://github.com/Sam-Frost/Brocode-Exchange-Backend",
  liveLink: "https://exchange.sat0ru.dev"
},
{
  title: "Coding Agent",
  description: "Double-entry accounting API for fintech startups. Sub-10ms reads.",
  stack: ["Golang", "Sqlite"],
  github: "https://github.com/Sam-Frost/Brocode-cli",
  liveLink: "https://codex.sat0ru.dev"
},{
  title: "Shadow-Link",
  description: "Custom VPN built in golang.",
  stack: ["Golang"],
  github: "https://github.com/Sam-Frost/Shadow-Link",
  liveLink: "https://shadowLink.sat0ru.dev"
},{
  title: "BastionX",
  description: "Custom SSH implementation in golang.",
  stack: ["Golang"],
  github: "https://github.com/Sam-Frost/BastionX",
  liveLink: "https://bastionx.sat0ru.dev"
}]
