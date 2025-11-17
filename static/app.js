const API_BASE = "/api";

const viewCoursesBtn = document.getElementById("viewCoursesBtn");
const submitCourseBtn = document.getElementById("submitCourseBtn");
const coursesView = document.getElementById("coursesView");
const submitView = document.getElementById("submitView");
const coursesList = document.getElementById("coursesList");
const courseForm = document.getElementById("courseForm");
const statusFilter = document.getElementById("statusFilter");
const formMessage = document.getElementById("formMessage");

viewCoursesBtn.addEventListener("click", () => {
	showView("courses");
	loadCourses();
});

submitCourseBtn.addEventListener("click", () => {
	showView("submit");
});

statusFilter.addEventListener("change", loadCourses);

courseForm.addEventListener("submit", async (e) => {
	e.preventDefault();

	const formData = {
		instructor_id: 1,
		title: document.getElementById("title").value,
		description: document.getElementById("description").value,
	};

	try {
		const response = await fetch(`${API_BASE}/courses`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(formData),
		});

		if (response.ok) {
			showMessage(
				"Course submitted successfully! Awaiting approval.",
				"success",
			);
			courseForm.reset();
		} else {
			showMessage("Failed to submit course. Please try again.", "error");
		}
	} catch (error) {
		showMessage("Error: " + error.message, "error");
	}
});

function showView(view) {
	coursesView.classList.remove("active");
	submitView.classList.remove("active");
	viewCoursesBtn.classList.remove("active");
	submitCourseBtn.classList.remove("active");

	if (view === "courses") {
		coursesView.classList.add("active");
		viewCoursesBtn.classList.add("active");
	} else {
		submitView.classList.add("active");
		submitCourseBtn.classList.add("active");
	}
}

async function loadCourses() {
	const status = statusFilter.value;
	const url = status
		? `${API_BASE}/courses?status=${status}`
		: `${API_BASE}/courses`;

	try {
		const response = await fetch(url);
		const courses = await response.json();
		renderCourses(courses);
	} catch (error) {
		coursesList.innerHTML = "<p>Error loading courses</p>";
	}
}

function renderCourses(courses) {
	if (!courses || courses.length === 0) {
		coursesList.innerHTML = "<p>No courses found</p>";
		return;
	}

	coursesList.innerHTML = courses
		.map(
			(course) => `
        <div class="course-card">
            <h3>${escapeHtml(course.title)}</h3>
            <p>${escapeHtml(course.description)}</p>
            <div class="course-meta">
                <span>Instructor ID: ${course.instructor_id}</span>
                <span class="status-badge status-${course.status}">${course.status}</span>
            </div>
        </div>
    `,
		)
		.join("");
}

function showMessage(message, type) {
	formMessage.textContent = message;
	formMessage.className = type;
	setTimeout(() => {
		formMessage.style.display = "none";
	}, 5000);
}

function escapeHtml(text) {
	const div = document.createElement("div");
	div.textContent = text;
	return div.innerHTML;
}

loadCourses();
