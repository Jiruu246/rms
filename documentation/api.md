# Restaurant Management System (RMS) API Documentation

## Table of Contents
- [Customer API]()
- [Menu Items API](#menu-items-api)
- [Categories API](#categories-api)
- [Addons API](#addons-api)
- [Example Payloads](#example-payloads)
- [Status Codes](#status-codes)

---

## Customer API
### Endpoints

| Method   | Endpoint              | Description                                            |
| -------- | --------------------- | ------------------------------------------------------ |
| `POST`   | `/api/users/register` | Register a new user                                    |
| `POST`   | `/api/users/login`    | Log in a user and return a JWT token                   |
| `GET`    | `/api/users/profile`  | Get the current user's profile (requires auth)         |
| `PUT`    | `/api/users/profile`  | Update the current user's profile (support partial update)        |
| `DELETE` | `/api/users/profile`  | Delete user account                                    |


## Menu Items API

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/menu-items` | Create a new menu item |
| `GET` | `/api/menu-items` | Get all menu items (with optional filters) |
| `GET` | `/api/menu-items/{id}` | Get a specific menu item |
| `PUT` | `/api/menu-items/{id}` | Update a menu item (full update) |
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

## Addons API

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/addons` | Create a new addon |
| `GET` | `/api/addons` | Get all addons |
| `GET` | `/api/addons/{id}` | Get a specific addon |
| `PUT` | `/api/addons/{id}` | Update an addon |
| `PATCH` | `/api/addons/{id}` | Partial update an addon |
| `DELETE` | `/api/addons/{id}` | Delete an addon |

---

## Example Payloads

### Create Menu Item
**POST** `/api/menu-items`

```json
{
  "name": "Margherita Pizza",
  "description": "Classic pizza with tomato and mozzarella",
  "price": 12.99,
  "category_id": 1,
  "image_url": "https://example.com/pizza.jpg",
  "is_available": true,
  "addon_ids": [1, 2, 3]
}
```

**Response (201 Created):**
```json
{
  "id": 42,
  "name": "Margherita Pizza",
  "description": "Classic pizza with tomato and mozzarella",
  "price": 12.99,
  "category_id": 1,
  "image_url": "https://example.com/pizza.jpg",
  "is_available": true,
  "addon_ids": [1, 2, 3],
  "created_at": "2025-10-27T10:30:00Z",
  "updated_at": "2025-10-27T10:30:00Z"
}
```

### Create Category
**POST** `/api/categories`

```json
{
  "name": "Pizzas",
  "description": "All our delicious pizzas",
  "display_order": 1,
  "is_active": true
}
```

**Response (201 Created):**
```json
{
  "id": 1,
  "name": "Pizzas",
  "description": "All our delicious pizzas",
  "display_order": 1,
  "is_active": true,
  "created_at": "2025-10-27T10:30:00Z",
  "updated_at": "2025-10-27T10:30:00Z"
}
```

### Create Addon
**POST** `/api/addons`

```json
{
  "name": "Extra Cheese",
  "price": 2.50,
  "is_available": true,
  "applicable_category_ids": [1, 2]
}
```

**Response (201 Created):**
```json
{
  "id": 1,
  "name": "Extra Cheese",
  "price": 2.50,
  "is_available": true,
  "applicable_category_ids": [1, 2],
  "created_at": "2025-10-27T10:30:00Z",
  "updated_at": "2025-10-27T10:30:00Z"
}
```

### Update Menu Item (Partial)
**PATCH** `/api/menu-items/42`

```json
{
  "price": 13.99,
  "is_available": false
}
```

**Response (200 OK):**
```json
{
  "id": 42,
  "name": "Margherita Pizza",
  "description": "Classic pizza with tomato and mozzarella",
  "price": 13.99,
  "category_id": 1,
  "image_url": "https://example.com/pizza.jpg",
  "is_available": false,
  "addon_ids": [1, 2, 3],
  "created_at": "2025-10-27T10:30:00Z",
  "updated_at": "2025-10-27T11:45:00Z"
}
```

### Get All Menu Items
**GET** `/api/menu-items?category_id=1&is_available=true&page=1&limit=10`

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": 42,
      "name": "Margherita Pizza",
      "description": "Classic pizza with tomato and mozzarella",
      "price": 13.99,
      "category_id": 1,
      "image_url": "https://example.com/pizza.jpg",
      "is_available": true,
      "addon_ids": [1, 2, 3],
      "created_at": "2025-10-27T10:30:00Z",
      "updated_at": "2025-10-27T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

---

## Bulk Operations (Optional)

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/menu-items/bulk` | Create multiple menu items |
| `PATCH` | `/api/menu-items/bulk` | Update multiple menu items |
| `DELETE` | `/api/menu-items/bulk` | Delete multiple menu items |

### Bulk Delete Example
**DELETE** `/api/menu-items/bulk`

```json
{
  "ids": [1, 2, 3, 4, 5]
}
```

**Response (200 OK):**
```json
{
  "deleted_count": 5,
  "deleted_ids": [1, 2, 3, 4, 5]
}
```

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

## Versioning

API versioning is included in the URL:
- Current version: `/api/v1/menu-items`
- Future versions: `/api/v2/menu-items`

---

## Rate Limiting

- Rate limit: 1000 requests per hour per API key
- Rate limit headers are included in responses:
  - `X-RateLimit-Limit`: Request limit
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Time when limit resets (Unix timestamp)