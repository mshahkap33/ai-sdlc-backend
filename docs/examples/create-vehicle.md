# Create Vehicle API

Creates a new vehicle master data record (Vehicle Onboarding), as defined in
the [Vehicle Onboarding TRD](https://github.com/mshahkap33/ai-sdlc/blob/main/docs/trd/trd-vehicle-onboarding.md).

## Endpoint

```
POST /api/v1/vehicles
```

### Authentication

This endpoint requires a JWT access token, signed with RS256, presented via
the standard OAuth2 bearer-token scheme in the `Authorization` request
header (`Authorization: <scheme> <token>`). The token's `roles` claim must
include `service_staff`.

### Request Body

| Field             | Type   | Required | Description                                             |
| ----------------- | ------ | -------- | --------------------------------------------------------|
| `vin`             | string | yes      | 17-character VIN (letters excluding I, O, Q, and digits) |
| `plate_number`    | string | yes      | Alphanumeric, optional hyphen/space separators          |
| `make`            | string | yes      | Vehicle manufacturer                                    |
| `model`           | string | yes      | Vehicle model                                           |
| `year`            | int    | yes      | 4-digit model year, not greater than current year + 1   |
| `category_id`     | string | yes      | ID of an existing, active vehicle category               |
| `mileage`         | number | yes      | Current odometer reading, >= 0                          |
| `fuel_level`      | number | yes      | Fuel level percentage, between 0 and 100                |
| `location`        | string | yes      | Current location identifier or description              |
| `condition_notes` | string | no       | Free-form condition notes                                |

### Response

`201 Created` with the created vehicle record:

```json
{
  "id": "83b2aa18-7b7e-492d-9d9b-c599d8192ff2",
  "vin": "1HGCM82633A004352",
  "plate_number": "ABC-1234",
  "make": "Honda",
  "model": "Accord",
  "year": 2023,
  "category_id": "11111111-1111-1111-1111-111111111111",
  "mileage": 1200,
  "fuel_level": 80,
  "location": "Depot A",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Error Responses

| Status | Condition                                                        |
| ------ | ----------------------------------------------------------------- |
| `400`  | Missing/invalid request fields (see `fields` in the response body) |
| `401`  | Missing, malformed, or invalid/expired bearer token                |
| `403`  | Authenticated user does not have the `service_staff` role         |
| `409`  | `vin` or `plate_number` already assigned to another vehicle       |
| `422`  | `category_id` does not reference an existing, active category     |

## Example

```sh
curl -X POST https://api.example.com/api/v1/vehicles \
  -H "Authorization: <scheme> <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "vin": "1HGCM82633A004352",
    "plate_number": "ABC-1234",
    "make": "Honda",
    "model": "Accord",
    "year": 2023,
    "category_id": "11111111-1111-1111-1111-111111111111",
    "mileage": 1200,
    "fuel_level": 80,
    "location": "Depot A",
    "condition_notes": "Minor scratch on rear bumper"
  }'
```

Successful response:

```json
{
  "id": "83b2aa18-7b7e-492d-9d9b-c599d8192ff2",
  "vin": "1HGCM82633A004352",
  "plate_number": "ABC-1234",
  "make": "Honda",
  "model": "Accord",
  "year": 2023,
  "category_id": "11111111-1111-1111-1111-111111111111",
  "mileage": 1200,
  "fuel_level": 80,
  "location": "Depot A",
  "condition_notes": "Minor scratch on rear bumper",
  "created_at": "2024-01-01T00:00:00Z"
}
```

Duplicate VIN error response (`409 Conflict`):

```json
{
  "error": "vehicle: vin already assigned to another vehicle"
}
```

Validation error response (`400 Bad Request`):

```json
{
  "error": "validation failed",
  "fields": {
    "vin": "must be a 17-character VIN using only letters (excluding I, O, Q) and digits"
  }
}
```
