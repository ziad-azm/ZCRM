> **Fetched from jira:** [ZCRM-4](https://ziadhosny007.atlassian.net/browse/ZCRM-4)  
> *Fetched 2026-08-22T19:43:00.174Z. Edit the sections below as needed; the planner reads this file verbatim.*


## Source — work item (from tracker)

**Title:** SETUP-3 — Frontend Project Structure (Angular)  
**Type:** Story  
**Status:** To Do  
**Assignee:** Ziad Hosny

### Description

As a developer, I want to set up the frontend structure in Angular, so that the UI code stays organized and maintainable.

Description: Scaffold the Angular application using a feature-based structure (core, shared, features) so screens and reusable pieces stay well organized as the app grows. A UI component library is added up front to keep the look consistent and professional. The story ends with a running app that can call the backend.

Tasks:

	Task: Run ng new and configure the routing module with a default route

	Task: Create the core, shared, and features folders

	Task: Install and theme the UI library (Angular Material or PrimeNG)

	Task: Create a base ApiService that reads the API URL from the environment

### Attachments

None.

---
# Story intake

Fill this template for each story you want planned. Keep it copy-paste-friendly: the planner reads **this file and the files in `attachments/`**, nothing else.

- Folder: `.squad/stories/project-setup-foundation/ZCRM-4/intake.md`
- Binaries (screenshots, PDFs, exports): put them in `attachments/` next to this file and list them below.
- Do **not** rely on external links (tracker URLs, wiki, chat) — the planner cannot open them. Paste the content you want considered.

This is **not** an implementation prompt. It is the input to the plan-generation meta-prompt bundled with squad-kit (`generate-plan.md` in the installed package).

---

## Feature

- **Feature name (display):**
- **Feature slug (folder under `plans/`):** `project-setup-foundation`

## Tracker (metadata only)

- **Tracker type:** `jira`
- **Work item id:** `ZCRM-4` *(used in filenames and plan tables; fill manually if empty)*
- **Work item type:** `Story`
- **Status:** `To Do`
- **Assignee:** `Ziad Hosny`
- **Labels:** ``

External tracker links are **not** followed by the planner. Keep the id for naming and traceability only.

---

## Title

*(Paste the work item title verbatim. Prefilled when `squad new-story` fetched from a tracker.)*

```
SETUP-3 — Frontend Project Structure (Angular)
```

---

## Description

*(Paste the full work item description. Prefilled when fetched from a tracker.)*

```
As a developer, I want to set up the frontend structure in Angular, so that the UI code stays organized and maintainable.

Description: Scaffold the Angular application using a feature-based structure (core, shared, features) so screens and reusable pieces stay well organized as the app grows. A UI component library is added up front to keep the look consistent and professional. The story ends with a running app that can call the backend.

Tasks:

	Task: Run ng new and configure the routing module with a default route

	Task: Create the core, shared, and features folders

	Task: Install and theme the UI library (Angular Material or PrimeNG)

	Task: Create a base ApiService that reads the API URL from the environment
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
