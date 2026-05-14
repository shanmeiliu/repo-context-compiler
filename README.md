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
* Support local LLMs through Ollama
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
│   └── uses local LLMs to generate semantic metadata
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

| Component  | Technology        |
| ---------- | ----------------- |
| Language   | Go                |
| Parser     | tree-sitter       |
| Local LLM  | Ollama            |
| Cache DB   | SQLite            |
| Embeddings | Ollama embeddings |
| Output     | JSON + Markdown   |

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

# Why Ollama?

Ollama enables fully local workflows.

This project is intended to work offline and locally:

* no cloud dependency
* no repo upload
* no privacy concerns
* supports local embeddings
* supports local summarization

Potential models:

* qwen2.5-coder
* deepseek-coder
* llama3
* codellama
* nomic embeddings

---

# Core Workflow

```text
1. Scan repository
2. Parse code structure
3. Build semantic summaries
4. Store cached representations
5. Generate compressed AI context pack
6. Ask LLM what files it needs
7. Provide only required source files
```

---

# Example Usage

## Generate context pack

```bash
repoctx pack .
```

Outputs:

```text
.ai-context/
├── ai-context.json
├── ai-context.md
└── embeddings.db
```

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
repoctx ask \
  --task "OAuth redirect loop after login"
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

