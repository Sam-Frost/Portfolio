import imageBlurring from "../assets/blogs/image-bluring-at-client-vs-server.md?raw";
import type { SiteContent } from "./types";

// Compiled-in copy of the site content, used when content.json can't be
// fetched (offline, local dev with no server, a transient S3/CDN failure)
// so the site is never blank. Keep this roughly in sync with the CMS seed
// in server/internal/db/migrations/0012_add_cms.sql; the published
// content.json is the real source of truth once Publish has run.
export const fallbackContent: SiteContent = {
  version: 0,
  publishedAt: "",
  summary: {
    domain: "sat0ru.dev",
    imageSubText: "Delhi, India · Remote-friendly",
    heroHighlightText: "Open to backend roles",
    heroName: "Samarth Negi",
    heroSubText: "Backend engineer · distributed systems",
    heroDetails:
      "I build high-throughput backend systems in Go and Python. Most recently cut API p99 latency by 60% for a platform serving 2M daily users at Finlay.",
  },
  projects: [
    {
      id: "seed-project-1",
      title: "Brocode Crypto Exchange",
      slug: "brocode-crypto-exchange",
      description: "Centralized Cryptocoin Exchange supporting spots & perpetuals.",
      stack: ["Golang", "Kafka", "kubernetes", "Postgres"],
      github: "https://github.com/Sam-Frost/Brocode-Exchange-Backend",
      liveLink: "https://exchange.sat0ru.dev",
      visible: true,
      order: 0,
    },
    {
      id: "seed-project-2",
      title: "Coding Agent",
      slug: "coding-agent",
      description: "Double-entry accounting API for fintech startups. Sub-10ms reads.",
      stack: ["Golang", "Sqlite"],
      github: "https://github.com/Sam-Frost/Brocode-cli",
      liveLink: "https://codex.sat0ru.dev",
      visible: true,
      order: 1,
    },
    {
      id: "seed-project-3",
      title: "Shadow-Link",
      slug: "shadow-link",
      description: "Custom VPN built in golang.",
      stack: ["Golang"],
      github: "https://github.com/Sam-Frost/Shadow-Link",
      liveLink: "https://shadowLink.sat0ru.dev",
      visible: true,
      order: 2,
    },
    {
      id: "seed-project-4",
      title: "BastionX",
      slug: "bastionx",
      description: "Custom SSH implementation in golang.",
      stack: ["Golang"],
      github: "https://github.com/Sam-Frost/BastionX",
      liveLink: "https://bastionx.sat0ru.dev",
      visible: true,
      order: 3,
    },
  ],
  experiences: [
    {
      id: "seed-exp-1",
      logo: "SL",
      position: "Software Engineer",
      company: "Saralya Tech Pvt. Ltd",
      description: "Built multi tenant LMS system from ground up, making lending easy for NBFCs",
      details: [
        "Handled DLT registration & setup, API Integration and mananing multi tentat DLT system with automated fallbacks",
      ],
      techStack: ["TypeScript", "NestJs", "PostgreSQL", "Redis", "Docker"],
      startDate: "June 2026",
      endDate: "Present",
      visible: true,
      order: 0,
    },
    {
      id: "seed-exp-2",
      logo: "BT",
      position: "Associate Software Engineer",
      company: "BitxiaTech Pvt. Ltd.",
      description:
        "Built scalable Spring Boot backend services for agriculture/eNAM systems with Kafka-based asynchronous communication, high availability patterns, and optimized database performance.",
      details: [
        "Developed and scaled 3 core services using Kafka-based asynchronous inter-service communication.",
        "Implemented ONDC protocol for agriculture services, enabling interoperability between buyers, sellers, and logistics providers.",
        "Achieved 99.9% availability by implementing Retry and Circuit Breaker patterns to prevent cascading failures.",
        "Cut API latency by 30% using caching and SQL query plan analysis.",
        "Designed scalable database schemas supporting 1M+ rows while keeping query execution under 50ms.",
        "Reduced deployment time by 60% using Docker, Kubernetes, and CI/CD layer caching.",
      ],
      techStack: ["Java", "Spring Boot", "Kafka", "PostgreSQL", "Redis", "Docker", "Kubernetes"],
      startDate: "Feb 2025",
      endDate: "Oct 2025",
      visible: true,
      order: 1,
    },
    {
      id: "seed-exp-3",
      logo: "TH",
      position: "Backend Developer",
      company: "Techostinger India Pvt. Ltd.",
      description:
        "Owned backend development for production APIs, real-time chat, media processing pipelines, payments, and automated deployment workflows.",
      details: [
        "Designed and developed RESTful APIs and GitLab CI pipelines for automated deployment.",
        "Engineered a real-time chat system using WebSocket and Redis supporting 1,000+ concurrent connections.",
        "Reduced database writes by 70% for real-time last-seen user activity tracking.",
        "Built an asynchronous media pipeline using AWS S3 and SQS, reducing main-thread load by 80%.",
        "Wrote 150+ unit and integration tests with Vitest and Supertest, reaching 80% code coverage.",
        "Integrated Stripe for secure transactions and subscription payments.",
      ],
      techStack: ["Node.js", "Express", "TypeScript", "Redis", "WebSocket", "AWS S3", "AWS SQS", "Stripe"],
      startDate: "June 2024",
      endDate: "Dec 2024",
      visible: true,
      order: 2,
    },
    {
      id: "seed-exp-4",
      logo: "JM",
      position: "Mobile Development Intern",
      company: "JMVL",
      description:
        "Built a cross-platform Flutter application with API integrations, authentication flows, and stable state management for Android and iOS.",
      details: [
        "Developed a cross-platform mobile application using Flutter for Android and iOS.",
        "Integrated 15+ REST APIs for real-time data handling, authentication, and core app functionality.",
        "Implemented Bloc and Provider state management, reducing memory leaks and improving app stability by 20%.",
      ],
      techStack: ["Flutter", "Dart", "REST APIs", "Bloc", "Provider"],
      startDate: "June 2023",
      endDate: "Dec 2023",
      visible: true,
      order: 3,
    },
  ],
  blogs: [
    {
      id: "seed-blog-1",
      title: "Image bluring at client vs server",
      slug: "image-bluring-at-client-vs-server",
      readTime: "5 min read",
      genre: "Backend",
      date: "Dec 2024",
      body: imageBlurring,
      visible: true,
      order: 0,
    },
  ],
};
