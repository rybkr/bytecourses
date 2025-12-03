## Course Pages

- [ ] Create course detail page with full course information
  - [ ] Display course title, description, instructor info
  - [ ] Show course content sections (lessons, materials, assignments)
  - [ ] Add enrollment button for students (if not enrolled)
  - [ ] Show enrollment status and progress for enrolled students
  - [ ] Display course announcements/updates
  - [ ] Add navigation between course sections
  - [ ] Implement section completion tracking UI

- [ ] Build instructor course management page
  - [ ] Course overview dashboard (enrollment stats, completion rates)
  - [ ] Content editor with section management (add/edit/delete sections)
  - [ ] Assignment creation and management interface
  - [ ] Announcement posting interface
  - [ ] Student roster view (list enrolled students)
  - [ ] Gradebook view (all assignments and student submissions)

- [ ] Create student course dashboard
  - [ ] List of enrolled courses
  - [ ] Course progress overview (completed vs. total sections)
  - [ ] Upcoming assignments and due dates
  - [ ] Recent announcements from instructors
  - [ ] Quick access to in-progress courses

## Enrollment System

- [ ] Database schema for enrollments
  - [ ] Create `enrollments` table (student_id, course_id, enrolled_at, status)
  - [ ] Add indexes for efficient queries
  - [ ] Migration files (up/down)

- [ ] Enrollment API endpoints
  - [ ] `POST /api/courses/{id}/enroll` - Student enrolls in course
  - [ ] `DELETE /api/courses/{id}/enroll` - Student unenrolls
  - [ ] `GET /api/courses/{id}/enrollments` - Get enrollment status (student) or list (instructor)
  - [ ] `GET /api/students/enrollments` - Get all courses student is enrolled in

- [ ] Enrollment business logic
  - [ ] Prevent duplicate enrollments
  - [ ] Handle enrollment limits (if applicable)
  - [ ] Validate student role for enrollment

- [ ] Frontend enrollment UI
  - [ ] Enroll/unenroll button on course detail page
  - [ ] Enrollment status indicator
  - [ ] My Courses page showing enrolled courses

## Assignment System

- [ ] Database schema for assignments
  - [ ] Create `assignments` table (course_id, title, description, due_date, points, created_at)
  - [ ] Create `submissions` table (assignment_id, student_id, content, file_paths, submitted_at, status)
  - [ ] Create `grades` table (submission_id, instructor_id, score, feedback, graded_at)
  - [ ] Add appropriate indexes and foreign keys

- [ ] Assignment API endpoints
  - [ ] `POST /api/courses/{id}/assignments` - Instructor creates assignment
  - [ ] `GET /api/courses/{id}/assignments` - List assignments for course
  - [ ] `PATCH /api/assignments/{id}` - Instructor updates assignment
  - [ ] `DELETE /api/assignments/{id}` - Instructor deletes assignment
  - [ ] `POST /api/assignments/{id}/submit` - Student submits assignment
  - [ ] `GET /api/assignments/{id}/submissions` - Instructor views submissions
  - [ ] `PATCH /api/submissions/{id}/grade` - Instructor grades submission

- [ ] Assignment frontend
  - [ ] Assignment creation form (instructor)
  - [ ] Assignment list view (course page)
  - [ ] Assignment detail page (description, due date, points)
  - [ ] Submission form (student) - text and file uploads
  - [ ] Submission status view (student)
  - [ ] Grading interface (instructor) - score input, feedback textarea
  - [ ] Grade view (student) - see score and feedback

## File Upload System

- [ ] File storage setup
  - [ ] Decide on storage backend (local filesystem vs. S3/MinIO)
  - [ ] Create upload directory structure
  - [ ] Implement file validation (type, size limits)
  - [ ] Generate unique filenames to prevent conflicts

- [ ] File upload API
  - [ ] `POST /api/upload` - Generic file upload endpoint
  - [ ] `GET /api/files/{id}` - Download/view file
  - [ ] `DELETE /api/files/{id}` - Delete file (with permission checks)
  - [ ] Store file metadata in database (optional `files` table)

- [ ] File upload frontend
  - [ ] File input component with drag-and-drop
  - [ ] File preview/display
  - [ ] Progress indicator for uploads
  - [ ] File list management (view, delete)

## Progress Tracking

- [ ] Database schema for progress
  - [ ] Create `course_progress` table (enrollment_id, section_id, completed_at, last_accessed)
  - [ ] Create `assignment_progress` table (submission_id, status, viewed_at)
  - [ ] Track overall course completion percentage

- [ ] Progress API endpoints
  - [ ] `POST /api/courses/{id}/sections/{section_id}/complete` - Mark section complete
  - [ ] `GET /api/courses/{id}/progress` - Get student progress
  - [ ] `GET /api/courses/{id}/progress/all` - Instructor views all student progress

- [ ] Progress frontend
  - [ ] Progress bar on course page
  - [ ] Section completion checkboxes
  - [ ] Overall progress dashboard
  - [ ] Completion certificates (future)

## Communication & Announcements

- [ ] Database schema for announcements
  - [ ] Create `announcements` table (course_id, instructor_id, title, content, created_at)
  - [ ] Create `announcement_reads` table (announcement_id, student_id, read_at) - optional

- [ ] Announcement API endpoints
  - [ ] `POST /api/courses/{id}/announcements` - Instructor creates announcement
  - [ ] `GET /api/courses/{id}/announcements` - List announcements for course
  - [ ] `DELETE /api/announcements/{id}` - Instructor deletes announcement

- [ ] Announcement frontend
  - [ ] Announcement creation form (instructor)
  - [ ] Announcement feed on course page
  - [ ] Recent announcements widget (student dashboard)

## Gradebook

- [ ] Gradebook API endpoints
  - [ ] `GET /api/courses/{id}/gradebook` - Instructor views all grades
  - [ ] `GET /api/students/grades` - Student views their grades
  - [ ] Export gradebook as CSV (optional)

- [ ] Gradebook frontend
  - [ ] Instructor gradebook view (table: students x assignments)
  - [ ] Student grades view (list of assignments with scores)
  - [ ] Overall course grade calculation
  - [ ] Grade statistics (average, distribution)

## UI/UX Enhancements

- [ ] Improve course content rendering
  - [ ] Rich text formatting support
  - [ ] Embedded media (videos, images)
  - [ ] Code syntax highlighting for technical courses
  - [ ] Markdown support for content

- [ ] Mobile responsiveness
  - [ ] Test and fix mobile layouts
  - [ ] Touch-friendly interactions
  - [ ] Mobile-optimized file uploads

- [ ] Notifications system (optional)
  - [ ] In-app notifications for new assignments, grades, announcements
  - [ ] Email notifications (future)

## Testing & Polish

- [ ] Write tests for critical paths
  - [ ] Enrollment flow
  - [ ] Assignment submission
  - [ ] Grading workflow

- [ ] Error handling improvements
  - [ ] Better error messages for users
  - [ ] Validation feedback

- [ ] Performance optimization
  - [ ] Database query optimization
  - [ ] File upload size limits and handling
  - [ ] Pagination for large lists

