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
    currentDraftId: null,
    draftsList: [],
    unsavedChanges: false,
    isLoading: false,

    init() {
        this.initElements();
        this.initEventListeners();
        this.loadApplications();
        this.loadDraftsFromBackend();
        this.initFormEnhancements();
    },

    getStatusIcon(status) {
        const icons = {
            draft: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>`,
            pending: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>`,
            approved: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>`,
            rejected: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line></svg>`
        };
        return icons[status] || '';
    },

    formatRelativeTime(dateString) {
        if (!dateString) return 'Unknown';
        const date = new Date(dateString);
        const now = new Date();
        const diffMs = now - date;
        const diffSecs = Math.floor(diffMs / 1000);
        const diffMins = Math.floor(diffSecs / 60);
        const diffHours = Math.floor(diffMins / 60);
        const diffDays = Math.floor(diffHours / 24);

        if (diffSecs < 60) return 'Just now';
        if (diffMins < 60) return `${diffMins} minute${diffMins !== 1 ? 's' : ''} ago`;
        if (diffHours < 24) return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
        if (diffDays < 30) return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;

        // Fall back to formatted date if older than 30 days
        return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined });
    },

    shouldShowEditButton(status) {
        return status === 'pending' || status === 'draft';
    },

    renderSkeletonLoader(count = 3) {
        return Array(count).fill(0).map(() =>
            `<div class="skeleton-loader skeleton-card"></div>`
        ).join('');
    },

    initElements() {
        this.applicationList = document.getElementById("applicationList");
        this.editApplicationModal = document.getElementById("editApplicationModal");
        this.editApplicationForm = document.getElementById("editApplicationForm");
        this.editTitleField = document.getElementById("editTitle");
        this.editDescriptionField = document.getElementById("editDescription");
        this.editDraftStatus = document.getElementById("editDraftStatus");
    },

    initEventListeners() {
        if (this.editApplicationForm) {
            this.editApplicationForm.addEventListener("submit", (e) => this.handleEditApplication(e));
        }

        const closeButtons = document.querySelectorAll(".close");
        closeButtons.forEach((btn) => {
            btn.addEventListener("click", (e) => {
                const modal = e.target.closest(".modal");
                if (modal) {
                    this.handleModalClose(modal);
                }
            });
        });

        window.addEventListener("click", (e) => {
            if (e.target.classList.contains("modal")) {
                this.handleModalClose(e.target);
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
        // Form enhancements for edit modal only
        if (this.editTitleField) {
            this.editTitleField.addEventListener("input", () => {
                this.updateCharCounter("editTitle", 255);
                this.clearFieldError("editTitle");
            });
            this.updateCharCounter("editTitle", 255);
        }

        if (this.editDescriptionField) {
            this.editDescriptionField.addEventListener("input", () => {
                this.updateCharCounter("editDescription");
                this.clearFieldError("editDescription");
            });
            this.updateCharCounter("editDescription");
        }
    },

    async loadApplications() {
        if (this.applicationList) {
            this.isLoading = true;
            this.applicationList.innerHTML = this.renderSkeletonLoader(3);
        }
        try {
            const applications = await api.instructor.getApplications();
            this.isLoading = false;
            this.renderApplicationList(applications || []);
        } catch (error) {
            this.isLoading = false;
            if (this.applicationList) {
                this.applicationList.innerHTML = "<p>Error loading applications. Please try again.</p>";
            }
        }
    },

    async loadDraftsFromBackend() {
        const draftsList = document.getElementById("draftsList");
        if (draftsList) {
            this.isLoading = true;
            draftsList.innerHTML = this.renderSkeletonLoader(2);
        }
        try {
            const applications = await api.instructor.getApplications();
            this.draftsList = (applications || []).filter((a) => a.status === "draft");
            this.isLoading = false;
            this.renderDraftList();
        } catch (error) {
            console.error("Failed to load drafts:", error);
            this.isLoading = false;
            this.draftsList = [];
            this.renderDraftList();
        }
    },

    renderDraftList() {
        const draftsList = document.getElementById("draftsList");
        const draftCount = document.getElementById("draftCount");
        if (!draftsList) return;

        if (draftCount) {
            draftCount.textContent = `(${this.draftsList.length}/16)`;
        }

        if (!this.draftsList || this.draftsList.length === 0) {
            draftsList.innerHTML = `
                <div class="empty-state">
                    <svg class="empty-state-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                        <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                    </svg>
                    <h3>No drafts yet</h3>
                    <p>Start creating your first course application</p>
                    <button class="btn-primary" onclick="applyModule.openNewApplicationModal()">Create Draft</button>
                </div>
            `;
            return;
        }

        draftsList.innerHTML = this.draftsList
            .map(
                (draft) => {
                    const displayTitle = draft.title && draft.title.trim() ? draft.title : this.generateDraftName("", this.draftsList);
                    const lastSaved = this.formatRelativeTime(draft.updated_at || draft.created_at);
                    const lastSavedAbsolute = new Date(draft.updated_at || draft.created_at).toLocaleString();
                    return `
                <div class="my-course-card draft-card">
                    <div class="my-course-card-header">
                        <div style="flex: 1;">
                            <h3>${escapeHtml(displayTitle)}</h3>
                            <div class="my-course-date" title="${lastSavedAbsolute}">Last saved: ${lastSaved}</div>
                        </div>
                        <span class="status-badge status-draft" style="background: var(--color-accent-blue);">${this.getStatusIcon('draft')} Draft</span>
                    </div>
                    <p>${escapeHtml((draft.description || "").length > 200 ? draft.description.substring(0, 200) + "..." : draft.description || "")}</p>
                    <div class="my-course-meta">
                        <div class="my-course-actions">
                            <a href="/apply/new/?draft=${draft.id}" class="btn-secondary continue-draft-btn" style="padding: var(--spacing-sm) var(--spacing-lg); font-size: 0.875rem; text-decoration: none; display: inline-block;">Continue</a>
                            <button class="delete-btn-small delete-draft-btn" data-draft-id="${draft.id}">Delete</button>
                        </div>
                    </div>
                </div>
            `;
                },
            )
            .join("");

        // Continue buttons are now links, no event listeners needed

        document.querySelectorAll(".delete-draft-btn").forEach((btn) => {
            btn.addEventListener("click", (e) => {
                const draftId = parseInt(e.target.dataset.draftId);
                this.deleteDraft(draftId);
            });
        });
    },

    generateDraftName(title, existingDrafts) {
        if (title && title.trim().length > 0) {
            return title;
        }

        const untitledPattern = /^Untitled Course( \d+)?$/;
        const untitledDrafts = existingDrafts.filter((d) => {
            const draftTitle = d.title || "";
            return untitledPattern.test(draftTitle.trim());
        });

        if (untitledDrafts.length === 0) {
            return "Untitled Course";
        }

        const numbers = untitledDrafts
            .map((d) => {
                const match = (d.title || "").match(/^Untitled Course( (\d+))?$/);
                return match && match[2] ? parseInt(match[2]) : 1;
            })
            .sort((a, b) => a - b);

        let nextNumber = 1;
        for (const num of numbers) {
            if (num === nextNumber) {
                nextNumber++;
            } else {
                break;
            }
        }

        return nextNumber === 1 ? "Untitled Course" : `Untitled Course ${nextNumber}`;
    },

    renderApplicationList(applications) {
        if (!this.applicationList) return;

        // Filter out drafts
        const nonDraftApplications = (applications || []).filter((a) => a.status !== "draft");

        if (!nonDraftApplications || nonDraftApplications.length === 0) {
            this.applicationList.innerHTML = `
                <div class="empty-state">
                    <svg class="empty-state-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                        <polyline points="14 2 14 8 20 8"></polyline>
                        <line x1="16" y1="13" x2="8" y2="13"></line>
                        <line x1="16" y1="17" x2="8" y2="17"></line>
                        <polyline points="10 9 9 9 8 9"></polyline>
                    </svg>
                    <h3>No applications submitted</h3>
                    <p>Submit your first course for review</p>
                    <button class="btn-primary" onclick="applyModule.openNewApplicationModal()">New Application</button>
                </div>
            `;
            return;
        }

        this.applicationList.innerHTML = nonDraftApplications
            .map(
                (app) => {
                    const submittedDate = this.formatRelativeTime(app.created_at);
                    const submittedDateAbsolute = new Date(app.created_at).toLocaleString();
                    const showEdit = this.shouldShowEditButton(app.status);
                    const isReadOnly = app.status === 'rejected';
                    return `
                        <div class="my-course-card application-card status-${app.status} ${isReadOnly ? 'read-only-card' : ''}">
                            <div class="my-course-card-header">
                                <div style="flex: 1;">
                                    <h3>${escapeHtml(app.title)}</h3>
                                    <div class="my-course-date" title="${submittedDateAbsolute}">Submitted ${submittedDate}</div>
                                </div>
                                <span class="status-badge status-${app.status}">${this.getStatusIcon(app.status)} ${app.status}</span>
                            </div>
                            <p>${escapeHtml(app.description.length > 200 ? app.description.substring(0, 200) + "..." : app.description)}</p>
                            <div class="my-course-meta">
                                <div class="my-course-actions">
                                    ${showEdit ? `<button class="edit-application-btn" data-course-id="${app.id}">Edit</button>` : ''}
                                    <button class="delete-btn-small delete-application-btn" data-course-id="${app.id}">Delete</button>
                                </div>
                            </div>
                        </div>
                    `;
                },
            )
            .join("");
    },

    openNewApplicationModal() {
        if (this.newApplicationModal) {
            // Check draft limit
            if (this.draftsList.length >= 16) {
                this.showMessage("Draft limit reached (16 max). Please delete a draft to create a new one.", "error");
                return;
            }

            this.newApplicationModal.style.display = "block";
            this.editingCourseId = null;
            this.currentDraftId = null;
            if (this.newApplicationForm) this.newApplicationForm.reset();
            this.clearAllFieldErrors("new");
            this.updateCharCounter("newTitle", 255);
            this.updateCharCounter("newDescription");
            this.unsavedChanges = false;
            this.checkForDraft();
            this.initAutoSave("new");
        }
    },

    // openDraftModal and populateDraftModal removed - navigation handled by links to /apply/new/

    async openEditApplicationModal(courseId) {
        try {
            const applications = await api.instructor.getApplications();
            const application = applications.find((a) => a.id === courseId);
            if (!application) {
                this.showMessage("Application not found", "error");
                return;
            }

            if (this.editApplicationModal) {
                this.editApplicationModal.style.display = "block";
                this.editingCourseId = courseId;
                if (this.editTitleField) this.editTitleField.value = application.title;
                if (this.editDescriptionField) this.editDescriptionField.value = application.description;
                this.clearAllFieldErrors("edit");
                this.updateCharCounter("editTitle", 255);
                this.updateCharCounter("editDescription");
            }
        } catch (error) {
            this.showMessage("Failed to load application", "error");
        }
    },

    // handleNewApplication removed - handled in apply-new.js

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
            await api.instructor.updateApplication(this.editingCourseId, formData);
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
            await api.applications.delete(courseId);
            this.showMessage("Application deleted successfully", "success");
            this.loadApplications();
        } catch (error) {
            this.showMessage(error.message || "Failed to delete application", "error");
        }
    },

    async deleteDraft(draftId) {
        if (!confirm("Are you sure you want to delete this draft?")) return;

        try {
            await api.applications.delete(draftId);
            this.showMessage("Draft deleted successfully", "success");
            this.loadDraftsFromBackend();
            // Draft deleted, no need to reset anything
        } catch (error) {
            this.showMessage(error.message || "Failed to delete draft", "error");
        }
    },

    handleModalClose(modal) {
        // Only edit modal remains
        modal.style.display = "none";
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

            // Set unsaved changes flag for auto-saves
            if (!showMessage) {
                this.unsavedChanges = true;
            }

            this.updateDraftStatus(formType);

            // If manual save, also save to backend
            if (showMessage) {
                this.saveDraftToBackend();
            }
        } catch (error) {
            if (error.name === "QuotaExceededError") {
                this.showMessage("Storage quota exceeded. Please clear some space.", "error");
            } else {
                this.showMessage("Failed to save draft", "error");
            }
        }
    },

    async saveDraftToBackend() {
        if (this.isSubmitting) return;

        const formData = {
            title: this.newTitleField.value.trim(),
            description: this.newDescriptionField.value.trim(),
        };

        try {
            let draft;
            if (this.currentDraftId) {
                // Update existing draft
                draft = await api.drafts.update(this.currentDraftId, formData);
            } else {
                // Create new draft
                draft = await api.drafts.create(formData);
                this.currentDraftId = draft.id;
            }

            this.unsavedChanges = false;
            this.lastSavedState = {
                title: formData.title,
                description: formData.description,
            };

            // Update localStorage with draft ID
            const localDraft = {
                title: formData.title,
                description: formData.description,
                savedAt: Date.now(),
                draftId: this.currentDraftId,
            };
            localStorage.setItem("courseDraft", JSON.stringify(localDraft));

            this.showMessage("Draft saved successfully!", "success");
            this.updateDraftStatus("new");
            this.loadDraftsFromBackend();
        } catch (error) {
            if (error.type === "validation" && error.fields) {
                Object.entries(error.fields).forEach(([field, message]) => {
                    this.showFieldError(`new${field.charAt(0).toUpperCase() + field.slice(1)}`, message);
                });
            } else {
                this.showMessage(error.message || "Failed to save draft to server", "error");
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

        // Hide the draft prompt
        const draftPrompt = document.getElementById("newDraftPrompt");
        if (draftPrompt) {
            draftPrompt.style.display = "none";
        }
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
        if (!draft) {
            this.initAutoSave("new");
            return;
        }
        // If we have a draftId, we're editing an existing draft, don't show prompt
        if (draft.draftId && draft.draftId === this.currentDraftId) {
            this.initAutoSave("new");
            return;
        }

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
            counter.textContent = `${length} characters`;
            counter.className = "char-counter";
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

