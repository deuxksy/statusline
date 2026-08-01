# Explanation: The Diátaxis Documentation Framework in OKF

This document explains why and how OKF (Open Knowledge Framework) v0.2 adopts the [Diátaxis Framework](https://diataxis.fr/).

---

## The Diátaxis Quadrants

Diátaxis organizes documentation into four distinct modes according to two axes:
1. **Learning vs Working**
2. **Practical vs Theoretical**

```
                       LEARNING (Education)
                                │
               Tutorials        │       How-To Guides
          (Learning-oriented)  │    (Task-oriented)
                                │
  PRACTICAL ────────────────────┼──────────────────── THEORETICAL
  (Action)                      │                     (Knowledge)
                                │
               Reference        │       Explanation
         (Information-oriented) │ (Understanding-oriented)
                                │
                        WORKING (Execution)
```

### Application in OKF v0.2

1. **Tutorials (`docs/okf/tutorials/`)**: Step-by-step onboarding for beginners to build and run statusline for the first time.
2. **How-To Guides (`docs/okf/how-to-guides/`)**: Goal-focused recipes for specific tasks (configuring elements, adding a new CLI adapter).
3. **Reference (`docs/okf/reference/`)**: Complete specifications and schemas (STDIN JSON payloads, config parameters, software architecture).
4. **Explanation (`docs/okf/explanation/`)**: Design philosophy, single-line rendering rationale, and fail-soft constraints.
