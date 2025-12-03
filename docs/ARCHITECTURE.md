# Architecture Document: Byte Courses

## Overview
This document outlines the technical architecture, technology choices, and key design decisions for the Byte Courses platform.

**Last Updated:** December 3, 2025

## Technology Stack

### Frontend
**Choice:** React 19+ with Vite

**Reasoning:**
- Fast development with hot module replacement
- Component-based architecture perfect for reusable UI elements (course cards, module lists)
- Large ecosystem for UI libraries and tools

**Key Libraries:**
- **UI Framework:** Tailwind CSS (utility-first, fast styling)
- **Routing:** React Router v6
- **State Management:** React Context API + useReducer (keeping it simple for MVP)
- **HTTP Client:** Axios
- **Form Handling:** React Hook Form
- **Validation:** Zod (for client-side validation)

**File Structure:**
```
/frontend
    /src
        /components
            /common (Button, Input, Card, ...)
            /course (CourseCard, CourseList, ModuleView)
            /auth   (LoginForm, RegisterForm)
        /pages
            /Home
            /Dashboard
            /CourseView
            /CreateCourse
        /hooks      (useAuth, useCourses)
        /context    (AuthContext)
        /utils      (api client, helpers)
        /assets
        App.jsx
        main.jsx
```

### Backend
**Choice:** Go 1.25+

**Reasoning:**
- Excellent performance and concurrency
- Strong typing prevents bugs
- Simple deployment (single binary)
- Great standard library
- Fast compilation for quick iteration

**Key Packages:**
- **Router:** gorilla/mux or chi
- **Database:** lib/pq (PostgreSQL driver) or pgx
- **Authentication:** golang-jwt/jwt for JWT tokens
- **Password Hashing:** golang.org/x/crypto/bcrypt
- **Validation:** go-playground/validator
- **Environment Variables:** godotenv
- **CORS:** gorilla/handlers or rs/cors

**File Structure:**
```
/backend
    /cmd
        /server
            main.go
    /internal
        /handlers (HTTP handlers)
        /models (data models)
        /middleware (auth, logging, CORS)
        /database (DB connection, queries)
        /services (business logic)
        /auth (JWT generation/validation)
    /migrations (SQL migration files)
    /config (config loading)
    go.mod
    go.sum
```

### Database
**Choice:** PostgreSQL 15+

**Reasoning:**
- Relational data (courses have modules, modules have assignments)
- ACID compliance important for grades/submissions
- Strong community and documentation
- Native JSON support for flexible content storage

### Authentication Strategy

**Choice:** JSON Web Tokens

**Flow:**
1. User submits email/password to `/api/auth/login`
2. Backend validates credentials, hashes password with bcrypt
3. If valid, backend generates JWT with user ID and role
4. Token sent to frontend in response body
5. Frontend stores token in localStorage
6. Frontend includes token in Authorization header for all requests: `Bearer <token>`
7. Backend middleware validates token on protected routes
8. Token expires after 7 days (refresh token for later phase)

**Token Payload:**
```json
{
  "user_id": 123,
  "email": "email@example.com",
  "role": "student",
  "exp": 1234567890
}
```

## API Design
**Style:** RESTful API\
**Base URL:** `http://localhost:8080/api`

**Authentication Endpoints:**
```
POST   /api/auth/register     - Create new account
POST   /api/auth/login        - Login and get JWT
GET    /api/auth/me           - Get current user info (requires auth)
```

**Course Endpoints:**
```
GET    /api/courses                     - List all published courses
GET    /api/courses/:id                 - Get course details
POST   /api/courses                     - Create course (instructor only)
PUT    /api/courses/:id                 - Update course (instructor only)
DELETE /api/courses/:id                 - Delete course (instructor only)
POST   /api/courses/:id/enroll          - Enroll in course (requires auth)
GET    /api/courses/:id/modules         - Get course modules
POST   /api/courses/:id/modules         - Add module (instructor only)
```

**Module Endpoints:**
```
GET    /api/modules/:id                 - Get module details
PUT    /api/modules/:id                 - Update module (instructor only)
DELETE /api/modules/:id                 - Delete module (instructor only)
POST   /api/modules/:id/complete        - Mark module complete (student)
```

**Assignment Endpoints:**
```
GET    /api/assignments/:id             - Get assignment details
POST   /api/assignments/:id/submit      - Submit assignment (student)
GET    /api/assignments/:id/submissions - Get all submissions (instructor)
PUT    /api/submissions/:id/grade       - Grade submission (instructor)
```

**User Endpoints:**
```
GET    /api/users/me/courses            - Get enrolled courses
GET    /api/users/me/progress/:courseId - Get progress in a course
```

**Response Format:**
```json
{
  "success": true,
  "data": { ... },
  "message": "Optional message"
}
```

**Error Format:**
```json
{
  "success": false,
  "error": "Error message here",
  "code": "ERROR_CODE"
}
```

## Architecture Patterns

### Backend Patterns

**Layered Architecture:**
- **Handlers:** HTTP request/response handling
- **Services:** Business logic
- **Models:** Data structures
- **Database:** Data access layer

**Middleware Chain:**
```
Request → CORS → Logging → Auth (if needed) → Handler → Response
```

**Example Handler Structure:**
```go
func (h *CourseHandler) CreateCourse(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request body
    // 2. Validate input
    // 3. Check authorization (is user an instructor?)
    // 4. Call service layer
    // 5. Return JSON response
}
```

### Frontend Patterns

**Component Composition:**
- Small, reusable components
- Container components handle data fetching
- Presentational components just display

**State Management:**
- Global state: AuthContext (user, login, logout)
- Local state: useState for component-specific state
- Server state: Fetch on mount, store in state

**API Client Pattern:**
```javascript
// utils/api.js
const api = {
    get: (url) => axios.get(url, { headers: { Authorization: `Bearer ${token}` }}),
    post: (url, data) => axios.post(url, data, { headers: {...} }),
    // ...
}
```

## Security Considerations

- [ ] All passwords hashed with bcrypt (cost factor 12+)
- [ ] JWT tokens with 7-day expiry
- [ ] Input validation on both frontend (React Hook Form + Zod) and backend (validator package)
- [ ] File upload restrictions:
    - Types: PDF, DOCX, TXT, PNG, JPG only
    - Size limit: 5MB
    - Filename sanitization
- [ ] CORS configured to allow only frontend domain
- [ ] Rate limiting on auth endpoints (5 requests per minute)
- [ ] HTTPS in production (handled by hosting platform)
- [ ] Environment variables for secrets
- [ ] SQL parameterization (pgx handles this automatically)
- [ ] XSS prevention (React escapes by default, sanitize any dangerouslySetInnerHTML)
- [ ] CSRF protection (not needed for JWT-based auth)

## Performance Considerations

**Backend:**
- Connection pooling for database (max 25 connections)
- Gzip compression for responses
- Graceful shutdown handling
- Efficient queries with proper indexes

**Frontend:**
- Code splitting with React.lazy()
- Lazy loading course modules (don't load all at once)
- Image optimization (use Cloudinary transformations)
- Pagination for course lists (20 per page)
- Debounce search inputs

**Database:**
- Indexes on foreign keys (already in schema)
- EXPLAIN ANALYZE for slow queries
- Avoid N+1 queries (use JOINs appropriately)

**Environment Variables:**

Backend (.env):
```bash
DATABASE_URL=postgres://user:pass@localhost:5432/clublearn?sslmode=disable
JWT_SECRET=your-secret-key-here-make-it-long-and-random
PORT=8080
FRONTEND_URL=http://localhost:5173
AWS_ACCESS_KEY_ID=your-key
AWS_SECRET_ACCESS_KEY=your-secret
AWS_REGION=us-east-1
AWS_BUCKET_NAME=club-learning-uploads
```

Frontend (.env):
```bash
VITE_API_URL=http://localhost:8080/api
```

## Database Migrations

Using simple SQL files in `/migrations` folder:
- `001_initial.sql` - Create all tables
- `002_add_indexes.sql` - Add performance indexes
- etc.

Run manually for MVP, can add migrate tool later (golang-migrate/migrate)

## Decisions Log

**Dec 3, 2024:** Chose PostgreSQL over MongoDB because the data is highly relational (courses → modules → assignments → submissions). We need strong consistency for grades and foreign key constraints.

**Dec 3, 2024:** Chose JWT over session-based auth because it's stateless and works well with separate frontend/backend deployments. Sessions would require sticky sessions or shared session store.

**Dec 3, 2024:** Using gorilla/mux for routing because it's simple, widely used, and has good middleware support. Chi is also a good alternative.

**Dec 3, 2024:** Storing files in S3 rather than database because it's more scalable and cheaper. Database should only store URLs.

**Dec 3, 2024:** Using React Context instead of Redux because the app state is relatively simple. Can migrate to Redux/Zustand later if needed.
