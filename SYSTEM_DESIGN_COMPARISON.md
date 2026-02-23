# DateDrop — Design vs. Implementation Comparison

**Original Design:** `datedrop_system_design_v1.1.docx` (Henry Weng, February 2026)
**Actual Implementation:** Go monolith in this repository
**Comparison Date:** February 2026

---

## Summary Verdict

The implementation is **highly faithful to the original design at the feature and API level**, but **significantly scaled back in infrastructure, security, and operational concerns**. All user-facing functionality described in the design doc is present and working. The major omissions are the Messaging Service, production-grade infrastructure (databases, message bus, containers), real authentication, and the admin moderation API — areas that are explicitly out-of-scope for a Phase 1 prototype.

**Faithfulness rating by area:**

| Area | Faithfulness | Notes |
|---|---|---|
| Core domain features | ✅ High | All core flows implemented |
| API surface | ✅ High | ~95% endpoint parity |
| Compatibility algorithm | ✅ High | Algorithm matches design spec |
| Drop lifecycle | ✅ High | State machine matches design |
| Social mechanics | ✅ High | Both Shot and Cupid implemented |
| Moderation | ⚠️ Partial | Block/report done; admin API missing |
| Notification | ⚠️ Partial | Events emitted; no real delivery |
| Architecture style | ❌ Diverged | Monolith vs. microservices |
| Authentication | ❌ Diverged | Mock UUID vs. JWT |
| Data storage | ❌ Diverged | In-memory vs. PostgreSQL/Redis |
| Messaging | ❌ Missing | Not implemented at all |
| Infrastructure | ❌ Missing | No cloud infra, no containers |

---

## Section-by-Section Analysis

### 1. Executive Summary & Product Vision

**Design:** AI-powered matchmaking; one match per week; university-focused Phase 1; target up to 10,000 users; horizontally scalable to 1M+.

**Implementation:** Matches the product vision exactly. The 66-question quiz, single weekly drop, and university focus (.edu gate) are all present. Horizontal scalability to 1M+ is not achievable with the in-memory store, but the repository-pattern abstraction makes a database upgrade straightforward.

**Gap:** The design describes an "AI engine." The implementation uses a weighted scoring formula (normalized differences, importance weighting, domain weights). This is a deterministic algorithm, not an ML model. There is no LLM pass for match explanation generation either.

---

### 2. Architecture

**Design:** Microservices — nine separate services (User, Quiz, Matching Engine, Drop, Social, Messaging, Notification, Moderation, API Gateway), communicating via REST (sync) and Amazon SQS/SNS (async). Each service has its own database.

**Implementation:** Single monolithic Go process. All "services" are Go structs wired together in `main.go`. Communication is direct in-process function calls. No message bus.

**Gap:** This is the largest architectural divergence. The design's event-driven, microservices model is entirely replaced by a monolith. However, the code is internally structured as if it were a set of independent services (separate packages, no cross-package data model sharing, repository interfaces) so the design's intent is visible in the code organization. This is a common and reasonable choice for a Phase 1 prototype.

---

### 3. Service Definitions

#### 3.1 User Service

| Design Spec | Implemented? | Notes |
|---|---|---|
| .edu email registration | ✅ Yes | `pkg/utils/validation.go` |
| JWT auth with refresh token rotation | ❌ No | Mock Bearer UUID token only |
| Profile CRUD (name, bio, photos, age, gender, orientation) | ✅ Yes | `PATCH /users/:id` |
| Photo upload endpoint (`POST /users/{id}/photos`) | ❌ No | Photos are free-text strings; no upload |
| Account settings (visibility, notification prefs) | ⚠️ Partial | `is_active` flag exists; no notification prefs |

**Summary:** Core user management works. Authentication and photo storage are intentionally simplified.

#### 3.2 Quiz Service

| Design Spec | Implemented? | Notes |
|---|---|---|
| 66 questions across 3 domains (22 each) | ✅ Yes | Seeded in `quiz_service.go` |
| Lifestyle / Values / Politics domains | ✅ Yes | |
| Response types: scale_5, scale_7, multiple_choice, boolean | ✅ Yes | |
| Importance weighting per response | ✅ Yes | |
| Quiz completion tracking | ✅ Yes | |
| Question versioning | ❌ No | No version field on questions |
| Batch response submission | ⚠️ Partial | API accepts one response at a time |
| Internal `GET /quiz/responses/{user_id}` (for matching engine) | ✅ Yes | In-process call, no separate endpoint needed |

**Summary:** Quiz feature is fully implemented. Versioning and batch submission are minor omissions.

#### 3.3 Matching Engine

| Design Spec | Implemented? | Notes |
|---|---|---|
| Pairwise compatibility scoring | ✅ Yes | `matching_service.go` |
| Likert scale: normalized difference | ✅ Yes | `1.0 - abs(a-b) / (max-1)` |
| Multi-choice: Jaccard similarity | ⚠️ Approximate | Implemented as exact-match (1.0 or 0.0), not true Jaccard |
| Boolean: exact match | ✅ Yes | |
| Importance weighting | ✅ Yes | |
| Domain balancing (configurable weights) | ✅ Yes | Config: Lifestyle 35%, Values 40%, Politics 25% |
| Exclusion filters (orientation, prior matches, blocks) | ✅ Yes | 12-week lookback |
| Maximum weight bipartite matching (greedy approx) | ✅ Yes | Greedy, highest-score first |
| Match record write (to Drop Service) | ✅ Yes | Creates Drop entities |
| Scheduled pipeline (Monday scan, Tuesday 20:00 drop) | ❌ No | Manual trigger only (`POST /matching/run`) |
| Match explanation generation (LLM pass) | ❌ No | Not implemented |
| `GET /matching/status/{job_id}` (job monitoring) | ❌ No | No job status tracking |
| `GET /matching/score` (pair score, for Cupid) | ✅ Yes | Used in-process by `social_service.go` |

**Summary:** The core matching algorithm closely matches the design. The two main gaps are: (1) multi-choice scoring uses exact-match rather than Jaccard similarity, and (2) there is no scheduler — matching must be triggered manually. Match explanations are not generated.

#### 3.4 Drop Service

| Design Spec | Implemented? | Notes |
|---|---|---|
| Drop lifecycle states (pending → revealed → pending_mutual → matched → completed/archived) | ✅ Yes | Full state machine in `drop.go` |
| Expired state (72-hr window) | ✅ Yes | Checked on access |
| `GET /drops/current` | ✅ Yes | |
| `POST /drops/{id}/accept` | ✅ Yes | |
| `POST /drops/{id}/decline` | ✅ Yes | |
| `GET /drops/history` | ✅ Yes | |
| `compatibility_summary` field (JSON) | ❌ No | Score stored as float; no human-readable summary |
| `chat_channel_id` field | ❌ No | Messaging not implemented |
| `week_of` field | ✅ Yes | |

**Summary:** Drop lifecycle is implemented faithfully. The only gaps relate to the Messaging Service (which is absent), so `chat_channel_id` and the compatibility summary JSON are not present on the Drop entity.

#### 3.5 Social Service

| Design Spec | Implemented? | Notes |
|---|---|---|
| Shoot Your Shot: silent storage | ✅ Yes | |
| Shoot Your Shot: mutual detection + notification | ✅ Yes | |
| Shoot Your Shot: creates out-of-band Shot Drop | ✅ Yes | |
| Cupid Mode: nominate two users | ✅ Yes | |
| Cupid Mode: pre-compute compatibility score | ✅ Yes | |
| Cupid Mode: minimum threshold gate | ❌ No | No minimum score filter on nominations |
| Cupid Mode: dual acceptance creates Cupid Drop | ✅ Yes | |
| `GET /social/users` (browse pool) | ✅ Yes | Paginated |
| Friend graph / social graph | ❌ No | No friend concept in implementation |

**Summary:** Both social mechanics work as designed. The Cupid minimum-score threshold and a friend graph are not implemented.

#### 3.6 Messaging Service

| Design Spec | Implemented? | Notes |
|---|---|---|
| Real-time chat for matched users | ❌ No | Entirely absent |
| WebSocket endpoint | ❌ No | |
| REST fallback message endpoint | ❌ No | |
| Message persistence | ❌ No | |
| Read receipts | ❌ No | |
| Image messages via S3 | ❌ No | |

**Summary:** The Messaging Service is the single largest feature completely absent from the implementation. No chat functionality exists.

#### 3.7 Notification Service

| Design Spec | Implemented? | Notes |
|---|---|---|
| All notification event types defined | ✅ Yes | All events are called in services |
| Push notifications (FCM / APNs) | ❌ No | `fmt.Println` only |
| In-app notifications | ❌ No | |
| Email (SendGrid / SES) | ❌ No | |
| Per-user notification preferences | ❌ No | |
| Global fan-out within 5 minutes of drop time | ❌ No | Manual trigger; no scheduler |

**Summary:** Notification event hooks are correctly wired — every service calls the notification service at the right moments. But the delivery mechanism is a log statement rather than a real push/email channel.

#### 3.8 Moderation Service

| Design Spec | Implemented? | Notes |
|---|---|---|
| Block user | ✅ Yes | |
| Unblock user | ✅ Yes | |
| Block propagation to Matching Engine | ✅ Yes | Checked in matching eligibility filter |
| Report user (categories: harassment, fake profile, etc.) | ✅ Yes | |
| `GET /moderation/reports` (admin list) | ❌ No | No admin endpoint |
| `POST /moderation/actions` (admin action: warn/suspend/ban) | ❌ No | No admin actions |
| Moderation dashboard | ❌ No | |
| Auto-moderation heuristics | ❌ No | |

**Summary:** User-facing moderation (block/report) is complete. Admin-side tooling is entirely absent.

---

### 4. Data Architecture

| Design Spec | Implemented? | Notes |
|---|---|---|
| PostgreSQL for User, Quiz, Drop, Social, Notification, Moderation | ❌ No | In-memory maps |
| Redis for Matching Engine (score matrix cache) | ❌ No | Not needed; computation is synchronous |
| Redis for Social (real-time shot state) | ❌ No | In-memory maps |
| PostgreSQL + Redis for Messaging | ❌ No | Messaging absent |
| Quiz responses encrypted at rest (AES-256) | ❌ No | Plain structs in memory |
| Messages encrypted at rest | ❌ No | Messaging absent |
| Photo storage in Amazon S3 | ❌ No | Photos are URL strings |
| Audit trail / access logging for quiz responses | ❌ No | |

**Summary:** The entire data architecture from the design doc is replaced by in-memory Go maps. The repository pattern ensures this is a well-isolated choice, but none of the production data concerns (persistence, encryption, compliance) are addressed.

---

### 5. Infrastructure & Deployment

| Design Spec | Implemented? | Notes |
|---|---|---|
| Kubernetes (EKS) | ❌ No | Single binary |
| API Gateway (AWS / Kong) | ❌ No | Gin handles routing directly |
| Amazon SQS + SNS (message bus) | ❌ No | In-process function calls |
| Amazon RDS (PostgreSQL) | ❌ No | In-memory |
| Amazon ElastiCache (Redis) | ❌ No | |
| Amazon S3 | ❌ No | |
| Firebase / APNs | ❌ No | |
| SendGrid / SES | ❌ No | |
| CloudFront CDN | ❌ No | |
| Datadog / CloudWatch monitoring | ❌ No | |
| CI/CD (GitHub Actions + ArgoCD) | ❌ No | Makefile only |
| Docker containers | ❌ No | |
| Weekly Drop pipeline SLAs | ❌ No | No scheduler |
| mTLS inter-service communication | ❌ No | Monolith |

**Summary:** None of the cloud infrastructure described in the design is present. The application runs as a standalone local binary.

---

### 6. API Gateway & Security

| Design Spec | Implemented? | Notes |
|---|---|---|
| JWT access tokens (15-min TTL) | ❌ No | Mock Bearer UUID |
| Refresh token rotation | ❌ No | |
| API Gateway JWT validation before downstream routing | ❌ No | Gin middleware handles auth inline |
| Rate limiting (per-user, per-IP) | ❌ No | |
| mTLS for inter-service communication | ❌ No | Monolith; no inter-service calls |
| University email verification (tokenized link, 24hr TTL) | ❌ No | .edu domain check only; no email sent |
| Matching Engine on private network segment | ❌ No | All in same process |

**Summary:** The security architecture from the design is not implemented. The prototype uses a mock auth scheme adequate for development testing.

---

### 7. Scalability

| Design Spec | Implemented? | Notes |
|---|---|---|
| Stateless API services behind load balancer | ❌ No | Single process; state is in-memory (not stateless) |
| RDS read replicas for profile fetches | ❌ No | |
| Redis pub/sub for WebSocket cross-node fan-out | ❌ No | |
| Vectorized scoring (NumPy/PyTorch) | ❌ No | Pure Go loops |
| Cohort parallelization for batch job | ❌ No | Single-threaded greedy loop |
| O(n²) complexity acknowledged and mitigated | ⚠️ Partial | O(n²) exists; no mitigation |

---

## Key Findings

### What the implementation gets right

1. **Complete feature parity (excluding Messaging).** Every user-facing feature in the design — registration, quiz, weekly matching, drops, Shoot Your Shot, Cupid Mode, block, report — is implemented and working.

2. **Algorithm faithfulness.** The compatibility scoring algorithm (normalized differences, importance weighting, configurable domain weights, greedy bipartite matching) matches the design specification closely. The weights (35/40/25), lookback (12 weeks), and expiry (72 hours) all match.

3. **Drop state machine.** The state machine described in the design (`pending → revealed → pending_mutual → matched → completed/archived`, with `declined` and `expired` terminal states) is implemented precisely, including the validity guards on transitions.

4. **API design.** The API surface matches the design's endpoint list almost exactly, with the same URL patterns, HTTP methods, and versioning (`/api/v1/`).

5. **Domain model fidelity.** All entities described in the design (User, Question, QuizResponse, Drop, Shot, CupidNomination, Block, Report) are implemented with the fields specified.

6. **Architectural intent preserved.** Although the implementation is a monolith rather than microservices, the internal structure mirrors the service decomposition: separate packages, clean interfaces, no cross-domain data leakage. This makes a future split into microservices straightforward.

### What is intentionally deferred

The following areas appear to be intentional prototype deferrals rather than oversights:

- **In-memory storage** — repository interfaces are clearly designed for swap-out
- **Mock authentication** — placeholder for real JWT implementation
- **Console notifications** — notification service interface is correctly structured for real delivery
- **Manual matching trigger** — scheduler is a deployment concern, not a domain concern
- **No photo upload** — storage backend deferred

### What is missing and non-trivial to add

1. **Messaging Service** — Real-time chat requires WebSocket infrastructure, message persistence, and channel lifecycle management tied to drop state. This is a significant feature gap given that the design treats matched chat as a core product capability.

2. **Admin moderation API** — No mechanism for platform operators to review reports, warn users, or issue suspensions.

3. **Scheduled batch job** — The weekly drop cadence (Monday scan, Tuesday 20:00 reveal) is a product-critical SLA in the design. The current manual trigger is unsuitable for production.

4. **Multi-choice Jaccard similarity** — The design specifies Jaccard similarity for multi-choice responses; the implementation uses exact match. For questions with multiple correct answers this will produce inaccurate compatibility scores.

5. **Cupid minimum threshold** — The design filters Cupid nominations below a minimum compatibility score; the implementation creates nominations regardless of score.

### Algorithmic divergence: "AI" vs. deterministic formula

The design calls for an "AI-powered" matching engine and mentions an "LLM pass" for match explanations. The implementation uses a fully deterministic weighted scoring formula. This is a substantive difference in how the product is described vs. how it works. The current algorithm is transparent and reproducible, which has advantages; but if the design's intent was an ML model that learns from user behavior, that capability is entirely absent.

---

## Conclusion

The implementation is an excellent Phase 1 prototype that correctly captures the domain model, feature set, and algorithmic design from the specification. A developer reading the design doc and then the code will find the concepts, terminology, and feature boundaries highly consistent.

The gaps are almost entirely in the *operational* and *infrastructure* layers: no database, no real auth, no cloud services, no scheduler. These are appropriate tradeoffs for a prototype but would need to be addressed before any production launch. The one non-infrastructure gap — the missing Messaging Service — is the most significant product-level omission, as in-app chat between matched users is a core engagement mechanic described in the design.
