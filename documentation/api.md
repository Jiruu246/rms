# Restaurant Management System (RMS) API Documentation

## Table of Contents
- [Auth API](#auth-api)
- [User & Staff Management API](#user--staff-management)
- [Restaurant Management API](#restaurant-management)
- [Menu Items API](#menu-items-api)
- [Categories API](#categories-api)
- [modifiers API](#modifiers-api)
- [Table API](#table-api)
- [Order API](#order-api)
- [Payment API](#payment-api)
- [Reservation API](#reservation-api)
- [Customer API](#customer-api)
- [Status Codes](#status-codes)

---

## Auth API
This endpoint is public

```
POST   /api/v1/auth/register          - Register new owner account
POST   /api/v1/auth/login             - Login (returns JWT)
POST   /api/v1/auth/refresh           - Refresh access token
POST   /api/v1/auth/logout            - Logout (invalidate token)
POST   /api/v1/auth/forgot-password   - Request password reset
POST   /api/v1/auth/reset-password    - Reset password with token
```

---

## User & Staff Management

```
GET    /api/v1/users/me                    [JWT: All]
PUT    /api/v1/users/me                    [JWT: All]
PUT    /api/v1/users/me/password           [JWT: All]

GET    /api/v1/staff                       [JWT: Owner, Manager]
POST   /api/v1/staff                       [JWT: Owner, Manager]
GET    /api/v1/staff/:id                   [JWT: Owner, Manager]
PUT    /api/v1/staff/:id                   [JWT: Owner, Manager]
DELETE /api/v1/staff/:id                   [JWT: Owner]
PUT    /api/v1/staff/:id/role              [JWT: Owner]
PUT    /api/v1/staff/:id/status            [JWT: Owner, Manager]
```

---

## Restaurant Management

```
GET    /api/v1/restaurants                 [JWT: Owner]
POST   /api/v1/restaurants                 [JWT: Owner]
GET    /api/v1/restaurants/:id             [JWT: Owner, Staff of restaurant]
PUT    /api/v1/restaurants/:id             [JWT: Owner]
DELETE /api/v1/restaurants/:id             [JWT: Owner]
PUT    /api/v1/restaurants/:id/settings    [JWT: Owner, Manager]
```

---

## Menu Items API

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/menu-items` | Create a new menu item |
| `GET` | `/api/menu-items` | Get all menu items for a restaurantId |
| `GET` | `/api/menu-items/{id}` | Get a specific menu item |
| `PATCH` | `/api/menu-items/{id}` | Partial update a menu item |
| `DELETE` | `/api/menu-items/{id}` | Delete a menu item |

### Query Parameters for GET /api/menu-items

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `category_id` | uuid | Filter by category | `?category_id=123` |
| `is_available` | boolean | Filter by availability | `?is_available=true` |
| `search` | string | Search by name/description | `?search=pizza` |
| `sort_by` | string | Sort field | `?sort_by=price` |
| `order` | string | Sort order (asc/desc) | `?order=asc` |
| `page` | integer | Page number for pagination | `?page=1` |
| `limit` | integer | Items per page | `?limit=20` |

---

## Categories API

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/categories` | Create a new category |
| `GET` | `/api/categories` | Get all categories |
| `GET` | `/api/categories/{id}` | Get a specific category |
| `PUT` | `/api/categories/{id}` | Update a category |
| `PATCH` | `/api/categories/{id}` | Partial update a category |
| `DELETE` | `/api/categories/{id}` | Delete a category |
| `GET` | `/api/categories/{id}/items` | Get all menu items in a category |

---

## Modifiers API

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/modifiers` | Create a new modifier |
| `GET` | `/api/modifiers` | Get all modifiers |
| `GET` | `/api/modifiers/{id}` | Get a specific modifier |
| `PATCH` | `/api/modifiers/{id}` | Partial update a modifier |
| `DELETE` | `/api/modifiers/{id}` | Delete a modifier |

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/modifiers/options` | Create a new modifier option |
| `GET` | `/api/modifiers/options` | Get all modifiers |
| `GET` | `/api/modifiers/options/{id}` | Get a specific modifier |
| `PATCH` | `/api/modifiers/options/{id}` | Partial update a modifier |
| `DELETE` | `/api/modifiers/options/{id}` | Delete a modifier |

---

## Table API

```
GET    /api/v1/restaurants/:id/tables      [JWT: Owner, Manager, Waiter]
POST   /api/v1/tables                      [JWT: Owner, Manager]
PUT    /api/v1/tables/:id                  [JWT: Owner, Manager]
DELETE /api/v1/tables/:id                  [JWT: Owner, Manager]
PUT    /api/v1/tables/:id/status           [JWT: Owner, Manager, Waiter]
GET    /api/v1/tables/:id/current-order    [JWT: Owner, Manager, Waiter]
```

---

## Order API

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/orders` | Create a new order |
| `GET` | `/api/orders` | Get all order for a restaurant |
| `GET` | `/api/orders/{id}` | Get a specific order |
| `PATCH` | `/api/orders/{id}` | Partial update an order |
| `DELETE` | `/api/orders/{id}` | Delete a modifier |

---

## Payment API

```
POST   /api/v1/orders/:id/checkout         [JWT: Cashier, Waiter, Manager]
POST   /api/v1/orders/:id/payment          [JWT: Cashier, Manager]
POST   /api/v1/orders/:id/split            [JWT: Cashier, Waiter, Manager]
POST   /api/v1/orders/:id/refund           [JWT: Manager, Owner]
GET    /api/v1/orders/:id/invoice          [JWT: Cashier, Manager, Owner]
```

---

## Reservation API

```
POST   /api/v1/reservations                [Public or JWT: Customer]
GET    /api/v1/reservations                [JWT: Owner, Manager, Waiter]
GET    /api/v1/reservations/:id            [JWT: All staff, or customer who created]
PUT    /api/v1/reservations/:id            [JWT: Manager, Waiter]
DELETE /api/v1/reservations/:id            [JWT: Manager, Owner]
PUT    /api/v1/reservations/:id/status     [JWT: Manager, Waiter]
GET    /api/v1/reservations/today          [JWT: Manager, Waiter]
```

---

## Customer API

```
GET    /api/v1/customers                   [JWT: Owner, Manager]
GET    /api/v1/customers/:id               [JWT: Owner, Manager]
POST   /api/v1/customers                   [JWT: Waiter, Manager]
PUT    /api/v1/customers/:id               [JWT: Manager]
GET    /api/v1/customers/:id/orders        [JWT: Owner, Manager]
POST   /api/v1/customers/:id/feedback      [JWT: Waiter, Manager]
```

---

## Bulk Operations (Optional)

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/menu-items/bulk` | Create multiple menu items |
| `PATCH` | `/api/menu-items/bulk` | Update multiple menu items |
| `DELETE` | `/api/menu-items/bulk` | Delete multiple menu items |

---

## Status Codes

| Code | Description | Usage |
|------|-------------|-------|
| `200` | OK | Successful GET, PUT, PATCH requests |
| `201` | Created | Successful POST requests |
| `204` | No Content | Successful DELETE requests |
| `400` | Bad Request | Invalid request payload or parameters |
| `401` | Unauthorized | Authentication required |
| `403` | Forbidden | Insufficient permissions |
| `404` | Not Found | Resource not found |
| `409` | Conflict | Duplicate resource (e.g., duplicate name) |
| `422` | Unprocessable Entity | Validation errors |
| `500` | Internal Server Error | Server-side error |

---

## Error Response Format

All error responses follow this structure:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {
        "field": "price",
        "message": "Price must be a positive number"
      }
    ]
  }
}
```

---

## Authentication

All API endpoints require authentication using Bearer tokens:

```
Authorization: Bearer <your_token_here>
```

---

## Rate Limiting

- Rate limit: 1000 requests per hour per API key
- Rate limit headers are included in responses:
  - `X-RateLimit-Limit`: Request limit
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Time when limit resets (Unix timestamp)