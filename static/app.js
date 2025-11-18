const API_BASE = "/api";

let authToken = sessionStorage.getItem("authToken");
let currentUser = null;

const authView = document.getElementById("authView");
const coursesView = document.getElementById("coursesView");
const submitView = document.getElementById("submitView");
const myCoursesView = document.getElementById("myCoursesView");
const adminView = document.getElementById("adminView");
const mainNav = document.getElementById("mainNav");
const userInfo = document.getElementById("userInfo");

const authForm = document.getElementById("authForm");
const authTitle = document.getElementById("authTitle");
const authSubmitBtn = document.getElementById("authSubmitBtn");
const authToggleBtn = document.getElementById("authToggleBtn");
const authToggleText = document.getElementById("authToggleText");
const roleGroup = document.getElementById("roleGroup");
const authMessage = document.getElementById("authMessage");

const viewCoursesBtn = document.getElementById("viewCoursesBtn");
const submitCourseBtn = document.getElementById("submitCourseBtn");
const myCoursesBtn = document.getElementById("myCoursesBtn");
const adminBtn = document.getElementById("adminBtn");
const logoutBtn = document.getElementById("logoutBtn");
const coursesList = document.getElementById("coursesList");
const courseForm = document.getElementById("courseForm");
const statusFilter = document.getElementById("statusFilter");
const formMessage = document.getElementById("formMessage");
const usersList = document.getElementById("usersList");
const pendingCoursesList = document.getElementById("pendingCoursesList");
const myCoursesList = document.getElementById("myCoursesList");
const myCoursesMessage = document.getElementById("myCoursesMessage");

const editModal = document.getElementById("editModal");
const editCourseForm = document.getElementById("editCourseForm");
const closeModal = document.getElementsByClassName("close")[0];

let isSignupMode = false;

if (authToken) {
	fetchCurrentUser();
}

closeModal.onclick = function () {
	editModal.style.display = "none";
};

window.onclick = function (event) {
	if (event.target == editModal) {
		editModal.style.display = "none";
	}
};

authToggleBtn.addEventListener("click", () => {
	isSignupMode = !isSignupMode;
	if (isSignupMode) {
		authTitle.textContent = "Sign Up";
		authSubmitBtn.textContent = "Sign Up";
		authToggleText.textContent = "Already have an account?";
		authToggleBtn.textContent = "Login";
		roleGroup.style.display = "block";
	} else {
		authTitle.textContent = "Login";
		authSubmitBtn.textContent = "Login";
		authToggleText.textContent = "Don't have an account?";
		authToggleBtn.textContent = "Sign up";
		roleGroup.style.display = "none";
	}
	authMessage.style.display = "none";
});

authForm.addEventListener("submit", async (e) => {
	e.preventDefault();

	const email = document.getElementById("authEmail").value;
	const password = document.getElementById("authPassword").value;
	const role = document.getElementById("authRole").value;

	const endpoint = isSignupMode ? "/auth/signup" : "/auth/login";
	const body = isSignupMode ? { email, password, role } : { email, password };

	try {
		const response = await fetch(`${API_BASE}${endpoint}`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(body),
		});

		if (response.ok) {
			const data = await response.json();
			authToken = data.token;
			currentUser = data.user;
			sessionStorage.setItem("authToken", authToken);
			showAuthenticatedUI();
			loadCourses();
		} else {
			const error = await response.text();
			showAuthMessage(error || "Authentication failed", "error");
		}
	} catch (error) {
		showAuthMessage("Error: " + error.message, "error");
	}
});

logoutBtn.addEventListener("click", () => {
	authToken = null;
	currentUser = null;
	sessionStorage.removeItem("authToken");
	showUnauthenticatedUI();
});

viewCoursesBtn.addEventListener("click", () => {
	showView("courses");
	loadCourses();
});

submitCourseBtn.addEventListener("click", () => {
	showView("submit");
});

myCoursesBtn.addEventListener("click", () => {
	showView("myCourses");
	loadMyCourses();
});

adminBtn.addEventListener("click", () => {
	showView("admin");
	loadAdminData();
});

statusFilter.addEventListener("change", loadCourses);

courseForm.addEventListener("submit", async (e) => {
	e.preventDefault();

	const formData = {
		title: document.getElementById("title").value,
		description: document.getElementById("description").value,
	};

	try {
		const response = await fetch(`${API_BASE}/courses`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
				Authorization: `Bearer ${authToken}`,
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

editCourseForm.addEventListener("submit", async (e) => {
	e.preventDefault();

	const courseId = document.getElementById("editCourseId").value;
	const formData = {
		title: document.getElementById("editTitle").value,
		description: document.getElementById("editDescription").value,
	};

	try {
		const response = await fetch(
			`${API_BASE}/instructor/courses?id=${courseId}`,
			{
				method: "PATCH",
				headers: {
					"Content-Type": "application/json",
					Authorization: `Bearer ${authToken}`,
				},
				body: JSON.stringify(formData),
			},
		);

		if (response.ok) {
			editModal.style.display = "none";
			showMyCoursesMessage("Course updated successfully!", "success");
			loadMyCourses();
		} else {
			alert("Failed to update course");
		}
	} catch (error) {
		alert("Error: " + error.message);
	}
});

function parseJWT(token) {
	try {
		const base64Url = token.split(".")[1];
		const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
		const jsonPayload = decodeURIComponent(
			atob(base64)
				.split("")
				.map(function (c) {
					return "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2);
				})
				.join(""),
		);
		return JSON.parse(jsonPayload);
	} catch (e) {
		return null;
	}
}

async function fetchCurrentUser() {
	const payload = parseJWT(authToken);
	if (!payload) {
		authToken = null;
		sessionStorage.removeItem("authToken");
		showUnauthenticatedUI();
		return;
	}

	if (payload.exp && payload.exp * 1000 < Date.now()) {
		authToken = null;
		sessionStorage.removeItem("authToken");
		showUnauthenticatedUI();
		return;
	}

	currentUser = {
		id: payload.user_id,
		email: payload.email,
		role: payload.role,
	};

	showAuthenticatedUI();
	loadCourses();
}

function showAuthenticatedUI() {
	authView.classList.remove("active");
	coursesView.classList.add("active");
	mainNav.style.display = "block";
	userInfo.style.display = "block";

	if (currentUser) {
		userInfo.textContent = `Logged in as ${currentUser.email} (${currentUser.role})`;

		if (currentUser.role === "admin") {
			adminBtn.style.display = "inline-block";
		}

		if (currentUser.role === "instructor" || currentUser.role === "admin") {
			myCoursesBtn.style.display = "inline-block";
		}
	}
}

function showUnauthenticatedUI() {
	authView.classList.add("active");
	coursesView.classList.remove("active");
	submitView.classList.remove("active");
	myCoursesView.classList.remove("active");
	adminView.classList.remove("active");
	mainNav.style.display = "none";
	userInfo.style.display = "none";
	adminBtn.style.display = "none";
	myCoursesBtn.style.display = "none";
	authForm.reset();
}

function showView(view) {
	coursesView.classList.remove("active");
	submitView.classList.remove("active");
	myCoursesView.classList.remove("active");
	adminView.classList.remove("active");
	viewCoursesBtn.classList.remove("active");
	submitCourseBtn.classList.remove("active");
	myCoursesBtn.classList.remove("active");
	adminBtn.classList.remove("active");

	if (view === "courses") {
		coursesView.classList.add("active");
		viewCoursesBtn.classList.add("active");
	} else if (view === "submit") {
		submitView.classList.add("active");
		submitCourseBtn.classList.add("active");
	} else if (view === "myCourses") {
		myCoursesView.classList.add("active");
		myCoursesBtn.classList.add("active");
	} else if (view === "admin") {
		adminView.classList.add("active");
		adminBtn.classList.add("active");
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

async function loadMyCourses() {
	try {
		const response = await fetch(`${API_BASE}/instructor/courses`, {
			headers: {
				Authorization: `Bearer ${authToken}`,
			},
		});

		if (response.ok) {
			const courses = await response.json();
			renderMyCourses(courses);
		} else {
			myCoursesList.innerHTML = "<p>Error loading your courses</p>";
		}
	} catch (error) {
		myCoursesList.innerHTML = "<p>Error loading your courses</p>";
	}
}

async function loadAdminData() {
	await loadUsers();
	await loadPendingCourses();
}

async function loadUsers() {
	try {
		const response = await fetch(`${API_BASE}/admin/users`, {
			headers: {
				Authorization: `Bearer ${authToken}`,
			},
		});

		if (response.ok) {
			const users = await response.json();
			renderUsers(users);
		} else {
			usersList.innerHTML = "<p>Error loading users</p>";
		}
	} catch (error) {
		usersList.innerHTML = "<p>Error loading users</p>";
	}
}

async function loadPendingCourses() {
	try {
		const response = await fetch(`${API_BASE}/admin/courses?status=pending`, {
			headers: {
				Authorization: `Bearer ${authToken}`,
			},
		});

		if (response.ok) {
			const courses = await response.json();
			renderPendingCourses(courses);
		} else {
			pendingCoursesList.innerHTML = "<p>Error loading pending courses</p>";
		}
	} catch (error) {
		pendingCoursesList.innerHTML = "<p>Error loading pending courses</p>";
	}
}

function renderUsers(users) {
	if (!users || users.length === 0) {
		usersList.innerHTML = "<p>No users found</p>";
		return;
	}

	usersList.innerHTML = users
		.map(
			(user) => `
        <div class="user-card">
            <div class="user-info">
                <div class="user-email">${escapeHtml(user.email)}</div>
                <div class="user-role">Role: ${user.role}</div>
            </div>
            <div>ID: ${user.id}</div>
        </div>
    `,
		)
		.join("");
}

function renderPendingCourses(courses) {
	if (!courses || courses.length === 0) {
		pendingCoursesList.innerHTML = "<p>No pending courses</p>";
		return;
	}

	pendingCoursesList.innerHTML = courses
		.map(
			(course) => `
        <div class="pending-course-card">
            <h4>${escapeHtml(course.title)}</h4>
            <p>${escapeHtml(course.description)}</p>
            <div>Instructor ID: ${course.instructor_id}</div>
            <div class="course-actions">
                <button class="approve-btn" onclick="approveCourse(${course.id})">Approve</button>
                <button class="reject-btn" onclick="rejectCourse(${course.id})">Reject</button>
            </div>
        </div>
    `,
		)
		.join("");
}

function renderMyCourses(courses) {
	if (!courses || courses.length === 0) {
		myCoursesList.innerHTML = "<p>You haven't created any courses yet</p>";
		return;
	}

	myCoursesList.innerHTML = courses
		.map(
			(course) => `
        <div class="my-course-card">
            <h3>${escapeHtml(course.title)}</h3>
            <p>${escapeHtml(course.description)}</p>
            <div class="my-course-meta">
                <span class="status-badge status-${course.status}">${course.status}</span>
                <div class="my-course-actions">
                    <button class="edit-btn" onclick="openEditModal(${course.id}, '${escapeHtml(course.title).replace(/'/g, "\\'")}', '${escapeHtml(course.description).replace(/'/g, "\\'")}')">Edit</button>
                    <button class="delete-btn-small" onclick="deleteMyCourse(${course.id})">Delete</button>
                </div>
            </div>
        </div>
    `,
		)
		.join("");
}

function openEditModal(id, title, description) {
	document.getElementById("editCourseId").value = id;
	document.getElementById("editTitle").value = title;
	document.getElementById("editDescription").value = description;
	editModal.style.display = "block";
}

async function deleteMyCourse(id) {
	if (!confirm("Are you sure you want to delete this course?")) {
		return;
	}

	try {
		const response = await fetch(`${API_BASE}/instructor/courses?id=${id}`, {
			method: "DELETE",
			headers: {
				Authorization: `Bearer ${authToken}`,
			},
		});

		if (response.ok) {
			showMyCoursesMessage("Course deleted successfully", "success");
			loadMyCourses();
		} else {
			alert("Failed to delete course");
		}
	} catch (error) {
		alert("Error: " + error.message);
	}
}

async function approveCourse(id) {
	try {
		const response = await fetch(`${API_BASE}/admin/courses/approve?id=${id}`, {
			method: "PATCH",
			headers: {
				Authorization: `Bearer ${authToken}`,
			},
		});

		if (response.ok) {
			loadPendingCourses();
		} else {
			alert("Failed to approve course");
		}
	} catch (error) {
		alert("Error: " + error.message);
	}
}

async function rejectCourse(id) {
	try {
		const response = await fetch(`${API_BASE}/admin/courses/reject?id=${id}`, {
			method: "PATCH",
			headers: {
				Authorization: `Bearer ${authToken}`,
			},
		});

		if (response.ok) {
			loadPendingCourses();
		} else {
			alert("Failed to reject course");
		}
	} catch (error) {
		alert("Error: " + error.message);
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

function showMyCoursesMessage(message, type) {
	myCoursesMessage.textContent = message;
	myCoursesMessage.className = type;
	setTimeout(() => {
		myCoursesMessage.style.display = "none";
	}, 3000);
}

function showAuthMessage(message, type) {
	authMessage.textContent = message;
	authMessage.className = type;
}

function escapeHtml(text) {
	const div = document.createElement("div");
	div.textContent = text;
	return div.innerHTML;
}
