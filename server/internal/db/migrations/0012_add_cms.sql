-- CMS: editable content for the public portfolio site. Every row here is
-- DRAFT state; the site only changes when a Publish serializes this to a
-- content.json (see internal/cms). Seeded with the content that currently
-- lives in portfolio-client/src/data/*.ts so the CMS opens pre-populated.

CREATE TABLE IF NOT EXISTS cms_projects (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    stack       JSONB NOT NULL DEFAULT '[]',
    github      TEXT NOT NULL DEFAULT '',
    live_link   TEXT NOT NULL DEFAULT '',
    visible     BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order  INT NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cms_experiences (
    id          TEXT PRIMARY KEY,
    logo        TEXT NOT NULL DEFAULT '',
    position    TEXT NOT NULL,
    company     TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    details     JSONB NOT NULL DEFAULT '[]',
    tech_stack  JSONB NOT NULL DEFAULT '[]',
    start_date  TEXT NOT NULL DEFAULT '',
    end_date    TEXT NOT NULL DEFAULT '',
    visible     BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order  INT NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cms_blogs (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    read_time   TEXT NOT NULL DEFAULT '',
    genre       TEXT NOT NULL DEFAULT '',
    date        TEXT NOT NULL DEFAULT '',
    body        TEXT NOT NULL DEFAULT '',
    visible     BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order  INT NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cms_summary (
    id                  TEXT PRIMARY KEY,
    domain              TEXT NOT NULL DEFAULT '',
    image_sub_text      TEXT NOT NULL DEFAULT '',
    hero_highlight_text TEXT NOT NULL DEFAULT '',
    hero_name           TEXT NOT NULL DEFAULT '',
    hero_sub_text       TEXT NOT NULL DEFAULT '',
    hero_details        TEXT NOT NULL DEFAULT '',
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cms_publications (
    id           TEXT PRIMARY KEY,
    version      INT NOT NULL,
    published_at TIMESTAMPTZ NOT NULL,
    status       TEXT NOT NULL,
    error        TEXT,
    snapshot     JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cms_publications_version ON cms_publications (version DESC);

-- ─────────────────────────────────────────────
-- Seed: current live content (portfolio-client/src/data/*.ts)
-- ─────────────────────────────────────────────

INSERT INTO cms_summary (id, domain, image_sub_text, hero_highlight_text, hero_name, hero_sub_text, hero_details)
VALUES (
    'singleton',
    'sat0ru.dev',
    'Delhi, India · Remote-friendly',
    'Open to backend roles',
    'Samarth Negi',
    'Backend engineer · distributed systems',
    'I build high-throughput backend systems in Go and Python. Most recently cut API p99 latency by 60% for a platform serving 2M daily users at Finlay.'
) ON CONFLICT (id) DO NOTHING;

INSERT INTO cms_projects (id, title, slug, description, stack, github, live_link, sort_order) VALUES
(
    'seed-project-1', 'Brocode Crypto Exchange', 'brocode-crypto-exchange',
    'Centralized Cryptocoin Exchange supporting spots & perpetuals.',
    '["Golang","Kafka","kubernetes","Postgres"]',
    'https://github.com/Sam-Frost/Brocode-Exchange-Backend', 'https://exchange.sat0ru.dev', 0
),
(
    'seed-project-2', 'Coding Agent', 'coding-agent',
    'Double-entry accounting API for fintech startups. Sub-10ms reads.',
    '["Golang","Sqlite"]',
    'https://github.com/Sam-Frost/Brocode-cli', 'https://codex.sat0ru.dev', 1
),
(
    'seed-project-3', 'Shadow-Link', 'shadow-link',
    'Custom VPN built in golang.',
    '["Golang"]',
    'https://github.com/Sam-Frost/Shadow-Link', 'https://shadowLink.sat0ru.dev', 2
),
(
    'seed-project-4', 'BastionX', 'bastionx',
    'Custom SSH implementation in golang.',
    '["Golang"]',
    'https://github.com/Sam-Frost/BastionX', 'https://bastionx.sat0ru.dev', 3
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO cms_experiences (id, logo, position, company, description, details, tech_stack, start_date, end_date, sort_order) VALUES
(
    'seed-exp-1', 'SL', 'Software Engineer', 'Saralya Tech Pvt. Ltd',
    'Built multi tenant LMS system from ground up, making lending easy for NBFCs',
    '["Handled DLT registration & setup, API Integration and mananing multi tentat DLT system with automated fallbacks"]',
    '["TypeScript","NestJs","PostgreSQL","Redis","Docker"]',
    'June 2026', 'Present', 0
),
(
    'seed-exp-2', 'BT', 'Associate Software Engineer', 'BitxiaTech Pvt. Ltd.',
    'Built scalable Spring Boot backend services for agriculture/eNAM systems with Kafka-based asynchronous communication, high availability patterns, and optimized database performance.',
    '["Developed and scaled 3 core services using Kafka-based asynchronous inter-service communication.","Implemented ONDC protocol for agriculture services, enabling interoperability between buyers, sellers, and logistics providers.","Achieved 99.9% availability by implementing Retry and Circuit Breaker patterns to prevent cascading failures.","Cut API latency by 30% using caching and SQL query plan analysis.","Designed scalable database schemas supporting 1M+ rows while keeping query execution under 50ms.","Reduced deployment time by 60% using Docker, Kubernetes, and CI/CD layer caching."]',
    '["Java","Spring Boot","Kafka","PostgreSQL","Redis","Docker","Kubernetes"]',
    'Feb 2025', 'Oct 2025', 1
),
(
    'seed-exp-3', 'TH', 'Backend Developer', 'Techostinger India Pvt. Ltd.',
    'Owned backend development for production APIs, real-time chat, media processing pipelines, payments, and automated deployment workflows.',
    '["Designed and developed RESTful APIs and GitLab CI pipelines for automated deployment.","Engineered a real-time chat system using WebSocket and Redis supporting 1,000+ concurrent connections.","Reduced database writes by 70% for real-time last-seen user activity tracking.","Built an asynchronous media pipeline using AWS S3 and SQS, reducing main-thread load by 80%.","Wrote 150+ unit and integration tests with Vitest and Supertest, reaching 80% code coverage.","Integrated Stripe for secure transactions and subscription payments."]',
    '["Node.js","Express","TypeScript","Redis","WebSocket","AWS S3","AWS SQS","Stripe"]',
    'June 2024', 'Dec 2024', 2
),
(
    'seed-exp-4', 'JM', 'Mobile Development Intern', 'JMVL',
    'Built a cross-platform Flutter application with API integrations, authentication flows, and stable state management for Android and iOS.',
    '["Developed a cross-platform mobile application using Flutter for Android and iOS.","Integrated 15+ REST APIs for real-time data handling, authentication, and core app functionality.","Implemented Bloc and Provider state management, reducing memory leaks and improving app stability by 20%."]',
    '["Flutter","Dart","REST APIs","Bloc","Provider"]',
    'June 2023', 'Dec 2023', 3
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO cms_blogs (id, title, slug, read_time, genre, date, body, sort_order) VALUES
(
    'seed-blog-1',
    'Image bluring at client vs server',
    'image-bluring-at-client-vs-server',
    '5 min read',
    'Backend',
    'Dec 2024',
    $md$# Image Blurring at Client vs Server

I was building a matrimonial platform when I came across this issue of image blurring. We needed to make sure the images of the users are blurred before they hit the client in order to maintain privacy of the user.

I encountered this again while browsing LinkedIn. I noticed that the "Viewers you might be interested in" section had images blurred from the backend, likely because it's a paid feature. If users could access the data just by inspecting elements, it wouldn't be much of a paywall. In contrast, the "People you may know" section applied the blur effect on the client side, making it easy to reveal image URLs by tinkering with the inspect element.

This experience led me to explore the differences between client-side and server-side blurring.

## Client Side Blurring

Client-side blurring is the process of blurring the image on the client's machine using CSS or JavaScript running on the browser.

The key advantage of client-side blurring is offloading the processing from the server. While blurring may seem lightweight initially, at scale, handling thousands of simultaneous requests can create significant computational overhead.

## Server Side Blurring

Server-side blurring involves processing the image on the server before sending it to the client.

Another approach is to generate and store two versions of the image at the time of upload, allowing the system to serve the appropriate version as needed. This approach is more efficient than the previous one since the image is blurred only once and stored, allowing us to serve the pre-processed version directly to the client. This reduces computational overhead and minimizes latency whenever the client requests the image.

## Why Prefer Server Side Blurring

- Client-side blurring is limited to CSS and JavaScript, while server-side processing allows the use of custom libraries and different programming languages.
- Server-side blurring offers more control over image modifications, enabling advanced processing beyond just blurring.

## Conclusion

Image blurring can be a computationally expensive task. It is preferred to offload this computationally expensive task to the client whenever possible.

Of course, not everything can be offloaded to the client. If there is sensitive information (for example, people who have viewed your profile on LinkedIn) which shall not be accessible to the user or is hidden behind a paywall, that blurring must happen server-side.

So the general approach would be to offload blurring tasks as much as we can to the client, but whenever we have to deal with sensitive information, the server has to bear the extra computational cost of blurring.
$md$,
    0
)
ON CONFLICT (id) DO NOTHING;
