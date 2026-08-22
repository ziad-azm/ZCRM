> **Fetched from jira:** [ZCRM-6](https://ziadhosny007.atlassian.net/browse/ZCRM-6)  
> *Fetched 2026-08-22T19:43:44.369Z. Edit the sections below as needed; the planner reads this file verbatim.*


## Source — work item (from tracker)

**Title:** SETUP-5 — Environment Configuration  
**Type:** Story  
**Status:** To Do  
**Assignee:** Ziad Hosny

### Description

As a developer, I want to configure environment variables and secrets, so that configuration is kept out of the code and safe across environments.

Description: Separate configuration from code so secrets (DB credentials, JWT secret) never get committed and dev/prod settings can differ safely. Both apps read their settings from environment files, and a committed example file documents what's required without exposing real values.

Tasks:

	Task: Add a config loader in Go that reads env vars with sensible defaults

	Task: Create .env.example and ensure the real .env is git-ignored

	Task: Configure Angular environment.ts and environment.prod.ts

### Attachments

None.

---
# Story intake

Fill this template for each story you want planned. Keep it copy-paste-friendly: the planner reads **this file and the files in `attachments/`**, nothing else.

- Folder: `.squad/stories/project-setup-foundation/ZCRM-6/intake.md`
- Binaries (screenshots, PDFs, exports): put them in `attachments/` next to this file and list them below.
- Do **not** rely on external links (tracker URLs, wiki, chat) — the planner cannot open them. Paste the content you want considered.

This is **not** an implementation prompt. It is the input to the plan-generation meta-prompt bundled with squad-kit (`generate-plan.md` in the installed package).

---

## Feature

- **Feature name (display):**
- **Feature slug (folder under `plans/`):** `project-setup-foundation`

## Tracker (metadata only)

- **Tracker type:** `jira`
- **Work item id:** `ZCRM-6` *(used in filenames and plan tables; fill manually if empty)*
- **Work item type:** `Story`
- **Status:** `To Do`
- **Assignee:** `Ziad Hosny`
- **Labels:** ``

External tracker links are **not** followed by the planner. Keep the id for naming and traceability only.

---

## Title

*(Paste the work item title verbatim. Prefilled when `squad new-story` fetched from a tracker.)*

```
SETUP-5 — Environment Configuration
```

---

## Description

*(Paste the full work item description. Prefilled when fetched from a tracker.)*

```
As a developer, I want to configure environment variables and secrets, so that configuration is kept out of the code and safe across environments.

Description: Separate configuration from code so secrets (DB credentials, JWT secret) never get committed and dev/prod settings can differ safely. Both apps read their settings from environment files, and a committed example file documents what's required without exposing real values.

Tasks:

	Task: Add a config loader in Go that reads env vars with sensible defaults

	Task: Create .env.example and ensure the real .env is git-ignored

	Task: Configure Angular environment.ts and environment.prod.ts
```

---

## Acceptance criteria

*(Checklist, bullets, Gherkin, etc. Prefilled for Azure DevOps when the work item has acceptance criteria.)*

```

```

---

## Attachments

Place files in `attachments/` next to this `intake.md`, then list them here so the planner knows what to open.

| File (relative to this folder) | What it is |
| ------------------------------ | ---------- |
| *(e.g. `attachments/flow.png`)* | *(e.g. UX flow)* |

*(Add rows per file. If none, write "None.")*

---

## Dependencies

- **Blocked by / related ids:** (tracker ids only; optional short note)
- **Depends on code areas or other stories:**

## Extra notes (optional)

- Anything not captured above (e.g. chat context) — keep short.

## Technical hints (optional)

- APIs, screens, services already discussed. Repos/roots: `.`. Primary language: `typescript`.

## Out of scope

- What this story explicitly does **not** cover:
