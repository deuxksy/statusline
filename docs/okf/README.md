# statusline v0.2 Documentation (OKF)

Welcome to the **Open Knowledge Framework (OKF)** documentation for `statusline` v0.2.

This documentation is structured according to the [Diátaxis Framework](https://diataxis.fr/), separating content by developer intent and need into four distinct categories.

---

## 🧭 Diátaxis Navigation System

```
                  LEARNING-ORIENTED
                          │
             Tutorials    │    How-To Guides
                          │
  PRACTICAL ──────────────┼────────────── THEORETICAL
  WORK                    │               KNOWLEDGE
                          │
             Reference    │    Explanation
                          │
                  INFORMATION-ORIENTED
```

### 1. 🎓 [Tutorials](file:///home/crong/git/statusline/docs/okf/tutorials/index.md) *(Learning-Oriented)*
Hands-on lessons for getting started with `statusline` from scratch.
- [getting-started.md](file:///home/crong/git/statusline/docs/okf/tutorials/getting-started.md): Installation, initialization, and running your first statusline pipeline.

### 2. 🛠️ [How-To Guides](file:///home/crong/git/statusline/docs/okf/how-to-guides/index.md) *(Task-Oriented)*
Step-by-step problem-solving guides for specific user goals.
- [configure-elements.md](file:///home/crong/git/statusline/docs/okf/how-to-guides/configure-elements.md): Customizing layout elements, colors, and tmux integration.
- [add-new-cli-adapter.md](file:///home/crong/git/statusline/docs/okf/how-to-guides/add-new-cli-adapter.md): Implementing a new CLI host adapter.

### 3. 📖 [Reference](file:///home/crong/git/statusline/docs/okf/reference/index.md) *(Information-Oriented)*
Technical descriptions, schemas, and specifications.
- [cli-payload-spec.md](file:///home/crong/git/statusline/docs/okf/reference/cli-payload-spec.md): Complete STDIN JSON schemas for Antigravity (`agy`), Claude Code, and Codex.
- [configuration-schema.md](file:///home/crong/git/statusline/docs/okf/reference/configuration-schema.md): `config.json` schema, layout properties, and threshold settings.
- [architecture.md](file:///home/crong/git/statusline/docs/okf/reference/architecture.md): Package layout, interface definitions, and data pipeline.

### 4. 💡 [Explanation](file:///home/crong/git/statusline/docs/okf/explanation/index.md) *(Understanding-Oriented)*
Background concepts, design rationale, and performance constraints.
- [design-decisions.md](file:///home/crong/git/statusline/docs/okf/explanation/design-decisions.md): Architectural decisions (Single-line rendering, fail-soft pipeline, <5ms budget).
- [diataxis-framework.md](file:///home/crong/git/statusline/docs/okf/explanation/diataxis-framework.md): How OKF v0.2 applies Diátaxis for developer documentation.
