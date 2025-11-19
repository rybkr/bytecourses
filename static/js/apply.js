function escapeHtml(text) {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
}

const applyModule = {
    lastSavedState: null,
    autoSaveInterval: null,
    autoSaveDebounce: null,
    isSubmitting: false,
    editingCourseId: null,

    init() {
        this.initElements();
        this.initEventListeners();
        this.loadApplications();
        this.initFormEnhancements();
        this.checkForDraft();
    },

    initElements() {
        this.applicationList = document.getElementById("applicationList");
        this.newApplicationBtn = document.getElementById("newApplicationBtn");
        this.newApplicationModal = document.getElementById("newApplicationModal");
        this.editApplicationModal = document.getElementById("editApplicationModal");
        this.newApplicationForm = document.getElementById("newApplicationForm");
        this.editApplicationForm = document.getElementById("editApplicationForm");
        this.newTitleField = document.getElementById("newTitle");
        this.newDescriptionField = document.getElementById("newDescription");
        this.editTitleField = document.getElementById("editTitle");
        this.editDescriptionField = document.getElementById("editDescription");
        this.newDraftStatus = document.getElementById("newDraftStatus");
        this.editDraftStatus = document.getElementById("editDraftStatus");
    },

    initEventListeners() {
        if (this.newApplicationBtn) {
            this.newApplicationBtn.addEventListener("click", () => this.openNewApplicationModal());
        }

        if (this.newApplicationForm) {
            this.newApplicationForm.addEventListener("submit", (e) => this.handleNewApplication(e));
        }

        if (this.editApplicationForm) {
            this.editApplicationForm.addEventListener("submit", (e) => this.handleEditApplication(e));
        }

        const loadNewDraftBtn = document.getElementById("loadNewDraftBtn");
        const discardNewDraftBtn = document.getElementById("discardNewDraftBtn");
        const saveNewDraftBtn = document.getElementById("saveNewDraftBtn");

        if (loadNewDraftBtn) {
            loadNewDraftBtn.addEventListener("click", () => this.loadDraft());
        }

        if (discardNewDraftBtn) {
            discardNewDraftBtn.addEventListener("click", () => this.discardDraft());
        }

        if (saveNewDraftBtn) {
            saveNewDraftBtn.addEventListener("click", () => this.saveDraft("new", true));
        }

        const closeButtons = document.querySelectorAll(".close");
        closeButtons.forEach((btn) => {
            btn.addEventListener("click", (e) => {
                const modal = e.target.closest(".modal");
                if (modal) modal.style.display = "none";
            });
        });

        window.addEventListener("click", (e) => {
            if (e.target.classList.contains("modal")) {
                e.target.style.display = "none";
            }
        });

        document.addEventListener("click", (e) => {
            if (e.target.classList.contains("edit-application-btn")) {
                const courseId = parseInt(e.target.dataset.courseId);
                this.openEditApplicationModal(courseId);
            } else if (e.target.classList.contains("delete-application-btn")) {
                const courseId = parseInt(e.target.dataset.courseId);
                this.deleteApplication(courseId);
            }
        });
    },

    initFormEnhancements() {
        if (this.newTitleField) {
            this.newTitleField.addEventListener("input", () => {
                this.updateCharCounter("newTitle", 255);
                this.clearFieldError("newTitle");
                this.scheduleAutoSave("new");
                this.updateDraftStatus("new");
            });
            this.updateCharCounter("newTitle", 255);
        }

        if (this.newDescriptionField) {
            this.newDescriptionField.addEventListener("input", () => {
                this.updateCharCounter("newDescription");
                this.clearFieldError("newDescription");
                this.scheduleAutoSave("new");
                this.updateDraftStatus("new");
            });
            this.updateCharCounter("newDescription");
        }

        if (this.editTitleField) {
            this.editTitleField.addEventListener("input", () => {
                this.updateCharCounter("editTitle", 255);
                this.clearFieldError("editTitle");
            });
        }

        if (this.editDescriptionField) {
            this.editDescriptionField.addEventListener("input", () => {
                this.updateCharCounter("editDescription");
                this.clearFieldError("editDescription");
            });
        }
    },

    async loadApplications() {
        try {
            const courses = await api.instructor.getCourses();
            this.renderApplicationList(courses);
        } catch (error) {
            if (this.applicationList) {
                this.applicationList.innerHTML = "<p>Error loading applications. Please try again.</p>";
            }
        }
    },

    renderApplicationList(courses) {
        if (!this.applicationList) return;

        if (!courses || courses.length === 0) {
            this.applicationList.innerHTML = `
				<div style="text-align: center; padding: var(--spacing-3xl); color: var(--color-text-light);">
					<p style="margin-bottom: var(--spacing-lg);">You haven't submitted any applications yet.</p>
					<button id="emptyStateNewBtn" class="btn-primary">Create Your First Application</button>
				</div>
			`;
            const emptyStateBtn = document.getElementById("emptyStateNewBtn");
            if (emptyStateBtn) {
                emptyStateBtn.addEventListener("click", () => this.openNewApplicationModal());
            }
            return;
        }

        this.applicationList.innerHTML = courses
            .map(
                (course) => `
			<div class="my-course-card">
				<h3>${escapeHtml(course.title)}</h3>
				<p>${escapeHtml(course.description.length > 200 ? course.description.substring(0, 200) + "..." : course.description)}</p>
				<div class="my-course-meta">
					<span class="status-badge status-${course.status}">${course.status}</span>
					<div class="my-course-actions">
						<button class="edit-application-btn" data-course-id="${course.id}">Edit</button>
						<button class="delete-btn-small delete-application-btn" data-course-id="${course.id}">Delete</button>
					</div>
				</div>
			</div>
		`,
            )
            .join("");
    },

    openNewApplicationModal() {
        if (this.newApplicationModal) {
            this.newApplicationModal.style.display = "block";
            this.editingCourseId = null;
            if (this.newApplicationForm) this.newApplicationForm.reset();
            this.clearAllFieldErrors("new");
            this.updateCharCounter("newTitle", 255);
            this.updateCharCounter("newDescription");
            this.checkForDraft();
            this.initAutoSave("new");
        }
    },

    async openEditApplicationModal(courseId) {
        try {
            const courses = await api.instructor.getCourses();
            const course = courses.find((c) => c.id === courseId);
            if (!course) {
                this.showMessage("Course not found", "error");
                return;
            }

            if (this.editApplicationModal) {
                this.editApplicationModal.style.display = "block";
                this.editingCourseId = courseId;
                if (this.editTitleField) this.editTitleField.value = course.title;
                if (this.editDescriptionField) this.editDescriptionField.value = course.description;
                this.clearAllFieldErrors("edit");
                this.updateCharCounter("editTitle", 255);
                this.updateCharCounter("editDescription");
            }
        } catch (error) {
            this.showMessage("Failed to load course", "error");
        }
    },

    async handleNewApplication(e) {
        e.preventDefault();
        this.clearAllFieldErrors("new");
        this.isSubmitting = true;
        this.stopAutoSave("new");

        const formData = {
            title: this.newTitleField.value.trim(),
            description: this.newDescriptionField.value.trim(),
        };

        try {
            await api.courses.create(formData);
            this.showMessage("Application submitted successfully!", "success");
            if (this.newApplicationModal) this.newApplicationModal.style.display = "none";
            if (this.newApplicationForm) this.newApplicationForm.reset();
            this.clearDraft();
            this.lastSavedState = null;
            this.loadApplications();
        } catch (error) {
            if (error.type === "validation" && error.fields) {
                Object.entries(error.fields).forEach(([field, message]) => {
                    this.showFieldError(`new${field.charAt(0).toUpperCase() + field.slice(1)}`, message);
                });
            } else {
                this.showMessage(error.message || "Failed to submit application", "error");
            }
        } finally {
            this.isSubmitting = false;
            this.initAutoSave("new");
        }
    },

    async handleEditApplication(e) {
        e.preventDefault();
        if (!this.editingCourseId) return;

        this.clearAllFieldErrors("edit");
        this.isSubmitting = true;

        const formData = {
            title: this.editTitleField.value.trim(),
            description: this.editDescriptionField.value.trim(),
        };

        try {
            await api.instructor.updateCourse(this.editingCourseId, formData);
            this.showMessage("Application updated successfully!", "success");
            if (this.editApplicationModal) this.editApplicationModal.style.display = "none";
            this.editingCourseId = null;
            this.loadApplications();
        } catch (error) {
            if (error.type === "validation" && error.fields) {
                Object.entries(error.fields).forEach(([field, message]) => {
                    this.showFieldError(`edit${field.charAt(0).toUpperCase() + field.slice(1)}`, message);
                });
            } else {
                this.showMessage(error.message || "Failed to update application", "error");
            }
        } finally {
            this.isSubmitting = false;
        }
    },

    async deleteApplication(courseId) {
        if (!confirm("Are you sure you want to delete this application?")) return;

        try {
            await api.instructor.deleteCourse(courseId);
            this.showMessage("Application deleted successfully", "success");
            this.loadApplications();
        } catch (error) {
            this.showMessage(error.message || "Failed to delete application", "error");
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

    hasUnsavedChanges(formType) {
        const titleField = formType === "new" ? this.newTitleField : this.editTitleField;
        const descriptionField = formType === "new" ? this.newDescriptionField : this.editDescriptionField;

        if (!titleField || !descriptionField) return false;

        const currentState = {
            title: titleField.value,
            description: descriptionField.value,
        };

        if (!this.lastSavedState) {
            return currentState.title.length > 0 || currentState.description.length > 0;
        }

        return (
            currentState.title !== this.lastSavedState.title ||
            currentState.description !== this.lastSavedState.description
        );
    },

    saveDraft(formType = "new", showMessage = false) {
        if (this.isSubmitting) return;

        const titleField = formType === "new" ? this.newTitleField : this.editTitleField;
        const descriptionField = formType === "new" ? this.newDescriptionField : this.editDescriptionField;

        if (!titleField || !descriptionField) return;

        const formData = {
            title: titleField.value,
            description: descriptionField.value,
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

            this.updateDraftStatus(formType);
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

        if (this.newTitleField) this.newTitleField.value = draft.title || "";
        if (this.newDescriptionField) this.newDescriptionField.value = draft.description || "";

        this.updateCharCounter("newTitle", 255);
        this.updateCharCounter("newDescription");

        this.lastSavedState = {
            title: draft.title || "",
            description: draft.description || "",
        };

        this.updateDraftStatus("new");
        this.initAutoSave("new");
    },

    discardDraft() {
        this.clearDraft();
        this.lastSavedState = null;
        this.updateDraftStatus("new");
        this.initAutoSave("new");
    },

    clearDraft() {
        try {
            localStorage.removeItem("courseDraft");
            this.lastSavedState = null;
            this.updateDraftStatus("new");
        } catch (error) {
            console.error("Failed to clear draft:", error);
        }
    },

    checkForDraft() {
        const draft = this.getDraft();
        if (!draft || !draft.title) return;

        const draftPrompt = document.getElementById("newDraftPrompt");
        if (!draftPrompt) return;

        const preview = document.getElementById("newDraftPreview");
        if (preview) {
            preview.innerHTML = `
				<p><strong>Title:</strong> ${escapeHtml(draft.title || "(empty)")}</p>
				<p><strong>Description:</strong> ${escapeHtml((draft.description || "").substring(0, 100))}${draft.description && draft.description.length > 100 ? "..." : ""}</p>
			`;
        }

        draftPrompt.style.display = "block";
    },

    scheduleAutoSave(formType = "new") {
        if (this.autoSaveDebounce) {
            clearTimeout(this.autoSaveDebounce);
        }

        this.autoSaveDebounce = setTimeout(() => {
            if (this.hasUnsavedChanges(formType)) {
                this.saveDraft(formType);
            }
        }, 2000);
    },

    initAutoSave(formType = "new") {
        this.stopAutoSave(formType);
        this.autoSaveInterval = setInterval(() => {
            if (this.hasUnsavedChanges(formType) && !this.isSubmitting) {
                this.saveDraft(formType);
            }
        }, 30000);
    },

    stopAutoSave(formType = "new") {
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
        const date = new Date(timestamp);
        return date.toLocaleString();
    },

    updateDraftStatus(formType = "new") {
        const draftStatus = formType === "new" ? this.newDraftStatus : this.editDraftStatus;
        if (!draftStatus) return;

        const draft = this.getDraft();
        if (draft && draft.savedAt) {
            draftStatus.style.display = "block";
            draftStatus.innerHTML = `<span style="color: var(--color-success);">Draft saved at ${this.formatTimestamp(draft.savedAt)}</span>`;
        } else {
            draftStatus.style.display = "none";
        }
    },

    updateCharCounter(fieldId, maxLength = null) {
        const field = document.getElementById(fieldId);
        const counter = document.getElementById(`${fieldId}Counter`);
        if (!field || !counter) return;

        const length = field.value.length;
        if (maxLength) {
            counter.textContent = `${length} / ${maxLength} characters`;
            counter.className = "char-counter";
            if (length > maxLength * 0.9) {
                counter.classList.add("warning");
            }
            if (length > maxLength) {
                counter.classList.add("error");
            }
        } else {
            counter.textContent = `${length} characters (minimum 10)`;
            counter.className = "char-counter";
            if (length < 10) {
                counter.classList.add("error");
            } else if (length < 20) {
                counter.classList.add("warning");
            }
        }
    },

    showFieldError(fieldId, message) {
        const errorDiv = document.getElementById(`${fieldId}Error`);
        if (errorDiv) {
            errorDiv.textContent = message;
            errorDiv.classList.add("show");
        }
    },

    clearFieldError(fieldId) {
        const errorDiv = document.getElementById(`${fieldId}Error`);
        if (errorDiv) {
            errorDiv.textContent = "";
            errorDiv.classList.remove("show");
        }
    },

    clearAllFieldErrors(formType) {
        const prefixes = formType === "new" ? ["newTitle", "newDescription"] : ["editTitle", "editDescription"];
        prefixes.forEach((prefix) => {
            this.clearFieldError(prefix);
        });
    },

    showMessage(message, type) {
        const messageDiv = document.getElementById("portalMessage");
        if (!messageDiv) return;

        messageDiv.textContent = message;
        messageDiv.className = `message ${type}`;
        messageDiv.style.display = "block";

        setTimeout(() => {
            messageDiv.style.display = "none";
        }, 5000);
    },
};

