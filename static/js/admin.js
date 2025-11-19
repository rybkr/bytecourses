const adminModule = {
	init() {
		document.addEventListener("click", (e) => {
			if (e.target.classList.contains("approve-btn")) {
				const courseId = parseInt(e.target.dataset.courseId);
				this.approveCourse(courseId);
			} else if (e.target.classList.contains("reject-btn")) {
				const courseId = parseInt(e.target.dataset.courseId);
				this.rejectCourse(courseId);
			}
		});
	},

	async load() {
		await this.loadUsers();
		await this.loadPendingCourses();
	},

	async loadUsers() {
		try {
			const users = await api.admin.getUsers();
			this.renderUsers(users);
		} catch (error) {
			document.getElementById("usersList").innerHTML =
				"<p>Error loading users</p>";
		}
	},

	renderUsers(users) {
		const usersList = document.getElementById("usersList");

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
	},

	async loadPendingCourses() {
		try {
			const applications = await api.admin.getApplications();
			this.renderPendingCourses(applications);
		} catch (error) {
			document.getElementById("pendingCoursesList").innerHTML =
				"<p>Error loading pending applications</p>";
		}
	},

	renderPendingCourses(applications) {
		const pendingCoursesList = document.getElementById("pendingCoursesList");

		if (!applications || applications.length === 0) {
			pendingCoursesList.innerHTML = "<p>No pending applications</p>";
			return;
		}

		pendingCoursesList.innerHTML = applications
			.map(
				(app) => `
            <div class="pending-course-card">
                <h4>${escapeHtml(app.title)}</h4>
                <p>${escapeHtml(app.description)}</p>
                <div class="course-actions">
                    <button class="approve-btn" data-course-id="${app.id}">Approve</button>
                    <button class="reject-btn" data-course-id="${app.id}">Reject</button>
                </div>
            </div>
        `,
			)
			.join("");
	},

	async approveCourse(id) {
		try {
			await api.admin.approveApplication(id);
			this.loadPendingCourses();
		} catch (error) {
			this.showError(error.message);
		}
	},

	async rejectCourse(id) {
		try {
			await api.admin.rejectApplication(id);
			this.loadPendingCourses();
		} catch (error) {
			this.showError(error.message);
		}
	},

	showError(message) {
		const pendingCoursesList = document.getElementById("pendingCoursesList");
		const errorDiv = document.createElement("div");
		errorDiv.className = "error";
		errorDiv.style.cssText =
			"background: #f8d7da; color: #721c24; padding: 1rem; border-radius: 4px; margin-bottom: 1rem;";
		errorDiv.textContent = message;
		pendingCoursesList.insertBefore(errorDiv, pendingCoursesList.firstChild);
		setTimeout(() => errorDiv.remove(), 5000);
	},
};
