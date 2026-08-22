> **Fetched from jira:** [ZCRM-3](https://ziadhosny007.atlassian.net/browse/ZCRM-3)  
> *Fetched 2026-08-22T19:37:09.791Z. Edit the sections below as needed; the planner reads this file verbatim.*


## Source — work item (from tracker)

**Title:** SETUP-2 — Backend Project Structure (Golang)  
**Type:** Story  
**Status:** To Do  
**Assignee:** Ziad Hosny

### Description

As a developer, I want to set up the backend structure in Golang, so that features can be built in an organized and scalable way.

Description: Scaffold the Go backend using a clean layered architecture (handlers → services → repositories) so business logic, HTTP handling, and data access stay separated. This makes the codebase easy to extend as new features are added and easy for others to navigate. The story ends with a running HTTP server exposing a health-check.

Tasks:

	Task: Run go mod init and choose the HTTP router (chi/gin/echo)

	Task: Create the layered folders (handlers, services, repositories, models) and the cmd/ entrypoint

	Task: Implement the GET /health endpoint returning 200

	Task: Add structured logging and graceful shutdown

### Attachments

None.

---
# Story intake

Fill this template for each story you want planned. Keep it copy-paste-friendly: the planner reads **this file and the files in `attachments/`**, nothing else.

- Folder: `.squad/stories/project-setup-foundation/ZCRM-3/intake.md`
- Binaries (screenshots, PDFs, exports): put them in `attachments/` next to this file and list them below.
- Do **not** rely on external links (tracker URLs, wiki, chat) — the planner cannot open them. Paste the content you want considered.

This is **not** an implementation prompt. It is the input to the plan-generation meta-prompt bundled with squad-kit (`generate-plan.md` in the installed package).

---

## Feature

- **Feature name (display):**
- **Feature slug (folder under `plans/`):** `project-setup-foundation`

## Tracker (metadata only)

- **Tracker type:** `jira`
- **Work item id:** `ZCRM-3` *(used in filenames and plan tables; fill manually if empty)*
- **Work item type:** `Story`
- **Status:** `To Do`
- **Assignee:** `Ziad Hosny`
- **Labels:** ``

External tracker links are **not** followed by the planner. Keep the id for naming and traceability only.

---

## Title

*(Paste the work item title verbatim. Prefilled when `squad new-story` fetched from a tracker.)*

```
SETUP-2 — Backend Project Structure (Golang)
```

---

## Description

*(Paste the full work item description. Prefilled when fetched from a tracker.)*

```
As a developer, I want to set up the backend structure in Golang, so that features can be built in an organized and scalable way.

Description: Scaffold the Go backend using a clean layered architecture (handlers → services → repositories) so business logic, HTTP handling, and data access stay separated. This makes the codebase easy to extend as new features are added and easy for others to navigate. The story ends with a running HTTP server exposing a health-check.

Tasks:

	Task: Run go mod init and choose the HTTP router (chi/gin/echo)

	Task: Create the layered folders (handlers, services, repositories, models) and the cmd/ entrypoint

	Task: Implement the GET /health endpoint returning 200

	Task: Add structured logging and graceful shutdown
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
