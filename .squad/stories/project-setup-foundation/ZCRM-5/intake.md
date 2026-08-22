> **Fetched from jira:** [ZCRM-5](https://ziadhosny007.atlassian.net/browse/ZCRM-5)  
> *Fetched 2026-08-22T19:43:24.809Z. Edit the sections below as needed; the planner reads this file verbatim.*


## Source — work item (from tracker)

**Title:** SETUP-4 — Database & Migrations (PostgreSQL)  
**Type:** Story  
**Status:** To Do  
**Assignee:** Ziad Hosny

### Description

As a developer, I want to set up PostgreSQL and a migration workflow, so that the application can store and manage data reliably.

Description: Provision PostgreSQL and connect the backend to it with a pooled connection. A migration tool is introduced so every schema change is versioned and repeatable across environments, avoiding manual database edits. The story ends with the backend connecting to the database on startup.

Tasks:

	Task: Add a docker-compose.yml with a PostgreSQL service

	Task: Configure the pooled DB connection in Go

	Task: Set up the migration tool (golang-migrate) and the initial migration

	Task: Add scripts/commands to run and roll back migrations

### Attachments

None.

---
# Story intake

Fill this template for each story you want planned. Keep it copy-paste-friendly: the planner reads **this file and the files in `attachments/`**, nothing else.

- Folder: `.squad/stories/project-setup-foundation/ZCRM-5/intake.md`
- Binaries (screenshots, PDFs, exports): put them in `attachments/` next to this file and list them below.
- Do **not** rely on external links (tracker URLs, wiki, chat) — the planner cannot open them. Paste the content you want considered.

This is **not** an implementation prompt. It is the input to the plan-generation meta-prompt bundled with squad-kit (`generate-plan.md` in the installed package).

---

## Feature

- **Feature name (display):**
- **Feature slug (folder under `plans/`):** `project-setup-foundation`

## Tracker (metadata only)

- **Tracker type:** `jira`
- **Work item id:** `ZCRM-5` *(used in filenames and plan tables; fill manually if empty)*
- **Work item type:** `Story`
- **Status:** `To Do`
- **Assignee:** `Ziad Hosny`
- **Labels:** ``

External tracker links are **not** followed by the planner. Keep the id for naming and traceability only.

---

## Title

*(Paste the work item title verbatim. Prefilled when `squad new-story` fetched from a tracker.)*

```
SETUP-4 — Database & Migrations (PostgreSQL)
```

---

## Description

*(Paste the full work item description. Prefilled when fetched from a tracker.)*

```
As a developer, I want to set up PostgreSQL and a migration workflow, so that the application can store and manage data reliably.

Description: Provision PostgreSQL and connect the backend to it with a pooled connection. A migration tool is introduced so every schema change is versioned and repeatable across environments, avoiding manual database edits. The story ends with the backend connecting to the database on startup.

Tasks:

	Task: Add a docker-compose.yml with a PostgreSQL service

	Task: Configure the pooled DB connection in Go

	Task: Set up the migration tool (golang-migrate) and the initial migration

	Task: Add scripts/commands to run and roll back migrations
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
