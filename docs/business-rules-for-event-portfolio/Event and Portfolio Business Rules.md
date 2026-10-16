# Event and Portfolio Business Rules

## 1. Create Event

The frontend has two buttons:

| Button | Status |
|---|---|
| Create | `IN_PROGRESS` |
| Create Draft | `DRAFT` |

The backend uses the status sent by the frontend. There is no default status.

### Budget

| Status | Budget | Allowed |
|---|---|---|
| `DRAFT` | Missing or null | Yes |
| `DRAFT` | Non-negative integer | Yes |
| `IN_PROGRESS` | Missing or null | No; return `400` |
| `IN_PROGRESS` | Non-negative integer | Yes |

- Missing budget does not automatically turn an Event into a draft.
- Budget uses `uint64`, not a floating-point type. Decimal and negative inputs are rejected.
- Warchest does not store monetary amounts. Saving the supplied budget in Money Module is still TODO.

### Other fields

- Event name and Society ID cannot be blank.
- Creator ID must be a UUID. Getting it from authentication is TODO.
- `CLOSED` is not allowed when creating an Event.
- Event time, location and description are optional.

## 2. Start Event

- Moving from `DRAFT` to `IN_PROGRESS` must also require a budget.
- TODO: implement this rule in the Start Event service; this change only covers Create Event.

## 3. Portfolio

- TODO: document Portfolio business rules.
