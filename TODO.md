# TODO.md

## Course Enrollment

### `/courses/{courseID}`

This page shall be accessible to unauthorized users to give them a preview for the course content.
It will have the enroll button, which will redirect an unauthorized user to the login page. 
It should show important course info in a nice, aesthetic way.

### `/courses/{courseID}/home`

This is the home page for the course, accessable to users who are enrolled.
It should be similar to `/courses/{courseID}`, but not have the enroll button and we will add more to it later.

### `/courses/{courseID}/content`

This should have a module-based navbar on the left to allow users to click between modules and view them dropped down.
This should be similar to brightspace.

### `/courses/{courseID}/modules/{moduleID}`

The view of a single module.
Should list the content inside that module in an intuitive way

## Codebase and Architecture
