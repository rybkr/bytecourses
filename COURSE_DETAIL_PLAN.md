# Course Detail Page Enhancement Plan

## Overview
Transform the current basic course viewing page into a full-featured Google Classroom-style course page with enrollment, progress tracking, assignments, announcements, and role-based views.

## Current State
- ✅ Basic course display (title, description, instructor, content sections)
- ✅ JSON content parsing and rendering
- ✅ Error handling and loading states
- ❌ No enrollment functionality
- ❌ No progress tracking
- ❌ No assignments display
- ❌ No announcements
- ❌ No role-based views (instructor vs student)
- ❌ No section navigation

---

## Phase 1: Enrollment System (Foundation)

### 1.1 Database Schema
**File:** `migrations/005_create_enrollments.up.sql`

```sql
CREATE TABLE enrollments (
    id SERIAL PRIMARY KEY,
    student_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    enrolled_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_accessed_at TIMESTAMP,
    UNIQUE(student_id, course_id)
);

CREATE INDEX idx_enrollments_student_id ON enrollments(student_id);
CREATE INDEX idx_enrollments_course_id ON enrollments(course_id);
CREATE INDEX idx_enrollments_student_course ON enrollments(student_id, course_id);
```

**Rollback:** `migrations/005_create_enrollments.down.sql`
```sql
DROP TABLE IF EXISTS enrollments;
```

### 1.2 Backend Models
**File:** `internal/models/enrollment.go` (NEW)
- `Enrollment` struct with fields: ID, StudentID, CourseID, EnrolledAt, LastAccessedAt

### 1.3 Store Layer
**File:** `internal/store/enrollments.go` (NEW)

**Methods:**
- `CreateEnrollment(ctx, studentID, courseID)` - Enroll student, prevent duplicates
- `DeleteEnrollment(ctx, studentID, courseID)` - Unenroll student
- `GetEnrollment(ctx, studentID, courseID)` - Get enrollment status
- `GetEnrollmentsByStudent(ctx, studentID)` - Get all courses student is enrolled in
- `GetEnrollmentsByCourse(ctx, courseID)` - Get all students enrolled in course (for instructor)
- `UpdateLastAccessed(ctx, studentID, courseID)` - Update last accessed timestamp
- `GetEnrollmentCount(ctx, courseID)` - Get total enrollment count

### 1.4 API Handlers
**File:** `internal/handlers/enrollments.go` (NEW)

**Endpoints:**
- `POST /api/courses/{id}/enroll` - Student enrolls (requires auth, student role)
- `DELETE /api/courses/{id}/enroll` - Student unenrolls (requires auth, ownership)
- `GET /api/courses/{id}/enrollments` - Get enrollment status (student) or list (instructor/admin)
- `GET /api/students/enrollments` - Get all courses student is enrolled in

**Business Logic:**
- Prevent duplicate enrollments (check before insert)
- Validate student role for enrollment
- Allow instructor to view enrollment list
- Allow student to view only their own enrollment status

### 1.5 API Integration
**File:** `static/js/api.js`
- Add `enrollments` object with methods:
  - `enroll(courseId)`
  - `unenroll(courseId)`
  - `getStatus(courseId)`
  - `list()` - Get all student enrollments

---

## Phase 2: Course Detail Page - Enrollment & Basic Enhancements

### 2.1 Update Course API Response
**File:** `internal/handlers/courses.go`

**Enhancement:** Add enrollment status to `GetCourse` response when user is authenticated
- Check if user is enrolled
- Include enrollment status in response
- Include enrollment count (public info)

**Response Structure:**
```json
{
  "course_with_instructor": { ... },
  "enrollment": {
    "is_enrolled": true/false,
    "enrolled_at": "...",
    "last_accessed_at": "..."
  },
  "enrollment_count": 42
}
```

### 2.2 Frontend Enrollment UI
**File:** `static/course/index.html`

**Add to course header:**
- Enrollment button (if not enrolled and user is student)
- Enrollment status badge (if enrolled)
- Enrollment count display
- "Unenroll" button (if enrolled)

**HTML Structure:**
```html
<div class="course-actions">
  <button id="enrollBtn" class="btn btn-primary">Enroll in Course</button>
  <button id="unenrollBtn" class="btn btn-secondary">Unenroll</button>
  <span class="enrollment-badge">Enrolled</span>
  <span class="enrollment-count">42 students enrolled</span>
</div>
```

### 2.3 Enrollment JavaScript
**File:** `static/js/course-viewer.js`

**Add methods:**
- `checkEnrollmentStatus()` - Check if user is enrolled
- `handleEnroll()` - Enroll user in course
- `handleUnenroll()` - Unenroll user
- `updateEnrollmentUI()` - Update UI based on enrollment status
- `updateLastAccessed()` - Update last accessed timestamp when page loads

**Integration:**
- Call `checkEnrollmentStatus()` after course loads
- Add event listeners for enroll/unenroll buttons
- Update UI dynamically based on enrollment state

### 2.4 Styling
**File:** `static/styles.css`
- Add styles for `.course-actions`, `.enrollment-badge`, `.enrollment-count`
- Button styles for enroll/unenroll actions
- Responsive design for mobile

---

## Phase 3: Progress Tracking

### 3.1 Database Schema
**File:** `migrations/006_create_course_progress.up.sql`

```sql
CREATE TABLE course_progress (
    id SERIAL PRIMARY KEY,
    enrollment_id INTEGER NOT NULL REFERENCES enrollments(id) ON DELETE CASCADE,
    section_id VARCHAR(100) NOT NULL,
    completed_at TIMESTAMP,
    last_accessed_at TIMESTAMP,
    UNIQUE(enrollment_id, section_id)
);

CREATE INDEX idx_progress_enrollment_id ON course_progress(enrollment_id);
CREATE INDEX idx_progress_section_id ON course_progress(section_id);
```

**Note:** `section_id` is a string identifier from the JSON content structure (e.g., "section-1", or use order index)

### 3.2 Store Layer
**File:** `internal/store/progress.go` (NEW)

**Methods:**
- `MarkSectionComplete(ctx, enrollmentID, sectionID)`
- `GetProgress(ctx, enrollmentID)` - Get all progress for an enrollment
- `GetProgressByCourse(ctx, studentID, courseID)` - Get progress for specific course
- `UpdateLastAccessed(ctx, enrollmentID, sectionID)`
- `GetCompletionPercentage(ctx, enrollmentID, totalSections)` - Calculate completion %

### 3.3 API Handlers
**File:** `internal/handlers/progress.go` (NEW)

**Endpoints:**
- `POST /api/courses/{id}/sections/{sectionId}/complete` - Mark section complete
- `GET /api/courses/{id}/progress` - Get student's progress for course
- `GET /api/courses/{id}/progress/all` - Instructor views all student progress

### 3.4 Frontend Progress UI
**File:** `static/course/index.html`

**Add:**
- Progress bar showing completion percentage
- Section completion checkboxes
- "Mark as Complete" buttons for each section

**HTML Structure:**
```html
<div class="course-progress">
  <div class="progress-bar">
    <div class="progress-fill" style="width: 45%"></div>
  </div>
  <span class="progress-text">45% Complete (9 of 20 sections)</span>
</div>
```

### 3.5 Progress JavaScript
**File:** `static/js/course-viewer.js`

**Add methods:**
- `loadProgress()` - Fetch progress data
- `renderProgress()` - Display progress bar and stats
- `markSectionComplete(sectionId)` - Mark section as complete
- `updateProgressUI()` - Update progress bar and checkboxes
- `renderSectionWithProgress(section, progress)` - Render section with completion checkbox

**Integration:**
- Load progress after course loads (if enrolled)
- Add completion checkboxes to each section
- Update progress bar when section is marked complete

---

## Phase 4: Assignments Integration

### 4.1 Database Schema
**File:** `migrations/007_create_assignments.up.sql`

```sql
CREATE TABLE assignments (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    due_date TIMESTAMP,
    points INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE submissions (
    id SERIAL PRIMARY KEY,
    assignment_id INTEGER NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT,
    submitted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'submitted',
    UNIQUE(assignment_id, student_id)
);

CREATE TABLE grades (
    id SERIAL PRIMARY KEY,
    submission_id INTEGER NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    instructor_id INTEGER NOT NULL REFERENCES users(id),
    score INTEGER,
    feedback TEXT,
    graded_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_assignments_course_id ON assignments(course_id);
CREATE INDEX idx_submissions_assignment_id ON submissions(assignment_id);
CREATE INDEX idx_submissions_student_id ON submissions(student_id);
CREATE INDEX idx_grades_submission_id ON grades(submission_id);
```

### 4.2 Store & Handlers
**Files:** `internal/store/assignments.go`, `internal/handlers/assignments.go`

**Note:** This is a large feature - see TODO.md for full breakdown. For course detail page, we only need:
- `GetAssignmentsByCourse(ctx, courseID)` - List assignments for course
- Display assignments on course page

### 4.3 Frontend Assignments Display
**File:** `static/course/index.html`

**Add assignments section:**
```html
<div class="course-assignments">
  <h2>Assignments</h2>
  <div id="assignmentsList"></div>
</div>
```

**File:** `static/js/course-viewer.js`

**Add methods:**
- `loadAssignments()` - Fetch assignments for course
- `renderAssignments(assignments)` - Display assignment list
- Show due dates, points, submission status

---

## Phase 5: Announcements

### 5.1 Database Schema
**File:** `migrations/008_create_announcements.up.sql`

```sql
CREATE TABLE announcements (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    instructor_id INTEGER NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_announcements_course_id ON announcements(course_id);
CREATE INDEX idx_announcements_created_at ON announcements(created_at);
```

### 5.2 Store & Handlers
**Files:** `internal/store/announcements.go`, `internal/handlers/announcements.go`

**Methods:**
- `CreateAnnouncement(ctx, announcement)`
- `GetAnnouncementsByCourse(ctx, courseID)`
- `DeleteAnnouncement(ctx, announcementID, instructorID)` - Verify ownership

**Endpoints:**
- `POST /api/courses/{id}/announcements` - Create (instructor only)
- `GET /api/courses/{id}/announcements` - List announcements
- `DELETE /api/announcements/{id}` - Delete (instructor only)

### 5.3 Frontend Announcements
**File:** `static/course/index.html`

**Add announcements section:**
```html
<div class="course-announcements">
  <h2>Announcements</h2>
  <div id="announcementsList"></div>
</div>
```

**File:** `static/js/course-viewer.js`

**Add methods:**
- `loadAnnouncements()` - Fetch announcements
- `renderAnnouncements(announcements)` - Display announcement feed
- Show instructor name, timestamp, content

---

## Phase 6: Role-Based Views & Navigation

### 6.1 Instructor View
**File:** `static/js/course-viewer.js`

**Add instructor-specific features:**
- Check if current user is course instructor
- Show "Edit Course" button (link to instructor management)
- Show "View Students" button (link to roster)
- Show "Manage Assignments" button
- Show "Post Announcement" button

**File:** `static/course/index.html`

**Add instructor controls section:**
```html
<div id="instructorControls" class="instructor-controls" style="display: none;">
  <button class="btn btn-primary">Edit Course</button>
  <button class="btn btn-secondary">View Students</button>
  <button class="btn btn-secondary">Manage Assignments</button>
  <button class="btn btn-secondary">Post Announcement</button>
</div>
```

### 6.2 Section Navigation
**File:** `static/course/index.html`

**Add table of contents sidebar:**
```html
<aside class="course-sidebar">
  <h3>Course Contents</h3>
  <nav id="courseTOC">
    <!-- Generated dynamically -->
  </nav>
</aside>
```

**File:** `static/js/course-viewer.js`

**Add methods:**
- `generateTableOfContents(sections)` - Create TOC from sections
- `scrollToSection(sectionId)` - Smooth scroll to section
- `highlightCurrentSection()` - Highlight section in viewport

### 6.3 Enhanced Content Rendering
**File:** `static/js/course-viewer.js`

**Enhance `formatContent()` method:**
- Support markdown (or rich text)
- Code syntax highlighting (if code blocks detected)
- Embedded images/videos
- Better link formatting

---

## Phase 7: Polish & UX

### 7.1 Loading States
- Skeleton loaders for course content
- Progressive loading (load assignments/announcements after main content)

### 7.2 Error Handling
- Better error messages
- Retry mechanisms
- Offline detection

### 7.3 Mobile Responsiveness
- Test and fix mobile layouts
- Touch-friendly interactions
- Collapsible sections on mobile

### 7.4 Accessibility
- ARIA labels for buttons
- Keyboard navigation
- Screen reader support

---

## Implementation Order

1. **Phase 1** - Enrollment System (Foundation - must be done first)
2. **Phase 2** - Enrollment UI on course page
3. **Phase 3** - Progress tracking
4. **Phase 4** - Assignments display (can be done in parallel with Phase 5)
5. **Phase 5** - Announcements
6. **Phase 6** - Role-based views & navigation
7. **Phase 7** - Polish & UX

## Dependencies

- **Phase 2** depends on **Phase 1** (enrollment system)
- **Phase 3** depends on **Phase 1** (enrollments table)
- **Phase 4** can be done independently but needs assignment system
- **Phase 5** can be done independently
- **Phase 6** depends on authentication/user context
- **Phase 7** can be done incrementally throughout

## Estimated Effort

- **Phase 1:** 4-6 hours (database, backend, API)
- **Phase 2:** 2-3 hours (frontend enrollment UI)
- **Phase 3:** 3-4 hours (progress tracking)
- **Phase 4:** 2-3 hours (assignments display - full assignment system is separate)
- **Phase 5:** 2-3 hours (announcements)
- **Phase 6:** 3-4 hours (role-based views, navigation)
- **Phase 7:** 2-3 hours (polish)

**Total:** ~18-26 hours for complete course detail page enhancement

