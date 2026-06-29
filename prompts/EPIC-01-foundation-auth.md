# Epic 01 — Foundation & Authentication

## Context
This is the first epic of the `event-mpos` project. Start by exploring the repository structure.
If the project folders (`android/`, `backend/`, `web/`) don't exist yet, scaffold them with sensible initial architecture before writing any feature code.

Read the `.cursorrules` file first to understand the overall system, tech stack, and conventions.

## Goal
Build the authentication foundation that all other epics will depend on.

Users have one of three roles: **Admin**, **BA (Beauty Advisor)**, or **Warehouse**.
A BA account is always assigned to a specific event — this mapping must be part of this epic.

## What Needs to Work When This Epic Is Done
- A user can log in with email and password and receive a JWT token
- The token is refreshed automatically when it expires
- Each role (Admin, BA, Warehouse) has appropriate access boundaries enforced on the API
- The Android app has a login screen; the token is stored securely on device
- The web dashboard has a login page with role-based routing after login
- An Admin can create user accounts and assign BA accounts to events via the web dashboard
- A BA who logs into the Android app sees only their assigned event

## Things Worth Paying Attention To
- Token storage on Android must be secure — explore the right approach for this
- Think about how the BA-to-event assignment affects what the Android app shows after login
- The backend should be structured so adding new protected endpoints later is straightforward

## When Done

```bash
git checkout -b feature/epic-01-auth
git add .
git commit -m "feat: epic-01 foundation and authentication"
git push origin feature/epic-01-auth
```

Then open a Pull Request to `main` on https://github.com/krnskrll/event-mpos
