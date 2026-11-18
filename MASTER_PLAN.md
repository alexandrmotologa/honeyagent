# honeyagent: Architecture and Engineering Master Plan
Autonomous AI Honeypot and Threat Deception Mesh for Enterprise RAG and Agentic Systems
Target: Zero-CGO Single-Binary Go Engine (< 20MB RAM) with Embedded React 19 Deception Studio and SIEM Dispatcher

## 1. Problem Statement and Scope

### 1.1 Adversarial Threats to Enterprise AI
Enterprise deployments of retrieval augmented generation (RAG) and autonomous agents face distinct attack vectors:
1. Indirect prompt injection targeting internal document stores.
2. Knowledge exfiltration searching for confidential records, credentials, and executive communications.
3. Unadvertised tool probing attempting privilege escalation.
4. Conventional perimeter firewalls cannot evaluate semantic payloads or multi-turn conversational manipulations.

### 1.2 The Deception Security Architecture
Honeypots and canary tokens provide deterministic, high-fidelity detection:
- Legitimate users do not query unadvertised canary records.
- Legitimate workflows do not call unadvertised administrative tools.
- Any interaction with a canary token or decoy tool represents a verified attack indicator with a 100% true-positive rate.

### 1.3 Core System Responsibilities
`honeyagent` provides:
- Canary Document Synthesis and Seeding: Injects synthetic documents containing HMAC-signed cryptographic tokens into vector databases (Qdrant, pgvector, Chroma, and built-in vector store).
- Decoy Tool Manifest Injection: Publishes tempting administrative decoy tools compatible with OpenAI, Anthropic, and Model Context Protocol (MCP) schemas.
- High-Speed Stream Scanner: Evaluates inputs, retrieved chunks, and responses for canary signatures without external dependencies.
- Deceptive Hallucination Sandbox (Tarpit): Reroutes tripped sessions to a synthetic environment that serves harmless mock responses with latency delays to consume attacker time.
- SIEM and Incident Alerting: Emits Common Event Format (CEF) and Elastic Common Schema (ECS) syslog events alongside Slack and PagerDuty notifications.
- Embedded React 19 Deception Studio: Provides real-time incident monitoring, canary status, attacker trajectory playback, and simulation testing.

## 2. Technical System Architecture

```
                       Adversary / External Request
                                     |
                                     v
                  +------------------------------------+
                  |    honeyagent Airlock Gateway      |
                  |     (High-Performance Go HTTP)     |
                  +------------------------------------+
                                     |
          +--------------------------+--------------------------+
          |                                                     |
          v                                                     v
+--------------------+                                +--------------------+
| Decoy Tool Sentinel|                                | Canary Detector    |
| Intercepts probes  |                                | Validates HMAC     |
+--------------------+                                +--------------------+
          |                                                     |
          +--------------------------+--------------------------+
                                     |
                              Is Canary Tripped?
                               /            \
                             No              Yes
                             /                \
                            v                  v
                 +-------------------+  +--------------------------------+
                 | Upstream RAG/LLM  |  | Deception Countermeasure Engine|
                 | Pass-through      |  | - Session Tarpit & Latency     |
                 +-------------------+  | - Synthetic Hallucinations     |
                                        | - SIEM, Slack, PagerDuty Alert |
                                        +--------------------------------+
                                                       |
                                                       v
                                        +--------------------------------+
                                        | Embedded React 19 Studio       |
                                        | (:7878 Live Incident Feed)     |
                                        +--------------------------------+
```

## 3. Cryptographic Token Specification

Each canary token embeds an HMAC-SHA256 signature to prevent forgery and false positives:
- Format: `CANARY_v1_<hex(doc_id:timestamp:nonce)>_<hex(hmac_sha256)>`
- Payload: Document identifier, timestamp, and unique nonce.
- Signature: Computed using the master secret key `K` over the payload string.
- Verification: Constant-time comparison using `hmac.Equal`.

## 4. Repository Structure

```
honeyagent/
├── cmd/
│   └── honeyagent/
│       └── main.go                 # CLI commands: serve, seed, simulate, status
├── internal/
│   ├── config/
│   │   ├── config.go               # YAML configuration loader and environment overrides
│   │   └── schema.go               # Configuration data models
│   ├── canary/
│   │   ├── token.go                # Cryptographic HMAC-SHA256 token generator and parser
│   │   ├── synthesizer.go          # Enterprise decoy document content generator
│   │   ├── seeder.go               # Vector database seeder (Qdrant, pgvector, Chroma, memory)
│   │   └── memory_store.go         # Built-in lightweight vector store for zero-config mode
│   ├── decoy/
│   │   ├── tool_manifest.go        # Decoy tool generator (OpenAI, Anthropic, MCP)
│   │   └── interceptor.go          # Tool invocation detector and logger
│   ├── detector/
│   │   ├── stream_scanner.go       # High-throughput token scanner for inputs and outputs
│   │   └── tripwire.go             # Threat state machine and attacker session tracker
│   ├── countermeasure/
│   │   ├── tarpit.go               # Synthetic hallucination generator and latency throttle
│   │   └── sandbox.go              # Session isolation and upstream cutoff logic
│   ├── alert/
│   │   ├── dispatcher.go           # Central alert dispatcher
│   │   ├── siem.go                 # CEF and ECS syslog event emitter
│   │   ├── slack.go                # Slack webhook alert client
│   │   └── pagerduty.go            # PagerDuty Events API v2 client
│   └── server/
│       ├── server.go               # HTTP server orchestration
│       ├── proxy.go                # Reverse proxy airlock for LLM and RAG queries
│       ├── api.go                  # REST API endpoints for studio and automation
│       ├── sse.go                  # Server-Sent Events broker for real-time updates
│       └── embed.go                # Embedded React 19 UI assets
├── web/                            # React 19 single-page console
│   ├── src/
│   │   ├── App.tsx                 # Main application component
│   │   ├── types.ts                # Frontend TypeScript models
│   │   ├── components/
│   │   │   ├── Header.tsx          # Status and quick controls
│   │   │   ├── ThreatFeed.tsx      # Real-time incident stream
│   │   │   ├── CanaryTable.tsx     # Seeded canary document inventory
│   │   │   ├── DecoyTools.tsx      # Tool probing activity and schemas
│   │   │   ├── TrajectoryModal.tsx # Attacker prompt history inspector
│   │   │   └── SimulatorCard.tsx   # Built-in attack simulator runner
│   │   └── index.css               # Clean styling
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── examples/
│   ├── honeyagent.yaml             # Example configuration
│   └── mock_rag_app.go             # Vulnerable sample RAG application protected by gateway
├── Makefile
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

## 5. Implementation Roadmap

- Task 1: Initialize project, Go modules, directory structure, and configuration loader.
- Task 2: Implement HMAC-SHA256 token generator, validator, and decoy document synthesizer.
- Task 3: Build multi-target vector store seeder (Qdrant, pgvector, Chroma, internal memory store).
- Task 4: Author decoy tool catalog with OpenAI, Anthropic, and MCP schemas and call interceptor.
- Task 5: Implement zero-allocation stream scanner and concurrency-safe threat session tripwire.
- Task 6: Build countermeasure engine (tarpit latency delays and synthetic hallucination generator).
- Task 7: Implement SIEM (CEF/ECS), Slack, and PagerDuty alert dispatcher.
- Task 8: Build airlock reverse proxy and REST management API with SSE live stream.
- Task 9: Develop React 19 Deception Studio console and embed into single binary.
- Task 10: Build CLI commands, example RAG app, and end-to-end integration tests.
