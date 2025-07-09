# Course Management API Documentation

This document provides comprehensive documentation for the Course Management API v1, designed for mobile applications and external integrations.

## Base URL

```
https://api.course-management.com/api/v1
```

## Authentication

The API uses JWT (JSON Web Tokens) for authentication. Tokens are obtained through Google OAuth flow.

### Token Types

- **Access Token**: Short-lived (1 hour), used for API requests
- **Refresh Token**: Long-lived (7 days), used to obtain new access tokens

### Headers

```
Authorization: Bearer <access_token>
Content-Type: application/json
```

## Standard Response Format

All API responses follow this format:

```json
{
  "success": true,
  "data": {...},
  "meta": {...},
  "error": null,
  "timestamp": 1640995200
}
```

### Error Response Format

```json
{
  "success": false,
  "data": null,
  "error": {
    "error": "validation_error",
    "message": "Request validation failed",
    "code": "VAL_001",
    "details": {
      "name": "Name is required"
    }
  },
  "timestamp": 1640995200
}
```

## Authentication Endpoints

### POST /auth/google/verify

Verify Google OAuth token and get JWT tokens.

**Request:**
```json
{
  "id_token": "google_id_token_here",
  "access_token": "google_access_token_here"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "tokens": {
      "access_token": "jwt_access_token",
      "refresh_token": "jwt_refresh_token",
      "token_type": "Bearer",
      "expires_in": 3600
    },
    "user": {
      "id": 123,
      "google_id": "google123",
      "email": "user@example.com",
      "name": "John Doe",
      "display_name": "Johnny",
      "picture": "https://...",
      "handicap": 18.5,
      "created_at": 1640995200,
      "updated_at": 1640995200
    }
  }
}
```

### POST /auth/refresh

Refresh access token using refresh token.

**Request:**
```json
{
  "refresh_token": "jwt_refresh_token"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "tokens": {
      "access_token": "new_jwt_access_token",
      "refresh_token": "new_jwt_refresh_token",
      "token_type": "Bearer",
      "expires_in": 3600
    }
  }
}
```

### GET /auth/status

Get current authentication status.

**Headers:** `Authorization: Bearer <token>` (optional)

**Response:**
```json
{
  "success": true,
  "data": {
    "authenticated": true,
    "user": {
      "id": 123,
      "email": "user@example.com",
      "name": "John Doe"
    }
  }
}
```

### POST /auth/logout

Logout and invalidate current token.

**Headers:** `Authorization: Bearer <token>` (required)

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Successfully logged out"
  }
}
```

## User Endpoints

### GET /user/profile

Get authenticated user's profile.

**Headers:** `Authorization: Bearer <token>` (required)

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 123,
    "google_id": "google123",
    "email": "user@example.com",
    "name": "John Doe",
    "display_name": "Johnny",
    "picture": "https://...",
    "handicap": 18.5,
    "created_at": 1640995200,
    "updated_at": 1640995200
  }
}
```

### PUT /user/profile

Update user's profile.

**Headers:** `Authorization: Bearer <token>` (required)

**Request:**
```json
{
  "display_name": "Johnny Golf"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 123,
    "display_name": "Johnny Golf",
    ...
  }
}
```

### PUT /user/handicap

Update user's handicap.

**Headers:** `Authorization: Bearer <token>` (required)

**Request:**
```json
{
  "handicap": 16.2
}
```

### GET /user/scores

Get user's score history.

**Headers:** `Authorization: Bearer <token>` (required)

**Query Parameters:**
- `page` (int, default: 1): Page number
- `per_page` (int, default: 20, max: 100): Items per page
- `course_id` (int, optional): Filter by course

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "course_id": 456,
      "course_name": "Pebble Beach Golf Links",
      "score": 85,
      "handicap": 18.5,
      "played_at": 1640995200,
      "notes": "Great day on the course!",
      "weather": "Sunny",
      "conditions": "Perfect",
      "created_at": 1640995200
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 50,
    "total_pages": 3
  }
}
```

### POST /user/scores

Add a new score.

**Headers:** `Authorization: Bearer <token>` (required)

**Request:**
```json
{
  "course_id": 456,
  "score": 85,
  "handicap": 18.5,
  "played_at": 1640995200,
  "notes": "Great round!",
  "weather": "Sunny",
  "conditions": "Perfect"
}
```

### DELETE /user/scores/:scoreId

Delete a user's score.

**Headers:** `Authorization: Bearer <token>` (required)

**Response:** 204 No Content

### GET /user/stats

Get user statistics.

**Headers:** `Authorization: Bearer <token>` (required)

**Response:**
```json
{
  "success": true,
  "data": {
    "total_rounds": 25,
    "average_score": 87.2,
    "best_score": 79,
    "current_handicap": 18.5,
    "courses_played": 12,
    "recent_trend": "improving"
  }
}
```

## Course Endpoints

### GET /courses

Get all courses.

**Headers:** `Authorization: Bearer <token>` (optional)

**Query Parameters:**
- `page` (int, default: 1): Page number
- `per_page` (int, default: 20, max: 100): Items per page
- `q` (string): Search query
- `lat` (float): Latitude for distance calculation
- `lng` (float): Longitude for distance calculation
- `radius` (float): Radius in kilometers
- `min_rating` (float): Minimum rating filter
- `max_rating` (float): Maximum rating filter
- `sort_by` (string): Sort field (name, rating, distance, created_at)
- `sort_order` (string): Sort order (asc, desc)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "Pebble Beach Golf Links",
      "address": "1700 17 Mile Dr, Pebble Beach, CA 93953",
      "description": "World-famous golf course",
      "phone": "+1-831-624-3811",
      "website": "https://pebblebeach.com",
      "latitude": 36.5674,
      "longitude": -121.9450,
      "holes": [
        {
          "number": 1,
          "par": 4,
          "yardage": 373,
          "description": "First hole description"
        }
      ],
      "created_by": 123,
      "created_at": 1640995200,
      "updated_at": 1640995200,
      "can_edit": true,
      "user_review": {
        "overall_rating": 9,
        "last_played": 1640995200,
        "times_played": 3,
        "best_score": 85,
        "average_score": 89.3
      },
      "stats": {
        "total_reviews": 150,
        "average_rating": 8.5,
        "total_rounds": 2500,
        "average_score": 85.2,
        "difficulty_level": "Hard"
      }
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

### GET /courses/:id

Get specific course.

**Headers:** `Authorization: Bearer <token>` (optional)

### POST /courses

Create new course.

**Headers:** `Authorization: Bearer <token>` (required)

**Request:**
```json
{
  "name": "New Golf Course",
  "address": "123 Golf Course Rd, City, State 12345",
  "description": "Beautiful new course",
  "phone": "+1-555-123-4567",
  "website": "https://newgolfcourse.com",
  "holes": [
    {
      "number": 1,
      "par": 4,
      "yardage": 400,
      "description": "Challenging opening hole"
    }
  ]
}
```

### PUT /courses/:id

Update course (owner only).

**Headers:** `Authorization: Bearer <token>` (required)

### DELETE /courses/:id

Delete course (owner only).

**Headers:** `Authorization: Bearer <token>` (required)

### GET /courses/search

Advanced course search.

**Query Parameters:** Same as GET /courses

### GET /courses/nearby

Get courses near location.

**Query Parameters:**
- `lat` (float, required): Latitude
- `lng` (float, required): Longitude  
- `radius` (float, default: 10): Radius in kilometers
- `page` (int, default: 1): Page number
- `per_page` (int, default: 20): Items per page

## Review Endpoints

### GET /courses/:courseId/reviews

Get reviews for a course.

**Headers:** `Authorization: Bearer <token>` (optional)

**Query Parameters:**
- `page` (int, default: 1): Page number
- `per_page` (int, default: 20): Items per page
- `sort_by` (string): Sort field (rating, date, helpful)
- `sort_order` (string): Sort order (asc, desc)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "course_id": 456,
      "course_name": "Pebble Beach Golf Links",
      "user_id": 123,
      "user_name": "John Doe",
      "user_display_name": "Johnny",
      "overall_rating": 9,
      "review_text": "Amazing course with beautiful views!",
      "price": "$$$$",
      "handicap_difficulty": 8,
      "hazard_difficulty": 7,
      "merch": "Great selection",
      "condition": "Excellent",
      "enjoyment_rating": 10,
      "vibe": "Upscale",
      "range": "Excellent",
      "amenities": "Full service",
      "food": "Outstanding",
      "atmosphere": "Premium",
      "value": 6,
      "maintenance": 10,
      "pace": 7,
      "staff": 9,
      "created_at": 1640995200,
      "updated_at": 1640995200,
      "can_edit": false,
      "helpful_count": 15,
      "is_helpful": true
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 50,
    "total_pages": 3
  }
}
```

### GET /courses/:courseId/reviews/summary

Get review summary for a course.

**Response:**
```json
{
  "success": true,
  "data": {
    "course_id": 456,
    "total_reviews": 150,
    "average_rating": 8.5,
    "rating_breakdown": {
      "1": 2,
      "2": 1,
      "3": 5,
      "4": 10,
      "5": 15,
      "6": 20,
      "7": 25,
      "8": 30,
      "9": 25,
      "10": 17
    },
    "categories": {
      "handicap_difficulty": 7.8,
      "hazard_difficulty": 6.9,
      "enjoyment_rating": 8.7,
      "value": 6.2,
      "maintenance": 9.1,
      "pace": 7.3,
      "staff": 8.4
    },
    "recent_reviews": [...]
  }
}
```

### POST /reviews

Create review.

**Headers:** `Authorization: Bearer <token>` (required)

**Request:**
```json
{
  "course_id": 456,
  "overall_rating": 9,
  "review_text": "Amazing course!",
  "price": "$$$$",
  "handicap_difficulty": 8,
  "hazard_difficulty": 7,
  "merch": "Great selection",
  "condition": "Excellent",
  "enjoyment_rating": 10,
  "vibe": "Upscale",
  "range": "Excellent",
  "amenities": "Full service",
  "food": "Outstanding",
  "atmosphere": "Premium",
  "value": 6,
  "maintenance": 10,
  "pace": 7,
  "staff": 9
}
```

### PUT /reviews/:id

Update review (author only).

**Headers:** `Authorization: Bearer <token>` (required)

### DELETE /reviews/:id

Delete review (author only).

**Headers:** `Authorization: Bearer <token>` (required)

### GET /reviews/user

Get user's reviews.

**Headers:** `Authorization: Bearer <token>` (required)

### POST /reviews/:id/helpful

Mark review as helpful/not helpful.

**Headers:** `Authorization: Bearer <token>` (required)

**Request:**
```json
{
  "helpful": true
}
```

## Map Endpoints

### GET /map/courses

Get courses optimized for map display.

**Headers:** `Authorization: Bearer <token>` (optional)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "Pebble Beach Golf Links",
      "address": "1700 17 Mile Dr, Pebble Beach, CA 93953",
      "latitude": 36.5674,
      "longitude": -121.9450,
      "overall_rating": 8.5,
      "total_reviews": 150,
      "can_edit": false,
      "distance": 5.2
    }
  ]
}
```

### GET /map/courses/bounds

Get courses within geographic bounds.

**Query Parameters:**
- `north_lat` (float, required): North latitude
- `south_lat` (float, required): South latitude
- `east_lng` (float, required): East longitude
- `west_lng` (float, required): West longitude
- `min_rating` (float, optional): Minimum rating
- `max_rating` (float, optional): Maximum rating

### GET /map/courses/clusters

Get clustered course data for map display.

**Query Parameters:** Same as `/map/courses/bounds` plus:
- `zoom` (int, default: 10): Map zoom level
- `max_cluster_size` (int, default: 50): Maximum cluster size

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "latitude": 36.5674,
      "longitude": -121.9450,
      "course_count": 5,
      "zoom_level": 10,
      "courses": [...]
    }
  ]
}
```

### POST /map/geocode

Geocode address to coordinates.

**Request:**
```json
{
  "address": "1700 17 Mile Dr, Pebble Beach, CA 93953"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "address": "1700 17 Mile Dr, Pebble Beach, CA 93953",
    "formatted_address": "1700 17-Mile Dr, Pebble Beach, CA 93953, USA",
    "latitude": 36.5674,
    "longitude": -121.9450,
    "confidence": 0.95,
    "components": {
      "street_number": "1700",
      "route": "17-Mile Dr",
      "city": "Pebble Beach",
      "state": "CA",
      "country": "USA",
      "postal_code": "93953"
    }
  }
}
```

### GET /map/reverse-geocode

Convert coordinates to address.

**Query Parameters:**
- `lat` (float, required): Latitude
- `lng` (float, required): Longitude

### GET /map/route

Get route between two points.

**Query Parameters:**
- `from_course_id` (int): From course ID
- `to_course_id` (int): To course ID
- `from_lat` (float): From latitude
- `from_lng` (float): From longitude
- `to_lat` (float): To latitude
- `to_lng` (float): To longitude

**Response:**
```json
{
  "success": true,
  "data": {
    "distance": 25.6,
    "duration": 1800,
    "geometry": "encoded_polyline_string",
    "steps": [
      {
        "distance": 150,
        "duration": 30,
        "instruction": "Head north on Main St",
        "geometry": "step_polyline",
        "start_lat": 36.5674,
        "start_lng": -121.9450,
        "end_lat": 36.5680,
        "end_lng": -121.9450
      }
    ],
    "waypoints": [
      {
        "latitude": 36.5674,
        "longitude": -121.9450,
        "name": "Starting Point"
      }
    ]
  }
}
```

### GET /map/statistics

Get map-related statistics.

**Response:**
```json
{
  "success": true,
  "data": {
    "total_courses": 1250,
    "courses_with_location": 1100,
    "location_coverage": 88.0,
    "bounding_box": {
      "north_lat": 49.3457,
      "south_lat": 24.3963,
      "east_lng": -66.9346,
      "west_lng": -124.7859
    },
    "popular_regions": [
      {
        "name": "California",
        "course_count": 250,
        "average_rating": 7.8,
        "center_lat": 36.7783,
        "center_lng": -119.4179
      }
    ]
  }
}
```

## Utility Endpoints

### GET /health

Health check endpoint.

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "version": "1.0.0",
    "environment": "production",
    "services": {
      "database": "connected",
      "auth": "operational",
      "geocoding": "operational"
    },
    "timestamp": 1640995200
  }
}
```

## Error Codes

| Code | Error | Description |
|------|--------|-------------|
| REQ_001 | bad_request | Invalid request format or parameters |
| VAL_001 | validation_error | Request validation failed |
| AUTH_001 | unauthorized | Authentication required |
| AUTH_002 | forbidden | Insufficient permissions |
| RES_001 | not_found | Resource not found |
| RES_002 | conflict | Resource conflict |
| SYS_001 | internal_server_error | Internal server error |
| SYS_002 | service_unavailable | Service temporarily unavailable |
| RATE_001 | rate_limit_exceeded | Too many requests |

## Rate Limiting

- **Default**: 120 requests per minute per IP/user
- **Headers**: 
  - `X-RateLimit-Limit`: Requests per minute
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Reset time

## Pagination

All list endpoints support pagination:

**Query Parameters:**
- `page` (int, default: 1, min: 1): Page number
- `per_page` (int, default: 20, min: 1, max: 100): Items per page

**Response Meta:**
```json
{
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

## Security Features

- **JWT Authentication**: Secure token-based authentication
- **HTTPS Only**: All API calls must use HTTPS in production
- **CORS**: Configured for mobile app origins
- **Rate Limiting**: Prevents abuse
- **Input Validation**: All inputs are validated
- **SQL Injection Protection**: Parameterized queries
- **XSS Protection**: Proper output encoding
- **CSRF Protection**: CSRF tokens for state-changing operations

## SDK and Client Libraries

Coming soon:
- iOS Swift SDK
- Android Kotlin SDK
- JavaScript/TypeScript SDK
- React Native SDK

## Support

For API support, please contact:
- Email: api-support@course-management.com
- Documentation: https://docs.course-management.com
- GitHub Issues: https://github.com/course-management/api/issues