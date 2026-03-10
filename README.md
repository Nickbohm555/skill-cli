<div style="background: linear-gradient(135deg, #0b0f1a 0%, #0a1a2f 50%, #1b0f2b 100%); border: 1px solid #15f4ee; padding: 18px 20px; border-radius: 12px; color: #e6fbff; box-shadow: 0 0 18px rgba(21, 244, 238, 0.25);">
  <div style="font-size: 12px; letter-spacing: 2px; text-transform: uppercase; color: #6fffe9;">System Banner</div>
  <div style="font-size: 28px; font-weight: 700; color: #15f4ee;">Skill Weaver (cli-skill)</div>
  <div style="margin-top: 6px; font-size: 14px; color: #f7b2ff;">
    Neon-grade CLI for forging Codex skills from a single docs URL.
  </div>
  <div style="margin-top: 10px; font-size: 13px; color: #9ee8ff;">
    Status: <span style="color:#faff6b;">live blueprint</span> • Mode: <span style="color:#ff7ad9;">fail-closed</span> • Output: <span style="color:#15f4ee;">Codex-ready</span>
  </div>
</div>

# Skill Weaver (cli-skill)

Drop a link into the grid. I map the signal, interrogate the gaps, and only then mint a validated, conflict-checked skill—installed after explicit approval and nothing earlier.

**Project Description**
Skill Weaver is a Go CLI that converts a single documentation URL into a Codex skill scaffold using a gated pipeline: bounded crawl, structured processing, adaptive refinement, strict validation, conflict resolution, and approval-only installation. It is built to keep every skill atomic, auditable, and immediately usable without surprise writes.

**Why It Exists**
- One-page links are easy; usable Codex skills are not. I bridge that gap with structured refinement and strict validation.
- Skill overlap is costly. I surface conflicts and require explicit resolution before anything is written.
- Installation safety matters. I never write before preview and approval.

**Core Principles**
- Local-first pipeline with deterministic output boundaries.
- Fail-closed validation and conflict gates.
- Explicit approval before any filesystem mutation.
- Single-capability skills with explicit in-scope and out-of-scope boundaries.
- Bounded, same-domain crawl starting from a single entry URL.

<div style="margin: 14px 0; height: 2px; background: linear-gradient(90deg, #15f4ee, #ff7ad9, #faff6b); border-radius: 2px;"></div>

**Architecture At A Glance**
```mermaid
flowchart LR
    A[Docs URL] --> B[Crawl & Ingestion]
    B --> C[Content Processing & Attribution]
    C --> D[Interactive Refinement]
    D --> E[Validation & Quality Gates]
    E --> F[Overlap & Conflict Resolution]
    F --> G[Approval-Gated Install]
    G --> H[Codex Skill Registry]

    E -. fail closed .-> D
    F -. unresolved .-> D
```

**Phase Map**
Phase 1: Crawl & Ingestion Foundation
- Same-domain crawl only, default cap 50 pages, transparent skip reasons, summary counts.

Phase 2: Content Processing & Attribution
- Normalize text, preserve structure, chunk and summarize with per-chunk source attribution.

Phase 3: Interactive Refinement Loop
- Adaptive question flow, confidence-driven deepening, `revise <field>` edits, sectioned review.

Phase 4: Validation & Quality Gates
- Structural and semantic validation, one-issue-at-a-time fix loop, explicit scope boundaries.

Phase 5: Overlap & Conflict Resolution
- Detect overlap with installed skills and require explicit update/merge/abort decision.

Phase 6: Approval-Gated Install & Activation
- Preview/diff, explicit approval, atomic install, post-install verification.

**Dataflow Contract**
```mermaid
sequenceDiagram
    autonumber
    participant U as User
    participant C as cli-skill
    participant R as Registry ($CODEX_HOME/skills)

    U->>C: Provide docs URL
    C->>C: Crawl same-domain, cap pages
    C->>C: Normalize + chunk + summarize
    C->>U: Ask adaptive refinement questions
    C->>C: Validate structure and scope
    C->>C: Detect overlap and ask for decision
    C->>U: Show preview/diff and ask approval
    U->>C: Approve or cancel
    C->>R: Atomic install if approved
    C->>U: Success or blocked with guidance
```

**Tech Stack**
Core CLI dependencies (pinned):
- Go `1.25.x`
- `github.com/spf13/cobra@v1.10.2`
- `charm.land/huh/v2@v2.0.3`
- `github.com/openai/openai-go/v3@v3.26.0`
- `github.com/spf13/viper@v1.21.0`

Phase stacks (planned, per research docs):
- Crawl foundation: `github.com/gocolly/colly/v2`, `net/url`, `mime`, `github.com/PuerkitoBio/goquery`
- Content processing: `codeberg.org/readeck/go-readability/v2`, `github.com/JohannesKaufmann/html-to-markdown/v2`, `github.com/tmc/langchaingo/textsplitter`, `github.com/pkoukk/tiktoken-go`
- Refinement loop: `github.com/santhosh-tekuri/jsonschema/v6`, optional `github.com/looplab/fsm`
- Validation gates: `github.com/yuin/goldmark`, `go.abhg.dev/goldmark/frontmatter`, `github.com/go-playground/validator/v10`
- Overlap resolution: `github.com/google/go-cmp`, `github.com/sergi/go-diff`
- Install pipeline: `github.com/spf13/afero` (tests), `github.com/sergi/go-diff`, stdlib atomic rename

**Repository Layout**
Current entrypoint:
- `cmd/cli-skill/` is the main package for the CLI binary.

Current internal packages:
- `internal/crawl/` includes crawl contracts and skip-reason taxonomy.

Planned internal packages (phase-aligned):
- `internal/content/` for extraction, normalization, chunking, summarization, presentation.
- `internal/refinement/` for adaptive questioning and revision flow.
- `internal/validation/` for schema and semantic validation gates.
- `internal/overlap/` for installed-skill indexing and conflict decisions.
- `internal/install/` for preview, approval, and atomic install transaction.

**How The Build Is Shaped**
1. Start with a single docs URL.
2. Crawl same-domain pages with a strict cap and explicit skip reasons.
3. Extract and normalize content with structure preserved.
4. Chunk and summarize content with per-chunk attribution.
5. Run an adaptive question flow to fill required skill fields.
6. Validate structure and scope; fix one blocking issue at a time.
7. Detect overlap with installed skills and require explicit decision.
8. Show preview/diff, require approval, install atomically.

**Boot Sequence**
I run clean and loud when your toolchain is ready.

Prerequisites:
- Go `1.25.x` installed and on `PATH`.
- `CODEX_HOME` set to your Codex home directory.

Initialize and run:
```bash
go mod init <module-path>
go get github.com/spf13/cobra@v1.10.2
go get charm.land/huh/v2@v2.0.3
go get github.com/openai/openai-go/v3@v3.26.0
go get github.com/spf13/viper@v1.21.0

go mod tidy
go fmt ./...
go vet ./...
go test ./...

go run ./cmd/cli-skill --help
```

Build and install:
```bash
go build -o bin/cli-skill ./cmd/cli-skill
./bin/cli-skill --version

go install ./cmd/cli-skill
```

**System Status**
- Phase 1 is in progress.
- Task 1 for Phase 1 Plan 01-01 is implemented.
- Verification for that task is pending.
- Roadmap coverage is complete for all v1 requirements.

**Useful If You Want**
- A deterministic, local-first path from docs to a Codex-ready skill.
- Strict safety gates and explicit approvals before install.
- A pipeline that surfaces what it skips, why it skips, and how it decides.

**Notes**
- The binary name is `cli-skill` per project instructions.
- The project name in planning docs is Skill Weaver; both refer to the same tool.
- The pipeline is intentionally bounded in v1 to keep outputs deterministic and audit-friendly.
