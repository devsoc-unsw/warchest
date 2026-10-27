# Budgets, Events and Portfolios

This document explains the main concepts, system responsibilities, lifecycles and database tables.

## 1. Concepts and Relationships

A Budget is the common Warchest record for an **Event** or a **Portfolio**. The `type` field tells them apart.

| | Event | Portfolio |
|---|---|---|
| Purpose | Manage an individual activity | Manage an ongoing Team budget or an annual cross-team event budget |
| Example | leetcode workshop, Welcome BBQ | Marketing Team; Starlight |
| Created by | Authorised Warchest users | Imported from AMBUSH |
| Lifecycle | Draft -> In Progress -> Closed | Ongoing; supports annual reset |
| Linked to | Member (`user_id`) | AMBUSH Team (`team_id`) |
| Financial data | Money Module | Money Module |

- Events and Portfolios are at the same level. An Event does **not** belong to a Portfolio.
- A Portfolio is linked to an AMBUSH Team through `team_id`.
- Warchest stores their details and lifecycle. Money Module stores their financial data.

## 2. System Responsibilities

Warchest has no local `users` table. It saves the user ID in `created_by`, getting names and memberships from AMBUSH.

Membership identifies the Portfolios related to the user. 
For example, Silvana is a member of the DevSoc Marketing Team in AMBUSH. Warchest uses the team_id as identification for the Marketing Portfolio. Her permissions decide what she can do with it.

Warchest has no monetary columns. Amounts and financial history come from Money Module.

## 3. Lifecycle

### Event

| State | Meaning |
|---|---|
| `DRAFT` | Event in planning; details can be edited |
| `IN_PROGRESS` | Started |
| `CLOSED` | Closed and archived; cannot be edited or restarted |

There are two creation options:

- **Create Draft** creates an Event in `DRAFT`.
- **Create** creates an Event directly in `IN_PROGRESS`.

After creation:

```text
DRAFT - Start -> IN_PROGRESS - Close -> CLOSED
  └──────────── Close ──────────────────> CLOSED
```

- The scheduled `event_time` does not automatically change the state.
- Cancel in editing only discards unsaved changes.
- Closing an Event or a draft event also archives it. It moves to archive tab. Its records are kept and can still be viewed there.
- Events cannot be reset or reopened. A new activity needs a new Event.
- Events cannot be deleted, and they can only be closed (archived).

### Portfolio

A Portfolio is ongoing. It does not use the three Event states. Each Portfolio is identified by its AMBUSH team_id.

Annual reset restores its **configured annual allowance** from Money Module. The user does not enter a new amount.

```text
Configured annual allowance: $1,000
Current available balance: $250
After reset:              $1,000
```

- Archive the previous year's records in archive tab and keep them for reference.
- Create an annual record for the new year.
- Reset does not affect any Events.
- Money Module handles the financial reset and keeps the financial history.

## 4. Data Model

### budgets

This table stores Events and Portfolios. The `type` field tells us which one it is.

**Fields used by both:**

| Field | Purpose |
|---|---|
| `id UUID` | Warchest Budget ID |
| `type` | `EVENT` or `PORTFOLIO` |
| `society_id TEXT` | The Society this Budget belongs to |
| `name TEXT` | Event or Portfolio name |
| `description TEXT` | Optional description |
| `money_budget_id UUID` | Money Module reference; TODO |
| `created_at`, `updated_at` | When the local Warchest record was saved and last updated |

For a Portfolio, Warchest keeps a local copy of Portfolio details; AMBUSH still manages the Team. For a Portfolio, created_at is when Warchest first saves this copy, and updated_at is when Warchest last updates it. These are not the portfolio’s creation or update times in AMBUSH.

**Fields used differently:**

| Field | Event | Portfolio |
|---|---|---|
| `team_id TEXT` | Not used | AMBUSH portfolio ID |
| `created_by UUID` | Creator's user ID | Not used |
| `event_time` | Scheduled date and time | Not used |
| `location` | Event location | Not used |
| `status` | `DRAFT`, `IN_PROGRESS` or `CLOSED` | Not used |
| `archived_at` | When the Event was closed and archived | Not used for annual reset |

### portfolio_years - yearly records for a Portfolio

Each row is one year for one Portfolio. It stores the year and its archive time, not the requests or amounts.
When a Portfolio is reset, the old year is archived and a new year record is added.

| Field | What it stores |
|---|---|
| `society_id` | Which Society the Portfolio belongs to |
| `team_id` | Which Portfolio this year belongs to |
| `year` | The year, such as 2026 |
| `archived_at` | Archive time; `NULL` means not archived |

For example, after Marketing resets for 2027:

| society_id | team_id | year | archived_at |
|---|---|---|---|
| DevSoc ID | Marketing portfolio ID | 2026 | Time of the reset |
| DevSoc ID | Marketing portfolio ID | 2027 | `NULL` |

AMBUSH manages portfolio. We do not handle a portfolio being removed or disabled here.
