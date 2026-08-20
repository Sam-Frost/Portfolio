type Experience = {
  logo: string
  position: string
  company: string
  description: string
  details: string[]
  techStack: string[]
  startDate: string
  endDate: string
}

export const experiences: Experience[] = [
  {
    logo: "SL",
    position: "Software Engineer",
    company: "Saralya Tech Pvt. Ltd",
    description:
      "Built multi tenant LMS system from ground up, making lending easy for NBFCs",
    details: [
      "Handled DLT registration & setup, API Integration and mananing multi tentat DLT system with automated fallbacks",
    ],
    techStack: ["TypeScript", "NestJs", "PostgreSQL", "Redis", "Docker"],
    startDate: "June 2026",
    endDate: "Present",
  },
  {
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
  },
  {
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
  },
  {
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
  },
]
