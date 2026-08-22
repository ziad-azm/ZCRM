> **Fetched from jira:** [ZCRM-9](https://ziadhosny007.atlassian.net/browse/ZCRM-9)  
> *Fetched 2026-08-22T20:20:07.679Z. Edit the sections below as needed; the planner reads this file verbatim.*


## Source — work item (from tracker)

**Title:** AUTH-1 — User Registration  
**Type:** Story  
**Status:** In Progress  
**Assignee:** Ziad Hosny

### Description

As a user, I want to register an account, so that I can access the CRM.

Description: New users create an account with their name, email, and password. Emails must be unique, and passwords are securely hashed before storage so plaintext is never kept. This is the entry point into the system and the foundation for every other authenticated action.

Tasks:

	Task: Create the users table via migration (id, name, email unique, password_hash, role, timestamps)

	Task: Implement POST /auth/register with validation and bcrypt hashing

	Task: Build the registration form with client-side validation

	Task: Write tests for valid input and duplicate-email cases

### Attachments

None.

---
# Story intake

Fill this template for each story you want planned. Keep it copy-paste-friendly: the planner reads **this file and the files in `attachments/`**, nothing else.

- Folder: `.squad/stories/authentication-user-management/ZCRM-9/intake.md`
- Binaries (screenshots, PDFs, exports): put them in `attachments/` next to this file and list them below.
- Do **not** rely on external links (tracker URLs, wiki, chat) — the planner cannot open them. Paste the content you want considered.

This is **not** an implementation prompt. It is the input to the plan-generation meta-prompt bundled with squad-kit (`generate-plan.md` in the installed package).

---

## Feature

- **Feature name (display):**
- **Feature slug (folder under `plans/`):** `authentication-user-management`

## Tracker (metadata only)

- **Tracker type:** `jira`
- **Work item id:** `ZCRM-9` *(used in filenames and plan tables; fill manually if empty)*
- **Work item type:** `Story`
- **Status:** `In Progress`
- **Assignee:** `Ziad Hosny`
- **Labels:** ``

External tracker links are **not** followed by the planner. Keep the id for naming and traceability only.

---

## Title

*(Paste the work item title verbatim. Prefilled when `squad new-story` fetched from a tracker.)*

```
AUTH-1 — User Registration
```

---

## Description

*(Paste the full work item description. Prefilled when fetched from a tracker.)*

```
As a user, I want to register an account, so that I can access the CRM.

Description: New users create an account with their name, email, and password. Emails must be unique, and passwords are securely hashed before storage so plaintext is never kept. This is the entry point into the system and the foundation for every other authenticated action.

Tasks:

	Task: Create the users table via migration (id, name, email unique, password_hash, role, timestamps)

	Task: Implement POST /auth/register with validation and bcrypt hashing

	Task: Build the registration form with client-side validation

	Task: Write tests for valid input and duplicate-email cases
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
