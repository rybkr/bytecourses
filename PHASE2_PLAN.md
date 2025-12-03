# Phase 2: Enrollment UI - Implementation Plan

## Overview
Add enrollment functionality to the course detail page, allowing students to enroll/unenroll and displaying enrollment status and count.

---

## Task 1: Update Course API Response

### File: `internal/handlers/courses.go`

**Goal:** Enhance `GetCourse` handler to include enrollment status when user is authenticated.

**Changes:**
1. Check if user is authenticated (optional - don't require auth for viewing course)
2. If authenticated, check enrollment status using `store.GetEnrollment()`
3. Get enrollment count using `store.GetEnrollmentCount()`
4. Include enrollment data in response

**Response Structure:**
```json
{
  "course_with_instructor": {
    "course": { ... },
    "instructor_name": "...",
    "instructor_email": "..."
  },
  "enrollment": {
    "is_enrolled": true/false,
    "enrolled_at": "2024-01-01T00:00:00Z" (if enrolled),
    "last_accessed_at": "2024-01-01T00:00:00Z" (if enrolled)
  },
  "enrollment_count": 42
}
```

**Implementation Details:**
- Use `middleware.GetUserFromContext()` to get user (may return nil if not authenticated)
- If user is nil, set `is_enrolled: false` and still return `enrollment_count`
- If user exists, check enrollment status
- Update last accessed timestamp when user views course (if enrolled)
- Handle errors gracefully (if enrollment check fails, default to not enrolled)

**Code Location:** Modify `GetCourse()` method in `internal/handlers/courses.go`

---

## Task 2: Add Enrollment UI Elements

### File: `static/course/index.html`

**Goal:** Add enrollment buttons and status indicators to the course header.

**Location:** Insert after the course description div (around line 217), before closing `course-header` div.

**HTML Structure:**
```html
<div class="course-actions" id="courseActions" style="display: none;">
  <div class="course-actions-primary">
    <button id="enrollBtn" class="btn btn-primary" style="display: none;">
      Enroll in Course
    </button>
    <button id="unenrollBtn" class="btn btn-secondary" style="display: none;">
      Unenroll
    </button>
    <span id="enrollmentBadge" class="enrollment-badge" style="display: none;">
      ✓ Enrolled
    </span>
  </div>
  <div class="course-actions-meta">
    <span id="enrollmentCount" class="enrollment-count" style="display: none;">
      <strong id="enrollmentCountNumber">0</strong> students enrolled
    </span>
  </div>
</div>
```

**Requirements:**
- `course-actions` container should be visible by default (remove `display: none` after initial load)
- Buttons should be hidden initially, shown based on enrollment status
- Enrollment count should always be visible (if > 0)
- Badge should only show when enrolled
- Responsive layout for mobile

---

## Task 3: Add Enrollment JavaScript Methods

### File: `static/js/course-viewer.js`

**Goal:** Add enrollment functionality to the course viewer module.

### 3.1 Add Properties
Add to `courseViewerModule` object:
- `courseId: null` - Store current course ID
- `enrollmentStatus: null` - Store enrollment status object

### 3.2 Add Methods

**`checkEnrollmentStatus(courseId)`**
- Fetch enrollment status from `api.enrollments.getStatus(courseId)`
- Store result in `this.enrollmentStatus`
- Call `updateEnrollmentUI()` to update UI
- Call `updateLastAccessed(courseId)` if enrolled
- Handle errors gracefully (show default state if API fails)

**`updateEnrollmentUI()`**
- Get enrollment status from `this.enrollmentStatus`
- Show/hide buttons based on:
  - If not authenticated: Hide all enrollment UI
  - If authenticated and not enrolled: Show "Enroll" button
  - If authenticated and enrolled: Show "Unenroll" button and badge
- Update enrollment count display
- Show `course-actions` container

**`handleEnroll()`**
- Call `api.enrollments.enroll(this.courseId)`
- On success: Refresh enrollment status and update UI
- On error: Show error message (use existing `showError` or create toast)
- Disable button during request to prevent double-clicks

**`handleUnenroll()`**
- Confirm with user: "Are you sure you want to unenroll from this course?"
- Call `api.enrollments.unenroll(this.courseId)`
- On success: Refresh enrollment status and update UI
- On error: Show error message
- Disable button during request

**`updateLastAccessed(courseId)`**
- Call `api.enrollments.updateLastAccessed(courseId)` (if we add this endpoint)
- OR: Just update locally, backend will handle on next enrollment status check
- This is optional - can be handled by backend when checking status

**`showEnrollmentMessage(message, type)`**
- Display success/error messages for enrollment actions
- Use existing error display or create toast notification
- Type: "success" or "error"

### 3.3 Integration Points

**Modify `init()`:**
- Store `courseId` in `this.courseId`

**Modify `loadCourse(courseId)`:**
- After successfully loading course, call `this.checkEnrollmentStatus(courseId)`
- Store courseId: `this.courseId = courseId`

**Modify `renderCourse(data)`:**
- Extract enrollment data from response if present
- Store in `this.enrollmentStatus`
- Call `this.updateEnrollmentUI()` after rendering course content

**Add Event Listeners:**
- In `init()` or after DOM ready, add:
  - `document.getElementById("enrollBtn").addEventListener("click", () => this.handleEnroll())`
  - `document.getElementById("unenrollBtn").addEventListener("click", () => this.handleUnenroll())`

---

## Task 4: Update API Response Handling

### File: `static/js/course-viewer.js`

**Modify `renderCourse(data)`:**
- Check if `data.enrollment` exists
- Check if `data.enrollment_count` exists
- Store enrollment data: `this.enrollmentStatus = data.enrollment || { is_enrolled: false }`
- Store enrollment count separately or in status object

**Fallback:**
- If enrollment data not in response, call `checkEnrollmentStatus()` separately
- This ensures backward compatibility if API doesn't include enrollment data

---

## Task 5: Add CSS Styles

### File: `static/styles.css`

**Goal:** Style enrollment UI elements to match existing design system.

**Styles to Add:**

```css
/* Course Actions Container */
.course-actions {
  margin-top: var(--spacing-xl);
  padding-top: var(--spacing-lg);
  border-top: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.course-actions-primary {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  flex-wrap: wrap;
}

/* Enrollment Badge */
.enrollment-badge {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-xs) var(--spacing-md);
  background: var(--color-success-light);
  color: var(--color-success);
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
  font-weight: 500;
}

/* Enrollment Count */
.enrollment-count {
  color: var(--color-text-light);
  font-size: 0.9375rem;
}

.enrollment-count strong {
  color: var(--color-text);
  font-weight: 600;
}

/* Course Actions Meta */
.course-actions-meta {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

/* Mobile Responsive */
@media (max-width: 768px) {
  .course-actions {
    flex-direction: column;
    gap: var(--spacing-sm);
  }

  .course-actions-primary {
    width: 100%;
  }

  .course-actions-primary button {
    flex: 1;
    min-width: 0;
  }
}
```

**Integration:**
- Use existing CSS variables (`--color-success`, `--color-border`, etc.)
- Match spacing and radius variables from existing styles
- Ensure buttons use existing `.btn`, `.btn-primary`, `.btn-secondary` classes

---

## Task 6: Handle Authentication State

### File: `static/js/course-viewer.js`

**Goal:** Show/hide enrollment UI based on authentication status.

**Implementation:**
- Check if user is authenticated (check for token in sessionStorage)
- If not authenticated:
  - Hide enrollment buttons
  - Still show enrollment count (public info)
- If authenticated:
  - Show appropriate enrollment UI based on status
  - Check user role (only students can enroll)

**Method:**
```javascript
isAuthenticated() {
  return !!sessionStorage.getItem("authToken");
}

getUserRole() {
  // Could decode JWT token or store in separate variable
  // For now, we'll rely on API responses
  return null; // Will be determined by API
}
```

**Note:** User role checking will be handled by backend API, so we can rely on API responses to determine what to show.

---

## Implementation Order

1. **Task 1** - Update backend API to include enrollment data
2. **Task 2** - Add HTML structure for enrollment UI
3. **Task 5** - Add CSS styles (can be done in parallel with Task 3)
4. **Task 3** - Add JavaScript methods for enrollment
5. **Task 4** - Update API response handling
6. **Task 6** - Add authentication state handling

---

## Testing Checklist

- [ ] Course page loads enrollment status when user is authenticated
- [ ] Course page shows enrollment count for all users (public)
- [ ] "Enroll" button appears for authenticated students who are not enrolled
- [ ] "Unenroll" button appears for enrolled students
- [ ] Enrollment badge shows when enrolled
- [ ] Clicking "Enroll" successfully enrolls user and updates UI
- [ ] Clicking "Unenroll" shows confirmation and successfully unenrolls
- [ ] Enrollment count updates after enrollment/unenrollment
- [ ] UI handles errors gracefully (API failures, network errors)
- [ ] Mobile responsive design works correctly
- [ ] Non-authenticated users see enrollment count but no buttons
- [ ] Instructors viewing their own courses see appropriate UI (no enroll button)

---

## Edge Cases to Handle

1. **User not authenticated:** Show enrollment count only, hide buttons
2. **User is instructor:** Don't show enroll button (instructors can't enroll in their own courses)
3. **User is admin:** May want to show different UI (optional)
4. **API errors:** Show error message, don't break page
5. **Network failures:** Handle gracefully, allow retry
6. **Concurrent enrollments:** Backend handles with unique constraint
7. **Enrollment count is 0:** Still show "0 students enrolled" or hide?

---

## Estimated Effort

- **Task 1 (Backend):** 1-2 hours
- **Task 2 (HTML):** 30 minutes
- **Task 3 (JavaScript):** 2-3 hours
- **Task 4 (Integration):** 1 hour
- **Task 5 (CSS):** 1 hour
- **Task 6 (Auth handling):** 30 minutes

**Total:** ~6-8 hours

