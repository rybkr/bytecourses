# The Byte Course Project - Knowledge Base

## Project Overview

**The Byte Course Project** is an initiative to encourage students to host short courses in specialized topics. The platform serves two primary purposes:
- **For Learners**: Provides a more relatable and nimble type of learning experience
- **For Instructors**: Helps students build their resumes by teaching specialized topics

## Technology Stack

### Backend
- **Language**: Go (1.25.1)
- **Web Framework**: Chi router (v5.2.3)
- **Authentication**: Custom session-based auth with bcrypt password hashing
- **Storage**: 
  - Current: In-memory store (for development)
  - Planned: SQL database backend (not yet implemented)

### Frontend
- **Approach**: Vanilla JavaScript, HTML, CSS
- **Templating**: Go HTML templates with layout inheritance
- **Structure**: Server-side rendered pages with client-side JavaScript for interactivity

### Testing
- **E2E Tests**: Python-based end-to-end tests (pytest)

## Key Features

### 1. Course Submission (Proposal System)
- Users can apply to be instructors by submitting course proposals
- Proposals require admin approval before becoming courses
- Proposal workflow states: draft → submitted → (approved/rejected/changes_requested/withdrawn)
- Users can create, edit, view, and manage their own proposals

### 2. Course Scanning (Planned)
- Users can view available courses
- Sort courses by relevant features (not yet implemented)

### 3. Course Viewing and Completion (Planned)
- Users can access course content
- Users can complete assignments (not yet implemented)

## Current Architecture

### Project Structure
```
bytecourses/
├── cmd/server/          # Application entry point
├── internal/
│   ├── app/            # Application initialization and routing
│   ├── auth/           # Authentication and session management
│   ├── domain/         # Domain models and types
│   ├── http/           # HTTP handlers and middleware
│   └── store/          # Data persistence layer (interfaces and implementations)
├── web/                # Frontend assets
│   ├── static/         # CSS, JavaScript files
│   └── templates/      # HTML templates
├── test/e2e/           # End-to-end tests
└── docs/               # Documentation
```

### Domain Models

#### User
- Roles: `student`, `instructor`, `admin`
- Fields: ID, Email, Name, PasswordHash, Role, CreatedAt

#### Proposal (Course Proposal)
- Status: `draft`, `submitted`, `withdrawn`, `approved`, `rejected`, `changes_requested`
- Fields: ID, Title, Summary, AuthorID, TargetAudience, LearningObjectives, Outline, AssumedPrerequisites, ReviewNotes, ReviewerID, CreatedAt, UpdatedAt, Status

### Current Implementation Status

**Implemented:**
- User authentication (register, login, logout)
- Session management (in-memory sessions)
- Proposal CRUD operations
- Proposal workflow actions (submit, approve, reject, etc.)
- Admin user seeding
- Basic page rendering with layout templates

**Not Yet Implemented:**
- SQL database backend
- Course entity (currently only proposals exist)
- Course browsing and filtering
- Course content delivery
- Assignment system
- Course completion tracking

## Coding Standards and Philosophy

### Core Principles
1. **Simplicity**: Prefer straightforward solutions over complex abstractions
2. **Locality**: Keep related code close together; minimize cross-file dependencies
3. **Cleanliness**: Write clear, self-documenting code
4. **Quality**: Produce high-quality, "cutthroat" software

### Style Guidelines
- **Comments**: Avoid excessive comments; code should be self-explanatory
- **Boilerplate**: Minimize repeated code patterns
- **Documentation**: File-level comments should explain purpose and role
- **Function Documentation**: Document functions with clear descriptions of inputs/outputs
- **External Dependencies**: Comment external function calls with their purpose

### Code Organization
- **Functional Modularity**: Well-defined, reusable functions with single, clear purposes
- **File Modularity**: Organize codebase across multiple files to reduce complexity
- **Black-box Design**: Intentionally isolate core modules into separate files

## Development Workflow

### Running the Application
- Server entry point: `cmd/server/main.go`
- Default HTTP address: `:8080`
- Storage backend: `memory` (default) or `sql` (planned)
- Admin seeding: Use `-seed-admin` flag to create test admin user

### Configuration
- HTTP listen address: `-http-addr` flag
- Storage backend: `-storage` flag (memory|sql)
- Database DSN: `-database-dsn` flag (required for SQL backend)
- Bcrypt cost: `-bcrypt-cost` flag

## API Endpoints

### Authentication
- `POST /api/register` - User registration
- `POST /api/login` - User login
- `POST /api/logout` - User logout
- `GET /api/me` - Get current user

### Proposals
- `POST /api/proposals` - Create proposal (requires auth)
- `GET /api/proposals` - List user's proposals (requires auth)
- `GET /api/proposals/{id}` - Get proposal (requires auth)
- `PATCH /api/proposals/{id}` - Update proposal (requires auth)
- `POST /api/proposals/{id}/actions/{action}` - Perform workflow action (requires auth)

### Pages
- `GET /` - Home page
- `GET /login` - Login page
- `GET /register` - Registration page
- `GET /profile` - User profile page
- `GET /proposals` - Proposals list page
- `GET /proposals/new` - New proposal page
- `GET /proposals/{id}` - View proposal page
- `GET /proposals/{id}/edit` - Edit proposal page

## Additional Details to Consider

### Missing Information for Knowledge Base

1. **Business Logic**
   - What criteria determine proposal approval/rejection?
   - Are there limits on proposals per user?
   - What happens to approved proposals? Do they automatically become courses?

2. **Course Model**
   - What is the relationship between Proposal and Course?
   - What fields does a Course have beyond Proposal fields?
   - Can a Course have multiple instructors?
   - How is course content structured (lessons, modules, etc.)?

3. **User Experience**
   - What "relevant features" should courses be sortable by? (topic, difficulty, duration, rating, etc.)
   - How do users discover courses?
   - What is the course completion mechanism?
   - Are there certificates or credentials upon completion?

4. **Technical Decisions**
   - Which SQL database will be used? (PostgreSQL, MySQL, SQLite?)
   - Are there plans for file uploads (course materials, images)?
   - Will there be real-time features (chat, notifications)?
   - Deployment strategy and hosting preferences?

5. **Security & Permissions**
   - Role-based access control details beyond basic roles
   - Can instructors edit their courses after approval?
   - Can students rate/review courses?
   - Data privacy and GDPR considerations?

6. **Content Management**
   - How is course content stored? (Markdown, HTML, files?)
   - Assignment submission format
   - Grading/feedback system for assignments

7. **Development Environment**
   - Local development setup requirements
   - Testing strategy beyond E2E tests
   - CI/CD pipeline preferences

8. **Performance & Scale**
   - Expected user volume
   - Caching strategy
   - Search functionality requirements

## Notes

- The project uses a strict mode-based development protocol (RESEARCH → INNOVATE → PLAN → EXECUTE)
- All code changes must follow explicit mode transitions
- Current focus appears to be on the proposal submission workflow
- The transition from proposals to courses is not yet defined in the codebase

