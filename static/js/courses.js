const coursesModule = {
	init() {
		const courseForm = document.getElementById("courseForm");
		const statusFilter = document.getElementById("statusFilter");

		courseForm.addEventListener("submit", this.handleSubmit.bind(this));
		statusFilter.addEventListener("change", this.load.bind(this));
	},

	async load() {
		const status = document.getElementById("statusFilter").value;
		const coursesList = document.getElementById("coursesList");

		try {
			const courses = await api.courses.list(status);
			this.render(courses);
		} catch (error) {
			coursesList.innerHTML = "<p>Error loading courses</p>";
		}
	},

	render(courses) {
		const coursesList = document.getElementById("coursesList");

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
                <div class="instructor-info">
                    <span>By: <span class="instructor-name">${escapeHtml(course.instructor_name || course.instructor_email || "Unknown")}</span></span>
                </div>
                <div class="course-meta">
                    <span class="status-badge status-${course.status}">${course.status}</span>
                </div>
            </div>
        `,
			)
			.join("");
	},

	async handleSubmit(e) {
		e.preventDefault();

		const formData = {
			title: document.getElementById("title").value,
			description: document.getElementById("description").value,
		};

		try {
			await api.courses.create(formData);
			this.showMessage(
				"Course submitted successfully! Awaiting approval.",
				"success",
			);
			document.getElementById("courseForm").reset();
		} catch (error) {
			if (error.type === "validation" && error.fields) {
				const fieldMessages = Object.entries(error.fields)
					.map(([field, msg]) => `${field}: ${msg}`)
					.join(", ");
				this.showMessage(fieldMessages, "error");
			} else {
				this.showMessage(error.message, "error");
			}
		}
	},

	showMessage(message, type) {
		const formMessage = document.getElementById("formMessage");
		formMessage.textContent = message;
		formMessage.className = type;
		setTimeout(() => {
			formMessage.style.display = "none";
		}, 5000);
	},
};
