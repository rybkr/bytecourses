const coursesModule = {
	lastSavedState: null,
	autoSaveInterval: null,
	autoSaveDebounce: null,
	isSubmitting: false,

	init() {
		const courseForm = document.getElementById("courseForm");
		const statusFilter = document.getElementById("statusFilter");

		courseForm.addEventListener("submit", this.handleSubmit.bind(this));
		statusFilter.addEventListener("change", this.load.bind(this));

		this.initFormEnhancements();
		this.checkForDraft();
	},

	initFormEnhancements() {
		const titleField = document.getElementById("title");
		const descriptionField = document.getElementById("description");
		const saveDraftBtn = document.getElementById("saveDraftBtn");
		const loadDraftBtn = document.getElementById("loadDraftBtn");
		const discardDraftBtn = document.getElementById("discardDraftBtn");

		if (titleField) {
			titleField.addEventListener("input", () => {
				this.updateCharCounter("title", 255);
				this.clearFieldError("title");
				this.scheduleAutoSave();
				this.updateDraftStatus();
			});
			this.updateCharCounter("title", 255);
		}

		if (descriptionField) {
			descriptionField.addEventListener("input", () => {
				this.updateCharCounter("description");
				this.clearFieldError("description");
				this.scheduleAutoSave();
				this.updateDraftStatus();
			});
			this.updateCharCounter("description");
		}

		if (saveDraftBtn) {
			saveDraftBtn.addEventListener("click", () => this.saveDraft(true));
		}

		if (loadDraftBtn) {
			loadDraftBtn.addEventListener("click", () => this.loadDraft());
		}

		if (discardDraftBtn) {
			discardDraftBtn.addEventListener("click", () => this.discardDraft());
		}
	},

	getDraft() {
		try {
			const draft = localStorage.getItem("courseDraft");
			return draft ? JSON.parse(draft) : null;
		} catch (error) {
			return null;
		}
	},

	hasUnsavedChanges() {
		const currentState = {
			title: document.getElementById("title").value,
			description: document.getElementById("description").value,
		};

		if (!this.lastSavedState) {
			return currentState.title.length > 0 || currentState.description.length > 0;
		}

		return (
			currentState.title !== this.lastSavedState.title ||
			currentState.description !== this.lastSavedState.description
		);
	},

	saveDraft(showMessage = false) {
		if (this.isSubmitting) return;

		const formData = {
			title: document.getElementById("title").value,
			description: document.getElementById("description").value,
			savedAt: Date.now(),
		};

		try {
			localStorage.setItem("courseDraft", JSON.stringify(formData));
			this.lastSavedState = {
				title: formData.title,
				description: formData.description,
			};

			if (showMessage) {
				this.showMessage("Draft saved successfully!", "success");
			}

			this.updateDraftStatus();
		} catch (error) {
			if (error.name === "QuotaExceededError") {
				this.showMessage("Storage quota exceeded. Please clear some space.", "error");
			} else {
				this.showMessage("Failed to save draft", "error");
			}
		}
	},

	loadDraft() {
		const draft = this.getDraft();
		if (!draft) return;

		const titleField = document.getElementById("title");
		const descriptionField = document.getElementById("description");

		if (titleField) titleField.value = draft.title || "";
		if (descriptionField) descriptionField.value = draft.description || "";

		this.updateCharCounter("title", 255);
		this.updateCharCounter("description");

		this.lastSavedState = {
			title: draft.title || "",
			description: draft.description || "",
		};

		this.hideDraftPrompt();
		this.updateDraftStatus();
		this.initAutoSave();
	},

	discardDraft() {
		this.clearDraft();
		this.hideDraftPrompt();
		this.lastSavedState = null;
		this.initAutoSave();
	},

	clearDraft() {
		try {
			localStorage.removeItem("courseDraft");
			this.lastSavedState = null;
			this.updateDraftStatus();
		} catch (error) {
			console.error("Failed to clear draft:", error);
		}
	},

	checkForDraft() {
		const draft = this.getDraft();
		if (!draft) {
			this.initAutoSave();
			return;
		}

		const hasContent = (draft.title && draft.title.trim()) || (draft.description && draft.description.trim());
		if (hasContent) {
			this.showDraftPrompt(draft);
		} else {
			this.clearDraft();
			this.initAutoSave();
		}
	},

	showDraftPrompt(draft) {
		const prompt = document.getElementById("draftPrompt");
		const preview = document.getElementById("draftPreview");

		if (prompt) {
			prompt.style.display = "block";
		}

		if (preview && draft.title) {
			preview.innerHTML = `<em>"${escapeHtml(draft.title)}"</em>`;
		}
	},

	hideDraftPrompt() {
		const prompt = document.getElementById("draftPrompt");
		if (prompt) {
			prompt.style.display = "none";
		}
	},

	scheduleAutoSave() {
		if (this.autoSaveDebounce) {
			clearTimeout(this.autoSaveDebounce);
		}

		this.autoSaveDebounce = setTimeout(() => {
			if (this.hasUnsavedChanges() && !this.isSubmitting) {
				this.saveDraft(false);
			}
		}, 2000);
	},

	initAutoSave() {
		this.stopAutoSave();

		this.autoSaveInterval = setInterval(() => {
			if (this.hasUnsavedChanges() && !this.isSubmitting) {
				this.saveDraft(false);
			}
			this.updateDraftStatus();
		}, 30000);
	},

	stopAutoSave() {
		if (this.autoSaveInterval) {
			clearInterval(this.autoSaveInterval);
			this.autoSaveInterval = null;
		}
		if (this.autoSaveDebounce) {
			clearTimeout(this.autoSaveDebounce);
			this.autoSaveDebounce = null;
		}
	},

	formatTimestamp(timestamp) {
		if (!timestamp) return "";

		const now = Date.now();
		const diff = now - timestamp;
		const seconds = Math.floor(diff / 1000);
		const minutes = Math.floor(seconds / 60);
		const hours = Math.floor(minutes / 60);

		if (seconds < 10) return "Just now";
		if (seconds < 60) return `${seconds} seconds ago`;
		if (minutes < 60) return `${minutes} minute${minutes > 1 ? "s" : ""} ago`;
		if (hours < 24) return `${hours} hour${hours > 1 ? "s" : ""} ago`;

		const date = new Date(timestamp);
		return `Saved at ${date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}`;
	},

	updateDraftStatus() {
		const statusDiv = document.getElementById("draftStatus");
		if (!statusDiv) return;

		const draft = this.getDraft();
		if (!draft || !draft.savedAt) {
			statusDiv.style.display = "none";
			return;
		}

		const hasChanges = this.hasUnsavedChanges();
		const timestamp = this.formatTimestamp(draft.savedAt);

		if (hasChanges) {
			statusDiv.innerHTML = `<span class="draft-status-unsaved">Unsaved changes</span> • ${timestamp}`;
			statusDiv.className = "draft-status draft-status-warning";
		} else {
			statusDiv.innerHTML = `<span class="draft-status-saved">✓ Saved</span> • ${timestamp}`;
			statusDiv.className = "draft-status draft-status-success";
		}

		statusDiv.style.display = "block";
	},

	updateCharCounter(fieldId, maxLength = null) {
		const field = document.getElementById(fieldId);
		const counter = document.getElementById(`${fieldId}Counter`);
		if (!field || !counter) return;

		const length = field.value.length;

		if (maxLength) {
			const remaining = maxLength - length;
			const percentage = (length / maxLength) * 100;

			counter.textContent = `${length} / ${maxLength} characters`;
			counter.classList.remove("warning", "error");

			if (percentage >= 90) {
				counter.classList.add("error");
			} else if (percentage >= 80) {
				counter.classList.add("warning");
			}
		} else {
			counter.textContent = `${length} characters`;
			counter.classList.remove("warning", "error");
		}
	},

	showFieldError(fieldId, message) {
		const field = document.getElementById(fieldId);
		const errorDiv = document.getElementById(`${fieldId}Error`);

		if (field) {
			field.classList.add("error");
			field.classList.remove("success");
		}

		if (errorDiv) {
			errorDiv.textContent = message;
			errorDiv.classList.add("show");
		}
	},

	clearFieldError(fieldId) {
		const field = document.getElementById(fieldId);
		const errorDiv = document.getElementById(`${fieldId}Error`);

		if (field) {
			field.classList.remove("error", "success");
		}

		if (errorDiv) {
			errorDiv.classList.remove("show");
			errorDiv.textContent = "";
		}
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

		this.clearAllFieldErrors();
		this.isSubmitting = true;
		this.stopAutoSave();

		const submitBtn = document.getElementById("submitBtn");
		const originalText = submitBtn ? submitBtn.textContent : "Submit for Approval";

		if (submitBtn) {
			submitBtn.disabled = true;
			submitBtn.textContent = "Submitting...";
		}

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
			this.updateCharCounter("title", 255);
			this.updateCharCounter("description");
			this.clearDraft();
			this.lastSavedState = null;
		} catch (error) {
			if (error.type === "validation" && error.fields) {
				Object.entries(error.fields).forEach(([field, msg]) => {
					this.showFieldError(field, msg);
				});
				this.showMessage("Please fix the errors below", "error");
			} else {
				this.showMessage(error.message, "error");
			}
		} finally {
			this.isSubmitting = false;
			if (submitBtn) {
				submitBtn.disabled = false;
				submitBtn.textContent = originalText;
			}
			this.initAutoSave();
		}
	},

	clearAllFieldErrors() {
		this.clearFieldError("title");
		this.clearFieldError("description");
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
