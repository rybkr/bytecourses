const courseViewerModule = {
	init() {
		const courseId = this.extractCourseId();
		if (!courseId) {
			this.showError("Invalid course ID");
			return;
		}
		this.loadCourse(courseId);
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
};

