# repo-context-compiler

A local AI-oriented codebase context compiler for large repositories.

`repo-context-compiler` scans a repository, builds a semantic understanding of the codebase, compresses it into structured representations, and generates AI-friendly context packs for LLM-assisted coding.

The goal is not to dump raw code into an LLM.

The goal is to maximize:

- signal
- architecture awareness
- dependency understanding
- semantic meaning

while minimizing token usage.

---

# Why This Exists

Even modern LLMs with massive context windows still struggle with large real-world repositories.

A common misconception is:

> "Once models support 1M tokens, context engineering disappears."

Unfortunately, reality is uglier.

Problems include:

- attention dilution
- lost-in-the-middle failures
- retrieval confusion
- contradictory context interference
- degraded reasoning consistency

The challenge is no longer:

> "Can the model hold the repository?"

The challenge is:

> "Can the model focus on the right things?"

This project attempts to solve that problem by transforming source code into compressed semantic representations.

---

# Core Idea

Instead of sending raw code:

```python
def process_payment(user_id, amount, currency):
    ...
```

Generate structured semantic metadata:

```json
{
  "function": "process_payment",
  "inputs": ["user_id", "amount", "currency"],
  "side_effects": ["charges user", "writes transaction"],
  "dependencies": ["payment_gateway", "db"],
  "risk": "high"
}
```

The LLM can then reason about:

* architecture
* dependencies
* ownership
* side effects
* risks
* module boundaries

without needing the entire codebase in context.

---

# Goals

* Scan large repositories locally
* Generate repository tree structures
* Parse source files into semantic representations
* Generate AI-friendly compressed context packs
* Support local and remote OpenAI-compatible LLM providers
* Support incremental semantic summarization
* Build dependency graphs and import relationships
* Recommend relevant files for tasks and bug fixes
* Work fully offline without requiring any LLM
* Reduce token usage for coding assistants
* Help LLMs request only the files they actually need

---

# Why Go?

This project is intentionally designed in Go.

## Why not Node.js?

Node.js is fast for prototyping, but this project is heavily:

* filesystem-oriented
* parsing-oriented
* concurrency-oriented
* CLI-oriented

Go performs better for:

* repo walking
* local tooling
* single-binary distribution
* concurrency
* indexing pipelines

## Why not Rust?

Rust is excellent for performance, but:

* development iteration is slower
* parser integrations are more complex
* ergonomics are heavier for rapid experimentation

Go provides the best balance of:

* speed
* simplicity
* portability
* maintainability

---

# High-Level Architecture

```text
repo-context-compiler
│
├── scan
│   └── walks repository and builds file inventory
│
├── parse
│   └── extracts functions, classes, imports, exports
│
├── summarize
│   └── uses local or remote LLMs to generate semantic metadata
│
├── dependency graph
│   └── builds architecture relationships between files
│
├── index
│   └── stores cached summaries and hashes
│
├── pack
│   └── generates AI-friendly outputs
│
└── query
    └── suggests relevant files for tasks
```

---

# Planned Technology Stack

| Component     | Technology                   |
| ------------- | ---------------------------- |
| Language      | Go                           |
| Parser        | tree-sitter                  |
| LLM Providers | Ollama / OpenAI / OpenRouter |
| Cache DB      | SQLite                       |
| Embeddings    | Ollama embeddings            |
| Output        | JSON + Markdown              |

---

# Why tree-sitter?

`tree-sitter` provides concrete syntax trees for source code.

This enables:

* function extraction
* class extraction
* import analysis
* dependency mapping
* symbol indexing

without relying on fragile regex parsing.

---

# Why OpenAI-Compatible Providers?

The project uses a provider-agnostic OpenAI-compatible HTTP client.

This allows support for:

* Ollama
* OpenAI
* OpenRouter
* Together.ai
* Groq
* local vLLM servers
* LiteLLM gateways

without requiring provider-specific implementations.

Most modern providers support:

```text
POST /v1/chat/completions
```

This dramatically simplifies the architecture.

The primary differences between providers become:

* base URL
* API key
* model name

---

# Why Ollama?

Ollama is still the preferred local-first option because it enables:

* fully offline workflows
* local summarization
* local embeddings
* privacy-first development
* no cloud dependency
* no repository upload

Potential models:

* qwen2.5-coder
* deepseek-coder
* llama3
* codellama
* nomic embeddings

---

# AI Providers Are Optional

`repo-context-compiler` is designed as:

```text
core semantic repository compiler
+ optional AI augmentation
```

The core system works completely offline without:

* OpenAI
* OpenRouter
* Ollama
* API keys
* internet access

This is an intentional architectural decision.

The repository compiler itself is responsible for:

* repository scanning
* tree generation
* tree-sitter parsing
* symbol extraction
* dependency indexing
* semantic structure generation
* JSON context packs
* Markdown context packs

LLMs are only used for optional enhancements such as:

* semantic summaries
* architecture descriptions
* side effect analysis
* risk classification
* embeddings
* semantic search

---

# Offline-First Design

The following command works entirely offline:

```bash
./repoctx pack .
```

This generates:

```text
.repoctx/
├── repoctx.db
├── ai-context.md
└── ai-context.json
```

without requiring any LLM provider configuration.

This makes the tool suitable for:

* local development
* secure environments
* private repositories
* air-gapped systems
* CI pipelines
* deterministic indexing

---

# AI-Enhanced Mode

LLMs are enabled only when explicitly requested:

```bash
./repoctx pack . --summarize
```

This prevents AI providers from becoming a hard dependency.

---

# Core Workflow

```text
1. Scan repository
2. Parse symbols and imports
3. Build dependency graph
4. Generate semantic summaries (optional)
5. Cache summaries incrementally
6. Store semantic representations locally
7. Generate compressed AI context packs
8. Ask the system what files are relevant
9. Provide only necessary source files to frontier LLMs
```

---

# How to Use This Tool

## 1. Install dependencies

```bash
go mod tidy
```

## 2. Build the CLI

```bash
go build -o repoctx ./cmd/repoctx
```

This creates a local binary:

```text
./repoctx
```

## 3. Initialize the local database

```bash
./repoctx init
```

This creates the local `.repoctx/` folder and SQLite database.

## 4. Generate an AI context pack

```bash
./repoctx pack .
```

This scans the current repository and generates AI-friendly context files.

You should now see:

```text
.repoctx/
├── repoctx.db
├── ai-context.md
└── ai-context.json
```

---
# Build, Test, and Verify

## Run Tests

Execute all unit tests:

```bash
go test ./...
```

Run static analysis:

```bash
go vet ./...
```

## Build

Build the CLI:

```bash
go build -o repoctx ./cmd/repoctx
```

This creates:

```text
./repoctx
```

## Generate a Context Pack

Generate a context pack for the current repository:

```bash
./repoctx pack .
```

Expected output:

```text
Generated: .repoctx/ai-context.md
Generated: .repoctx/ai-context.json
```

Generated artifacts:

```text
.repoctx/
├── repoctx.db
├── ai-context.md
└── ai-context.json
```

## Generate Semantic Summaries (Optional)

Enable LLM-powered semantic summaries:

```bash
./repoctx pack . --summarize
```

Example output:

```text
Generated 5 file summaries
Reused 42 cached summaries
Generated: .repoctx/ai-context.md
Generated: .repoctx/ai-context.json
```

Summaries are cached using SHA256 file hashes. Unchanged files are not re-summarized.

## Inspect Generated Context

View repository metadata:

```bash
cat .repoctx/ai-context.md
```

Inspect generated JSON:

```bash
jq 'keys' .repoctx/ai-context.json
```

Inspect dependency relationships:

```bash
jq '.dependencies[:10]' .repoctx/ai-context.json
```

Inspect extracted symbols:

```bash
jq '.symbols[:10]' .repoctx/ai-context.json
```

## Ask for Relevant Files

Use the local recommendation engine to identify files related to a task:

```bash
./repoctx ask "markdown tree output"
```

Example output:

```text
Likely relevant files:

- internal/pack/markdown.go
  - path matches task term: markdown

- internal/scanner/scanner.go
  - dependency neighbor of internal/pack/markdown.go
```

This command uses:

* file paths
* extracted symbols
* dependency relationships
* semantic summaries (if available)

to recommend files that should be reviewed or provided to an LLM.

## Typical Development Workflow

```bash
go test ./...
go vet ./...
go build -o repoctx ./cmd/repoctx

./repoctx pack .

./repoctx ask "OAuth redirect loop"
./repoctx ask "add markdown output"
./repoctx ask "summary caching"
```

## Typical AI-Assisted Workflow

```text
1. Generate ai-context.md

2. Ask repoctx which files are relevant

3. Provide only those files to ChatGPT, Claude, or another coding assistant

4. Implement the change

5. Regenerate the context pack
```

This minimizes token usage while maximizing architecture awareness.

---

# Optional: Generate Semantic Summaries

By default the system works entirely offline without any LLM.

To enable semantic file summaries:

```bash
./repoctx pack . --summarize
```

This adds:

* file purpose summaries
* dependency descriptions
* side effect analysis
* risk labeling
* architectural notes

into:

* `ai-context.md`
* `ai-context.json`

Example summary:

```text
Purpose:
Handles authentication session lifecycle.

Dependencies:
db, oauth provider, session middleware

Risk:
high
```

---

# Incremental Summary Caching

Summaries are cached by SHA256 file hash.

Unchanged files are not re-summarized.

This dramatically reduces:

* LLM cost
* local inference time
* repeated processing

Example:

```text
Generated 3 file summaries
Reused 42 cached summaries
```

---

# Example `.env`

```env
REPOCTX_LLM_BASE_URL=http://localhost:11434/v1
REPOCTX_LLM_MODEL=qwen2.5-coder:7b
REPOCTX_LLM_API_KEY=ollama
```

The `.env` file is optional.

CLI flags override environment variables.

Priority order:

```text
CLI flags
↓
.env values
↓
built-in defaults
```

---

# Use with ChatGPT or Another Coding LLM

Open:

```text
.repoctx/ai-context.md
```

Paste it into your coding LLM and ask something like:

```text
I want to fix an OAuth redirect loop. Based on this context pack, what files do you need?
```

The LLM should then request only the relevant files instead of needing the entire repository.

---

# Task-Aware File Recommendations

The project includes an offline semantic recommendation engine.

Example:

```bash
./repoctx ask "OAuth redirect loop after login"
```

Potential output:

```text
Likely relevant files:

- backend/auth/routes.py
- backend/auth/service.py
- frontend/src/AuthProvider.tsx
- nginx.conf
```

The recommendation system combines:

* path analysis
* symbol matching
* semantic summaries
* dependency relationships

This helps determine which files should be sent to a frontier LLM.

---

# Why This Matters

Instead of sending:

```text
entire repository
```

the workflow becomes:

```text
repository
↓
semantic compiler
↓
compressed context
↓
task-aware file selection
↓
frontier LLM
```

This significantly reduces:

* token usage
* retrieval noise
* lost-in-the-middle failures
* irrelevant context

---

# Example AI Workflow

```text
1. Generate ai-context.md

2. Paste ai-context.md into ChatGPT

3. Ask:
   "I want to add OAuth refresh tokens."

4. LLM replies:
   "Please send these files..."

5. Only required files are shared
```

This avoids sending the entire repository.

---

# Repository Tree Output

Example:

```json
{
  "tree": [
    {
      "path": "backend/app/main.py",
      "type": "file",
      "language": "python",
      "summary": "FastAPI entrypoint."
    }
  ]
}
```

---

# Structured Symbol Output

Example:

```json
{
  "symbol": "create_session",
  "kind": "function",
  "path": "backend/app/auth/service.py",
  "inputs": ["user_id", "provider"],
  "outputs": ["session_token"],
  "side_effects": [
    "writes session row to database"
  ],
  "dependencies": [
    "db",
    "settings"
  ],
  "risk": "high",
  "summary": "Creates a persistent login session."
}
```

---

# AI-Friendly Markdown Output

Example:

```markdown
# AI Context Pack

## Repo Summary
FastAPI + React language learning app with LLM-generated flashcards.

## Architecture Tree
...

## High-Risk Areas
- Auth/session handling
- OAuth
- Billing

## Module Summaries

### backend/app/auth/routes.py

Purpose:
Authentication routes.

Dependencies:
- auth service
- db
- session middleware

Risk:
high
```

---

# Why Generate Both JSON and Markdown?

## JSON

Best for:

* machine-readable processing
* indexing
* querying
* future automation
* embeddings
* graph generation

## Markdown

Best for:

* direct LLM consumption
* human readability
* prompt injection into ChatGPT
* quick debugging

The JSON acts as the source of truth.

The Markdown acts as the AI context pack.

---

# Context Compression Philosophy

The goal is:

> Compress meaning, not syntax.

LLMs care more about:

* intent
* dependencies
* side effects
* architecture
* boundaries

than raw implementation details.

---

# Context Hierarchy Model

```text
Code
  ↓
Function Summary
  ↓
File Summary
  ↓
Module Summary
  ↓
System Summary
```

The system should progressively expand detail only when needed.

---

# Repository Intermediate Representation (Repository IR)

The project increasingly behaves like a compiler pipeline:

```text
source code
↓
syntax trees
↓
symbols
↓
dependency graph
↓
semantic summaries
↓
compressed repository IR
```

This is conceptually similar to traditional compiler pipelines:

```text
source code
↓
AST
↓
IR
↓
machine code
```

The generated context packs act as:

```text
repository IR for AI systems
```

This enables:

* scalable AI-assisted coding
* architecture-aware retrieval
* semantic compression
* graph-aware reasoning
* task-specific context assembly

---

# Planned Features

## MVP 1

* repository scanning
* tree generation
* markdown output
* file summaries

## MVP 2

* function extraction
* import graph
* dependency graph
* JSON output

## MVP 3

* Ollama summarization
* SQLite cache
* incremental updates

## MVP 4

* task-aware querying
* "what files are needed for this bug?"

## MVP 5

* embeddings
* semantic search
* hybrid retrieval

## MVP 6

* graph visualization
* architecture maps
* service relationship diagrams

---

# Long-Term Vision

This project is not just a repository summarizer.

The long-term goal is:

> A local semantic memory layer for AI-assisted software engineering.

Potential future directions:

* IDE integration
* MCP integration
* semantic diff generation
* automatic architecture documentation
* AI-oriented code review
* task-aware context packing
* distributed system dependency analysis
* event flow tracing

---

# Future Research Directions

## Semantic Compression

Instead of compressing syntax:

Compress:

* behavior
* architecture
* ownership
* side effects
* intent

## Context Ranking

Not all context deserves equal weight.

The future likely involves:

* dynamic relevance scoring
* dependency-aware ranking
* graph-based retrieval
* task-specific context construction

## AI-Friendly Codebases

Eventually repositories may be designed partly for:

* human developers
* AI coding systems

---

# Microservices Are Not Automatically Better

Microservices can reduce context scope because each service becomes a smaller reasoning unit.

However:

* distributed systems create hidden context
* APIs become dependencies
* events become dependencies
* runtime behavior spans repositories

A poorly designed microservice architecture can actually increase LLM reasoning difficulty.

The real solution is not microservices.

The real solution is:

* explicit boundaries
* stable interfaces
* good metadata
* semantic representations

---

# Inspirations

* repomix
* Sourcegraph
* tree-sitter
* pgvector
* RAG systems
* semantic code indexing systems

---

# Example Future Query

```bash
./repoctx ask "OAuth redirect loop after login"
```

Potential output:

```text
Likely relevant files:

- backend/auth/routes.py
- backend/auth/service.py
- frontend/src/AuthProvider.tsx
- nginx.conf
- session_middleware.py
```

---

# Design Principles

* Local-first
* Privacy-first
* AI-oriented
* Incremental
* Semantic over syntactic
* Token-efficiency focused
* LLM-friendly
* Language-agnostic architecture

---

# License

MIT
