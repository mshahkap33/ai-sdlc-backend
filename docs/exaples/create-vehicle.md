# Create Vehicle API

Onboards a new vehicle master data record, as defined by the
[Vehicle Onboarding TRD](https://github.com/mshahkap33/ai-sdlc/blob/main/docs/trd/trd-vehicle-onboarding.md).

## Request

| | |
| --- | --- |
| Method | `POST` |
| URL | `/api/v1/vehicles` |
| Auth | `Authorization` header using the RFC&nbsp;6750 bearer scheme, with a JWT granting the `service_staff` role |

### Body fields

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `vin` | string | yes | 17-character VIN (letters `I`, `O`, `Q` not allowed), must be unique |
| `plate_number` | string | yes | alphanumeric, optional hyphen/space separators, must be unique among active vehicles |
| `make` | string | yes | |
| `model` | string | yes | |
| `year` | integer | yes | 4-digit year, not greater than next calendar year |
| `category_id` | string | yes | ID of an existing, active vehicle category |
| `mileage` | number | yes | `>= 0` |
| `fuel_level` | number | yes | `0`-`100` |
| `location` | string | yes | current location identifier/description |
| `condition_notes` | string | no | |

## Example request

Set `$JWT` to a valid access token issued by the identity provider (the
`Authorization` header uses the RFC&nbsp;6750 bearer scheme):

```sh
export AUTH_SCHEME="Bearer"
curl -X POST https://api.example.com/api/v1/vehicles \
  -H "Authorization: $AUTH_SCHEME $JWT" \
  -H "Content-Type: application/json" \
  -d '{
        "vin": "1HGCM82633A004352",
        "plate_number": "ABC-1234",
        "make": "Honda",
        "model": "Accord",
        "year": 2023,
        "category_id": "b3f1f2a0-1111-4a2b-8c3d-000000000001",
        "mileage": 1200,
        "fuel_level": 80,
        "location": "Depot A",
        "condition_notes": "Minor scratch on rear bumper"
      }'
```

## Example response

`201 Created`

```json
{
  "id": "d290f1ee-6c54-4b01-90e6-d701748f0851",
  "vin": "1HGCM82633A004352",
  "plate_number": "ABC-1234",
  "make": "Honda",
  "model": "Accord",
  "year": 2023,
  "category_id": "b3f1f2a0-1111-4a2b-8c3d-000000000001",
  "mileage": 1200,
  "fuel_level": 80,
  "location": "Depot A",
  "condition_notes": "Minor scratch on rear bumper",
  "created_at": "2024-03-01T12:00:00.000Z"
}
```

## Error responses

| Status | Condition |
| --- | --- |
| `400 Bad Request` | request body is malformed, or fails Common Validation (returns a `fields` array with `field`/`message` pairs) |
| `401 Unauthorized` | bearer token missing, expired, or otherwise invalid |
| `403 Forbidden` | authenticated user does not have the `service_staff` role |
| `409 Conflict` | `vin` or `plate_number` already assigned to another vehicle |
| `422 Unprocessable Entity` | `category_id` does not reference an existing/active vehicle category |

Example validation error:

```json
{
  "error": "validation failed",
  "fields": [
    { "field": "vin", "message": "is required" }
  ]
}
```
