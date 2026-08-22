> **Fetched from jira:** [ZCRM-2](https://ziadhosny007.atlassian.net/browse/ZCRM-2)  
> *Fetched 2026-08-22T19:35:54.357Z. Edit the sections below as needed; the planner reads this file verbatim.*


## Source — work item (from tracker)

**Title:** SETUP-1 — Initialize the Repository  
**Type:** Story  
**Status:** To Do  
**Assignee:** Ziad Hosny

### Description

As a developer, I want to initialize the Git repository with a clear folder structure, so that the project has a consistent and organized starting point.

Description: Set up the single source of truth for the codebase. We use a monorepo so the Angular frontend and Golang backend live side by side and share tooling and versioning, while staying cleanly separated in their own folders. This story ends with a working repo on the remote that any developer can clone and understand from the README alone.

Tasks:

	Task: Create the remote repository and clone it locally

	Task: Add the /backend + /frontend folder structure and a root README with setup steps

	Task: Add .gitignore for Go, Node/Angular, and environment files

	Task: Create and push main + develop branches, and protect main

### Attachments

None.

---
# Story intake

Fill this template for each story you want planned. Keep it copy-paste-friendly: the planner reads **this file and the files in `attachments/`**, nothing else.

- Folder: `.squad/stories/project-setup-foundation/ZCRM-2/intake.md`
- Binaries (screenshots, PDFs, exports): put them in `attachments/` next to this file and list them below.
- Do **not** rely on external links (tracker URLs, wiki, chat) — the planner cannot open them. Paste the content you want considered.

This is **not** an implementation prompt. It is the input to the plan-generation meta-prompt bundled with squad-kit (`generate-plan.md` in the installed package).

---

## Feature

- **Feature name (display):**
- **Feature slug (folder under `plans/`):** `project-setup-foundation`

## Tracker (metadata only)

- **Tracker type:** `jira`
- **Work item id:** `ZCRM-2` *(used in filenames and plan tables; fill manually if empty)*
- **Work item type:** `Story`
- **Status:** `To Do`
- **Assignee:** `Ziad Hosny`
- **Labels:** ``

External tracker links are **not** followed by the planner. Keep the id for naming and traceability only.

---

## Title

*(Paste the work item title verbatim. Prefilled when `squad new-story` fetched from a tracker.)*

```
SETUP-1 — Initialize the Repository
```

---

## Description

*(Paste the full work item description. Prefilled when fetched from a tracker.)*

```
As a developer, I want to initialize the Git repository with a clear folder structure, so that the project has a consistent and organized starting point.

Description: Set up the single source of truth for the codebase. We use a monorepo so the Angular frontend and Golang backend live side by side and share tooling and versioning, while staying cleanly separated in their own folders. This story ends with a working repo on the remote that any developer can clone and understand from the README alone.

Tasks:

	Task: Create the remote repository and clone it locally

	Task: Add the /backend + /frontend folder structure and a root README with setup steps

	Task: Add .gitignore for Go, Node/Angular, and environment files

	Task: Create and push main + develop branches, and protect main
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
