# Create Maintenance Schedule API

Creates (or configures) a maintenance-schedule rule that determines when a
maintenance work order should be automatically generated for a vehicle
category or an individual vehicle, as defined in the
[Maintenance and Service Scheduling TRD](https://github.com/mshahkap33/ai-sdlc/blob/main/docs/trd/trd-maintenance-service-scheduling.md).

## Endpoint

```
POST /api/v1/maintenance-schedules
```

### Authentication

This endpoint requires a JWT access token, signed with RS256, presented via
the standard OAuth2 bearer-token scheme in the `Authorization` request
header (`Authorization: <scheme> <token>`). The token's `roles` claim must
include `service_staff` or `operations_manager`.

### Request Body

| Field                    | Type    | Required | Description                                                                                     |
| ------------------------ | ------- | -------- | ------------------------------------------------------------------------------------------------ |
| `vehicleCategoryId`      | string  | one of*  | ID of an existing vehicle category this rule applies to                                          |
| `vehicleId`              | string  | one of*  | ID of an existing vehicle this rule applies to                                                   |
| `triggerType`             | string  | yes      | One of `date_interval`, `mileage_interval`, `manufacturer_schedule`, `telematics_condition`       |
| `intervalDays`           | integer | conditional | Required and must be `> 0` when `triggerType = date_interval`                                  |
| `intervalMileage`        | number  | conditional | Required and must be `> 0` when `triggerType = mileage_interval`                                |
| `manufacturerReference`  | string  | conditional | Required when `triggerType = manufacturer_schedule`                                            |
| `telematicsCondition`    | string  | conditional | Required when `triggerType = telematics_condition`                                             |

\* Exactly one of `vehicleCategoryId` or `vehicleId` must be provided.

### Response

`201 Created` with the created schedule rule:

```json
{
  "id": "83b2aa18-7b7e-492d-9d9b-c599d8192ff2",
  "vehicleCategoryId": "11111111-1111-1111-1111-111111111111",
  "triggerType": "date_interval",
  "intervalDays": 180,
  "active": true
}
```

### Error Responses

| Status | Condition                                                                                   |
| ------ | -------------------------------------------------------------------------------------------- |
| `400`  | Missing/invalid request fields (see `fields` in the response body)                             |
| `401`  | Missing, malformed, or invalid/expired bearer token                                            |
| `403`  | Authenticated user does not have the `service_staff` or `operations_manager` role              |
| `404`  | `vehicleCategoryId` or `vehicleId` does not reference an existing record                       |

## Example

```sh
curl -X POST https://api.example.com/api/v1/maintenance-schedules \
  -H "Authorization: <scheme> <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "vehicleCategoryId": "11111111-1111-1111-1111-111111111111",
    "triggerType": "date_interval",
    "intervalDays": 180
  }'
```

Successful response:

```json
{
  "id": "83b2aa18-7b7e-492d-9d9b-c599d8192ff2",
  "vehicleCategoryId": "11111111-1111-1111-1111-111111111111",
  "triggerType": "date_interval",
  "intervalDays": 180,
  "active": true
}
```

### Mileage-interval example for a single vehicle

```sh
curl -X POST https://api.example.com/api/v1/maintenance-schedules \
  -H "Authorization: <scheme> <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "vehicleId": "22222222-2222-2222-2222-222222222222",
    "triggerType": "mileage_interval",
    "intervalMileage": 5000
  }'
```

Not-found error response (`404 Not Found`):

```json
{
  "error": "maintenanceschedule: vehicle_category_id does not reference an existing vehicle category"
}
```

Validation error response (`400 Bad Request`):

```json
{
  "error": "validation failed",
  "fields": {
    "intervalDays": "is required and must be greater than 0 for trigger_type date_interval"
  }
}
```
