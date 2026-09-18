# Transition Vehicle Lifecycle Status API

This document shows how to call the REST API that transitions a vehicle's
lifecycle status, as defined by the
[Vehicle Status and Availability TRD](https://github.com/mshahkap33/ai-sdlc/blob/main/docs/trd/trd-vehicle-status-availability.md).

The vehicle status state machine allows the following statuses:
`available`, `reserved`, `rented`, `cleaning`, `maintenance`, `damaged`,
`retired`.

Both endpoints below require a JWT in the `Authorization` header, using the
standard `<scheme> <token>` format. The token's `roles` claim must contain a
role permitted for the endpoint being called (see
[Authentication](#authentication)).

## Running the server

```sh
go run ./cmd/server
```

The server reads `SERVER_ADDR` (default `:8080`) and `JWT_SECRET` from the
environment (or `.env`/`.env.local`); see the root [README](../../README.md)
for details.

## Setting up a token for the examples below

```sh
# The auth scheme used by this API's Authorization header.
AUTH_SCHEME="Bearer"
# A JWT signed with the server's JWT_SECRET, containing a "roles" claim.
JWT_TOKEN="<a signed JWT with the required role(s)>"
```

## 1. Submit a triggering event (system-to-system transition)

Automatically transitions a vehicle when an upstream service reports an
event such as `booking_confirmed`, `handover_completed`, `return_completed`,
`maintenance_scheduled`, or `damage_reported`. This endpoint is restricted to
callers whose JWT `roles` claim includes `system_service`.

```sh
curl -X POST "http://localhost:8080/api/v1/vehicles/11111111-1111-1111-1111-111111111111/status-events" \
  -H "Authorization: $AUTH_SCHEME $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
        "eventType": "booking_confirmed",
        "effectiveAt": "2026-01-15T09:30:00Z",
        "sourceSystem": "booking-service"
      }'
```

Successful response (`200 OK`):

```json
{
  "vehicleId": "11111111-1111-1111-1111-111111111111",
  "status": "reserved",
  "statusSince": "2026-01-15T09:30:00Z"
}
```

If the event's target status requires a reason (as configured in
`vehicle_status_triggers.requires_reason`, e.g. `damage_reported`), include a
`reason` field:

```sh
curl -X POST "http://localhost:8080/api/v1/vehicles/11111111-1111-1111-1111-111111111111/status-events" \
  -H "Authorization: $AUTH_SCHEME $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
        "eventType": "damage_reported",
        "reason": "Rear bumper collision reported by handover inspector",
        "sourceSystem": "inspection-service"
      }'
```

If the transition is not allowed from the vehicle's current status, the API
returns `409 Conflict`:

```json
{
  "error": "transition not allowed from current status"
}
```

## 2. Manually override status

Allows Service staff or Operations managers to set a vehicle's status
directly, with a mandatory reason. This endpoint requires the caller's JWT
`roles` claim to include `service_staff` or `operations_manager`.

```sh
curl -X PATCH "http://localhost:8080/api/v1/vehicles/11111111-1111-1111-1111-111111111111/status" \
  -H "Authorization: $AUTH_SCHEME $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
        "newStatus": "maintenance",
        "reason": "Scheduled brake inspection"
      }'
```

Successful response (`200 OK`):

```json
{
  "vehicleId": "11111111-1111-1111-1111-111111111111",
  "status": "maintenance",
  "statusSince": "2026-01-15T10:05:00Z"
}
```

## Error responses

| Status | Cause |
| --- | --- |
| `400 Bad Request` | Malformed `vehicleId`/request body, unknown `eventType`, invalid `newStatus`, or a missing/oversized `reason` |
| `401 Unauthorized` | Missing, malformed, or invalid JWT in the `Authorization` header |
| `403 Forbidden` | Authenticated caller lacks a required role |
| `404 Not Found` | `vehicleId` does not reference an existing, non-deleted vehicle |
| `409 Conflict` | The requested transition is not allowed from the vehicle's current status |

## Authentication

Tokens must be signed with HS256 (or higher) using the server's
`JWT_SECRET` and include a `roles` claim (array of strings) and `sub`
(the acting user or system account, recorded as `created_by` on every status
history entry). For local testing you can mint a token with any JWT library
using the same value configured for `JWT_SECRET`, for example with the
[jwt.io](https://jwt.io) debugger or a short script using
`github.com/golang-jwt/jwt/v5`.
