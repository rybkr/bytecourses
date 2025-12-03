const instructorModule = {
	init() {
		const editCourseForm = document.getElementById("editCourseForm");
		const closeModal = document.getElementsByClassName("close")[0];
		const editModal = document.getElementById("editModal");

		editCourseForm.addEventListener("submit", this.handleEdit.bind(this));

		closeModal.onclick = () => (editModal.style.display = "none");
		window.onclick = (event) => {
			if (event.target == editModal) {
				editModal.style.display = "none";
			}
		};

		document.addEventListener("click", (e) => {
			if (e.target.classList.contains("edit-btn")) {
				const courseId = parseInt(e.target.dataset.courseId);
				const title = e.target.dataset.courseTitle;
				const description = e.target.dataset.courseDescription;
				let content = e.target.dataset.courseContent || "";
				content = content
					.replace(/&amp;/g, "&")
					.replace(/&lt;/g, "<")
					.replace(/&gt;/g, ">")
					.replace(/&quot;/g, '"')
					.replace(/&#39;/g, "'");
				this.openEditModal(courseId, title, description, content);
			} else if (e.target.classList.contains("delete-btn-small")) {
				const courseId = parseInt(e.target.dataset.courseId);
				this.deleteCourse(courseId);
			}
		});
	},

	async load() {
		try {
			const courses = await api.instructor.getCourses();
			this.render(courses);
		} catch (error) {
			document.getElementById("myCoursesList").innerHTML =
				"<p>Error loading your courses</p>";
		}
	},

	render(courses) {
		const myCoursesList = document.getElementById("myCoursesList");

		if (!courses || courses.length === 0) {
			myCoursesList.innerHTML = "<p>You haven't created any courses yet</p>";
			return;
		}

		myCoursesList.innerHTML = courses
			.map(
				(course) => {
					const content = course.content || "";
					const escapedContent = content
						.replace(/&/g, "&amp;")
						.replace(/</g, "&lt;")
						.replace(/>/g, "&gt;")
						.replace(/"/g, "&quot;")
						.replace(/'/g, "&#39;");
					return `
            <div class="my-course-card">
                <h3>${escapeHtml(course.title)}</h3>
                <p>${escapeHtml(course.description)}</p>
                <div class="my-course-meta">
                    <span class="status-badge status-${course.status}">${course.status}</span>
                    <div class="my-course-actions">
                        <button class="edit-btn" data-course-id="${course.id}" data-course-title="${escapeHtml(course.title).replace(/"/g, "&quot;")}" data-course-description="${escapeHtml(course.description).replace(/"/g, "&quot;")}" data-course-content="${escapedContent}">Edit</button>
                        <button class="delete-btn-small" data-course-id="${course.id}">Delete</button>
                    </div>
                </div>
            </div>
        `;
				},
			)
			.join("");
	},

	openEditModal(id, title, description, content) {
		document.getElementById("editCourseId").value = id;
		document.getElementById("editTitle").value = title;
		document.getElementById("editDescription").value = description;

		const contentTextarea = document.getElementById("editContent");
		const sectionsContainer = document.getElementById("sectionsContainer");
		const toggleJSONBtn = document.getElementById("toggleJSONViewBtn");
		let isJSONView = false;

		if (sectionsContainer && toggleJSONBtn) {
			contentEditorModule.init(content || "");
			sectionsContainer.style.display = "block";
			contentTextarea.style.display = "none";
			isJSONView = false;

			toggleJSONBtn.textContent = "Show JSON";
			toggleJSONBtn.onclick = () => {
				if (isJSONView) {
					const json = contentEditorModule.getJSON();
					contentTextarea.value = json;
					sectionsContainer.style.display = "block";
					contentTextarea.style.display = "none";
					toggleJSONBtn.textContent = "Show JSON";
					isJSONView = false;
				} else {
					const json = contentEditorModule.getJSON();
					contentTextarea.value = json;
					sectionsContainer.style.display = "none";
					contentTextarea.style.display = "block";
					toggleJSONBtn.textContent = "Show Visual Editor";
					isJSONView = true;
				}
			};

			const addSectionBtn = document.getElementById("addSectionBtn");
			if (addSectionBtn) {
				addSectionBtn.onclick = () => {
					contentEditorModule.addSection();
				};
			}
		} else {
			contentTextarea.value = content || "";
		}

		document.getElementById("editModal").style.display = "block";
	},

	async handleEdit(e) {
		e.preventDefault();

		const courseId = document.getElementById("editCourseId").value;
		const contentTextarea = document.getElementById("editContent");
		const sectionsContainer = document.getElementById("sectionsContainer");

		let contentValue = "";
		if (sectionsContainer && sectionsContainer.style.display !== "none") {
			contentValue = contentEditorModule.getJSON();
		} else {
			contentValue = contentTextarea.value.trim();
		}

		const formData = {
			title: document.getElementById("editTitle").value,
			description: document.getElementById("editDescription").value,
			content: contentValue,
		};

		if (contentValue) {
			try {
				JSON.parse(contentValue);
			} catch (jsonError) {
				this.showMessage("Invalid JSON format in content field. Please check your JSON syntax.", "error");
				return;
			}
		}

		try {
			await api.instructor.updateCourse(courseId, formData);
			document.getElementById("editModal").style.display = "none";
			this.showMessage("Course updated successfully!", "success");
			this.load();
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

	async deleteCourse(id) {
		if (!confirm("Are you sure you want to delete this course?")) {
			return;
		}

		try {
			await api.instructor.deleteCourse(id);
			this.showMessage("Course deleted successfully", "success");
			this.load();
		} catch (error) {
			this.showMessage(error.message, "error");
		}
	},

	showMessage(message, type) {
		const myCoursesMessage = document.getElementById("myCoursesMessage");
		myCoursesMessage.textContent = message;
		myCoursesMessage.className = type;
		setTimeout(() => {
			myCoursesMessage.style.display = "none";
		}, 3000);
	},
};
