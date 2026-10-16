# Budgets, Events and Portfolios

This document explains how Warchest manages Events and Portfolios, how they relate to other services, and what is currently implemented.

## 1. Overview

Warchest manages two types of budgets: **Events** and **Portfolios**.

| | Event | Portfolio |
|---|---|---|
| Purpose | Manage an individual activity | Manage an ongoing Team budget or an annual cross-team event budget |
| Example | leetcode workshop, Welcome BBQ | Marketing Team; Starlight |
| Created by | Authorised Warchest users | Imported from AMBUSH |
| Lifecycle | Draft → In Progress → Closed | Ongoing; supports annual reset |
| Linked to | Creator (`user_id`) | AMBUSH Team (`team_id`) |
| Financial data | Money Module | Money Module |

**Key decisions:**

- Events and Portfolios are peers. An Event does **not** belong to a Portfolio.
- Both use the `budgets` table, distinguished by `type = EVENT` or `PORTFOLIO`.
- Warchest manages metadata and lifecycle, **not money**.
- Funding an Event from a Portfolio creates a financial relationship in Money Module, not organisational ownership.

For example, Marketing may fund Starlight without Starlight becoming part of Marketing's Portfolio.

## 2. System Responsibilities

| System | Owns |
|---|---|
| **AMBUSH** | Users, Teams, Roles and Memberships |
| **Warchest** | Local Budget IDs, Event metadata and lifecycle, Portfolio-to-Team mapping |
| **Money Module** | Allowances, budgets, expenditure, balances, allocations, transfers and financial history |

**User information:** Warchest has no local `users` table. It stores external IDs where needed (for example, Event `created_by`). Names, emails and memberships come from AMBUSH. 

**Financial information:** Warchest has no monetary columns. Financial values come from Money Module.

## 3. Database Design

One table stores both types:

```text
budgets
├── EVENT
│   ├── Leetcode Workshop
│   └── Welcome BBQ
└── PORTFOLIO
    ├── Marketing
    └── Sponsorship
```

### Shared fields

| Field | Purpose |
|---|---|
| `id UUID` | Warchest's local Budget ID |
| `type` | `EVENT` or `PORTFOLIO` |
| `society_id TEXT` | External Society scope |
| `name TEXT` | Event name or imported Team/Portfolio label |
| `description TEXT` | Optional description |
| `money_budget_id UUID` | Nullable Money Module reference; linking is not yet implemented |
| `archived_at` | When hidden from normal active listings |
| `created_at`, `updated_at` | Audit timestamps |

### Type-specific fields

| Field | Event | Portfolio |
|---|---|---|
| `team_id TEXT` | Not used | Required by business rules |
| `event_time` | Optional | Not used |
| `location` | Optional | Not used |
| `status` | Required | Not used |
| `created_by UUID` | Required | Not used for system imports |

## 4. Events

An Event is a single activity created by an authorised user.

### Lifecycle

```text
DRAFT ── Start ──> IN_PROGRESS ── Close ──> CLOSED
  └──────────── Cancel ──────────────────> CLOSED
```

| State | Meaning |
|---|---|
| `DRAFT` | Planning; metadata can be edited |
| `IN_PROGRESS` | Explicitly started |
| `CLOSED` | Completed or cancelled; cannot be edited or restarted |

**Rules:**

- `POST /events` requires an explicit `status`: **Create** sends `IN_PROGRESS`; **Create Draft** sends `DRAFT`.
- Missing status, `CLOSED` and other creation values are rejected by Service validation.
- `event_time` does not automatically change lifecycle state.
- A closed Event can be archived, hiding it from active lists without deleting financial history.
- Events cannot be reset or reopened; a new activity requires a new Event.
- Creating or starting an Event does not currently require a Money Module link.

### Create Event API

`POST /events`

The intended authenticated request shape is:

```json
{
  "event_name": "Welcome BBQ",
  "event_time": "2026-11-01T18:00:00Z",
  "location": "UNSW",
  "description": "Annual society event",
  "society_id": "society-123",
  "status": "DRAFT"
}
```


## 5. Portfolios

A Portfolio is an ongoing budget-management unit associated with an AMBUSH Team.

```text
AMBUSH Team ── team_id ──> Warchest Portfolio
```


**Finding a user's Portfolios:**

```text
User → AMBUSH Membership Service → team_ids → Warchest Portfolios
```

Membership identifies relevant Portfolios, but permission checks are still required for actions.

### Annual reset

Reset restores the Portfolio to its **configured annual allowance** in Money Module, without asking the user to enter an amount.

```text
Configured annual allowance: $1,000
Current available balance: $250
After reset:              $1,000
```

- Reset keeps the same Warchest Portfolio ID and `team_id`.
- Money Module owns the allowance, reset operation, history and safe retry behaviour.
