const courseViewerModule = {
	courseId: null,
	enrollmentStatus: null,
	enrollmentCount: 0,
	isInstructor: false,
	userRole: null,

	init() {
		const courseId = this.extractCourseId();
		if (!courseId) {
			this.showError("Invalid course ID");
			return;
		}
		this.courseId = courseId;
		this.detectUserRole();
		this.loadCourse(courseId);
		this.setupEventListeners();
	},

	detectUserRole() {
		const token = sessionStorage.getItem("authToken");
		if (!token) {
			this.userRole = null;
			return;
		}

		try {
			const base64Url = token.split(".")[1];
			const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
			const jsonPayload = decodeURIComponent(
				atob(base64)
					.split("")
					.map((c) => "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2))
					.join(""),
			);
			const payload = JSON.parse(jsonPayload);
			this.userRole = payload.role || null;
		} catch (error) {
			console.error("Error parsing JWT:", error);
			this.userRole = null;
		}
	},

	setupEventListeners() {
		const enrollBtn = document.getElementById("enrollBtn");
		const unenrollBtn = document.getElementById("unenrollBtn");
		const editCourseBtn = document.getElementById("editCourseBtn");
		const viewStudentsBtn = document.getElementById("viewStudentsBtn");
		const postAnnouncementBtn = document.getElementById("postAnnouncementBtn");

		if (enrollBtn) {
			enrollBtn.addEventListener("click", () => this.handleEnroll());
		}

		if (unenrollBtn) {
			unenrollBtn.addEventListener("click", () => this.handleUnenroll());
		}

		if (editCourseBtn) {
			editCourseBtn.addEventListener("click", () => this.handleEditCourse());
		}

		if (viewStudentsBtn) {
			viewStudentsBtn.addEventListener("click", () => this.handleViewStudents());
		}

		if (postAnnouncementBtn) {
			postAnnouncementBtn.addEventListener("click", () => this.handlePostAnnouncement());
		}
	},

	extractCourseId() {
		const path = window.location.pathname;
		const match = path.match(/\/course\/(\d+)\/?/);
		return match ? parseInt(match[1], 10) : null;
	},

	async loadCourse(courseId) {
		this.showLoading();
		try {
			const data = await api.courses.get(courseId);
			this.renderCourse(data);
		} catch (error) {
			console.error("Error loading course:", error);
			if (error.status === 404 || (error.message && error.message.includes("not found"))) {
				this.showError("Course not found");
			} else {
				this.showError("Failed to load course. Please try again.");
			}
		}
	},

	renderCourse(data) {
		const courseWithInstructor = data.course_with_instructor || data;
		const course = courseWithInstructor.course || courseWithInstructor;

		document.getElementById("courseTitle").textContent = course.title || "Untitled Course";
		document.getElementById("courseDescription").textContent = course.description || "";

		const instructorName = courseWithInstructor.instructor_name || courseWithInstructor.instructor_email || "Unknown";
		document.getElementById("courseInstructor").textContent = instructorName;

		const createdAt = course.created_at ? new Date(course.created_at).toLocaleDateString() : "Unknown";
		document.getElementById("courseCreatedAt").textContent = createdAt;

		if (course.content) {
			this.renderContent(course.content);
		} else {
			document.getElementById("courseSections").innerHTML = "<p style='color: var(--color-text-light); font-style: italic;'>No course content available yet.</p>";
		}

		if (data.enrollment) {
			this.enrollmentStatus = data.enrollment;
		} else {
			this.enrollmentStatus = { is_enrolled: false };
		}

		if (data.enrollment_count !== undefined) {
			this.updateEnrollmentCount(data.enrollment_count);
			this.enrollmentCount = data.enrollment_count;
		}

		if (data.is_instructor !== undefined) {
			this.isInstructor = data.is_instructor;
		}

		this.updateEnrollmentUI();
		this.updateInstructorUI();

		document.getElementById("courseLoading").style.display = "none";
		document.getElementById("courseContent").style.display = "block";
	},

	renderContent(contentJson) {
		const sectionsContainer = document.getElementById("courseSections");

		if (!contentJson || contentJson.trim() === "") {
			sectionsContainer.innerHTML = "<p style='color: var(--color-text-light); font-style: italic;'>No course content available yet.</p>";
			return;
		}

		try {
			const content = JSON.parse(contentJson);

			if (content.sections && Array.isArray(content.sections)) {
				const sortedSections = content.sections.sort((a, b) => (a.order || 0) - (b.order || 0));

				sectionsContainer.innerHTML = sortedSections.map((section, index) => {
					const sectionHtml = this.renderSection(section, index);
					return sectionHtml;
				}).join("");
			} else {
				sectionsContainer.innerHTML = `<div class="course-section"><div class="course-section-content">${this.escapeHtml(contentJson)}</div></div>`;
			}
		} catch (error) {
			console.error("Error parsing course content:", error);
			sectionsContainer.innerHTML = `<div class="course-section"><div class="course-section-content">${this.escapeHtml(contentJson)}</div></div>`;
		}
	},

	renderSection(section, index) {
		const title = section.title || `Section ${index + 1}`;
		const type = section.type || "lesson";
		const content = section.content || "";

		return `
			<div class="course-section" data-type="${type}">
				<h2 class="course-section-title">${this.escapeHtml(title)}</h2>
				<div class="course-section-content">
					${this.formatContent(content)}
				</div>
			</div>
		`;
	},

	formatContent(content) {
		if (!content) return "";

		const escaped = this.escapeHtml(content);
		const withLineBreaks = escaped.replace(/\n/g, "<br>");
		const withLinks = withLineBreaks.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>');

		return withLinks;
	},

	escapeHtml(text) {
		const div = document.createElement("div");
		div.textContent = text;
		return div.innerHTML;
	},

	showLoading() {
		document.getElementById("courseLoading").style.display = "block";
		document.getElementById("courseError").style.display = "none";
		document.getElementById("courseContent").style.display = "none";
	},

	showError(message) {
		document.getElementById("courseLoading").style.display = "none";
		document.getElementById("courseError").style.display = "block";
		document.getElementById("courseErrorMessage").textContent = message;
		document.getElementById("courseContent").style.display = "none";
	},

	async checkEnrollmentStatus(courseId) {
		try {
			const data = await api.enrollments.getStatus(courseId);
			this.enrollmentStatus = data;
			this.updateEnrollmentUI();
			if (data.enrollment_count !== undefined) {
				this.updateEnrollmentCount(data.enrollment_count);
			}
		} catch (error) {
			console.error("Error checking enrollment status:", error);
			this.enrollmentStatus = { is_enrolled: false };
			this.updateEnrollmentUI();
		}
	},

	updateEnrollmentUI() {
		const courseActions = document.getElementById("courseActions");
		const enrollBtn = document.getElementById("enrollBtn");
		const unenrollBtn = document.getElementById("unenrollBtn");
		const enrollmentBadge = document.getElementById("enrollmentBadge");

		if (!courseActions) return;

		const isAuthenticated = this.isAuthenticated();
		const isEnrolled = this.enrollmentStatus && this.enrollmentStatus.is_enrolled;

		if (!isAuthenticated) {
			courseActions.style.display = "flex";
			if (enrollBtn) enrollBtn.style.display = "none";
			if (unenrollBtn) unenrollBtn.style.display = "none";
			if (enrollmentBadge) enrollmentBadge.style.display = "none";
			return;
		}

		courseActions.style.display = "flex";

		if (isEnrolled) {
			if (enrollBtn) enrollBtn.style.display = "none";
			if (unenrollBtn) unenrollBtn.style.display = "block";
			if (enrollmentBadge) enrollmentBadge.style.display = "inline-flex";
		} else {
			if (enrollBtn) enrollBtn.style.display = "block";
			if (unenrollBtn) unenrollBtn.style.display = "none";
			if (enrollmentBadge) enrollmentBadge.style.display = "none";
		}
	},

	updateEnrollmentCount(count) {
		const enrollmentCount = document.getElementById("enrollmentCount");
		const enrollmentCountNumber = document.getElementById("enrollmentCountNumber");

		if (enrollmentCount && enrollmentCountNumber) {
			enrollmentCountNumber.textContent = count;
			enrollmentCount.style.display = "inline";
		}
	},

	async handleEnroll() {
		const enrollBtn = document.getElementById("enrollBtn");
		if (!enrollBtn || !this.courseId) return;

		enrollBtn.disabled = true;
		enrollBtn.textContent = "Enrolling...";

		try {
			await api.enrollments.enroll(this.courseId);
			this.showEnrollmentMessage("Successfully enrolled in course!", "success");
			await this.checkEnrollmentStatus(this.courseId);
			await this.loadCourse(this.courseId);
		} catch (error) {
			console.error("Error enrolling:", error);
			this.showEnrollmentMessage(error.message || "Failed to enroll in course. Please try again.", "error");
		} finally {
			enrollBtn.disabled = false;
			enrollBtn.textContent = "Enroll in Course";
		}
	},

	async handleUnenroll() {
		if (!confirm("Are you sure you want to unenroll from this course?")) {
			return;
		}

		const unenrollBtn = document.getElementById("unenrollBtn");
		if (!unenrollBtn || !this.courseId) return;

		unenrollBtn.disabled = true;
		unenrollBtn.textContent = "Unenrolling...";

		try {
			await api.enrollments.unenroll(this.courseId);
			this.showEnrollmentMessage("Successfully unenrolled from course.", "success");
			await this.checkEnrollmentStatus(this.courseId);
			await this.loadCourse(this.courseId);
		} catch (error) {
			console.error("Error unenrolling:", error);
			this.showEnrollmentMessage(error.message || "Failed to unenroll from course. Please try again.", "error");
		} finally {
			unenrollBtn.disabled = false;
			unenrollBtn.textContent = "Unenroll";
		}
	},

	showEnrollmentMessage(message, type) {
		const courseActions = document.getElementById("courseActions");
		if (!courseActions) return;

		let messageEl = document.getElementById("enrollmentMessage");
		if (!messageEl) {
			messageEl = document.createElement("div");
			messageEl.id = "enrollmentMessage";
			messageEl.className = `enrollment-message enrollment-message-${type}`;
			courseActions.insertBefore(messageEl, courseActions.firstChild);
		}

		messageEl.textContent = message;
		messageEl.className = `enrollment-message enrollment-message-${type}`;
		messageEl.style.display = "block";

		setTimeout(() => {
			messageEl.style.display = "none";
		}, 5000);
	},

	isAuthenticated() {
		return !!sessionStorage.getItem("authToken");
	},

	updateInstructorUI() {
		const instructorDashboard = document.getElementById("instructorDashboard");
		const courseActions = document.getElementById("courseActions");
		const statEnrollmentCount = document.getElementById("statEnrollmentCount");

		if (!instructorDashboard) return;

		if (this.isInstructor) {
			instructorDashboard.style.display = "block";
			if (courseActions) {
				courseActions.style.display = "none";
			}
			if (statEnrollmentCount) {
				statEnrollmentCount.textContent = this.enrollmentCount || 0;
			}
		} else {
			instructorDashboard.style.display = "none";
		}
	},

	handleEditCourse() {
		window.location.href = "/";
		setTimeout(() => {
			if (typeof instructorModule !== "undefined" && instructorModule.load) {
				instructorModule.load();
				const editBtn = document.querySelector(`[data-course-id="${this.courseId}"]`);
				if (editBtn) {
					editBtn.click();
				}
			}
		}, 100);
	},

	handleViewStudents() {
		alert("Student roster feature coming soon!");
	},

	handlePostAnnouncement() {
		alert("Announcement posting feature coming soon!");
	},
};

