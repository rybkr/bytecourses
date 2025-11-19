function escapeHtml(text) {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
}

const applyNewModule = {
    lastSavedState: null,
    autoSaveInterval: null,
    autoSaveDebounce: null,
    isSubmitting: false,
    currentDraftId: null,
    unsavedChanges: false,

    init() {
        this.initElements();
        this.initEventListeners();
        this.initFormEnhancements();
        this.checkForDraft();
    },

    initElements() {
        this.form = document.getElementById("applicationForm");
        this.titleField = document.getElementById("title");
        this.descriptionField = document.getElementById("description");
        this.learningObjectivesField = document.getElementById("learningObjectives");
        this.prerequisitesField = document.getElementById("prerequisites");
        this.courseFormatField = document.getElementById("courseFormat");
        this.categoryTagsField = document.getElementById("categoryTags");
        this.skillLevelField = document.getElementById("skillLevel");
        this.courseDurationField = document.getElementById("courseDuration");
        this.instructorQualificationsField = document.getElementById("instructorQualifications");
        this.draftStatus = document.getElementById("draftStatus");
        this.draftPrompt = document.getElementById("draftPrompt");
        this.saveDraftBtn = document.getElementById("saveDraftBtn");
        this.loadDraftBtn = document.getElementById("loadDraftBtn");
        this.discardDraftBtn = document.getElementById("discardDraftBtn");
    },

    initEventListeners() {
        if (this.form) {
            this.form.addEventListener("submit", (e) => this.handleSubmit(e));
        }

        if (this.saveDraftBtn) {
            this.saveDraftBtn.addEventListener("click", () => this.saveDraft(true));
        }

        if (this.loadDraftBtn) {
            this.loadDraftBtn.addEventListener("click", () => this.loadDraft());
        }

        if (this.discardDraftBtn) {
            this.discardDraftBtn.addEventListener("click", () => this.discardDraft());
        }

        // Warn before leaving with unsaved changes
        window.addEventListener("beforeunload", (e) => {
            if (this.unsavedChanges) {
                e.preventDefault();
                e.returnValue = "";
            }
        });
    },

    initFormEnhancements() {
        const fields = [
            { id: "title", maxLength: 255 },
            { id: "description" },
            { id: "learningObjectives" },
            { id: "prerequisites" },
            { id: "categoryTags", maxLength: 200 },
            { id: "courseDuration", maxLength: 100 },
            { id: "instructorQualifications" }
        ];

        fields.forEach(field => {
            const element = document.getElementById(field.id);
            if (element) {
                element.addEventListener("input", () => {
                    this.updateCharCounter(field.id, field.maxLength);
                    this.clearFieldError(field.id);
                    this.scheduleAutoSave();
                    this.updateDraftStatus();
                });
                this.updateCharCounter(field.id, field.maxLength);
            }
        });

        // Handle select fields
        [this.courseFormatField, this.skillLevelField].forEach(field => {
            if (field) {
                field.addEventListener("change", () => {
                    this.clearFieldError(field.id);
                    this.scheduleAutoSave();
                    this.updateDraftStatus();
                });
            }
        });
    },

    async checkForDraft() {
        // Check URL for draft ID
        const urlParams = new URLSearchParams(window.location.search);
        const draftId = urlParams.get("draft");

        if (draftId) {
            try {
                const applications = await api.instructor.getApplications();
                const draft = (applications || []).find(a => a.id === parseInt(draftId) && a.status === "draft");
                if (draft) {
                    this.currentDraftId = draft.id;
                    this.loadDraftFromBackend(draft);
                    return;
                }
            } catch (error) {
                console.error("Failed to load draft from URL:", error);
            }
        }

        // Check localStorage for unsaved draft
        const draft = this.getDraft();
        if (draft && !draft.draftId) {
            this.showDraftPrompt(draft);
        } else {
            this.initAutoSave();
        }
    },

    loadDraftFromBackend(draft) {
        // Load all fields from backend draft (convert snake_case to camelCase)
        if (this.titleField) this.titleField.value = draft.title || "";
        if (this.descriptionField) this.descriptionField.value = draft.description || "";
        if (this.learningObjectivesField) this.learningObjectivesField.value = draft.learning_objectives || "";
        if (this.prerequisitesField) this.prerequisitesField.value = draft.prerequisites || "";
        if (this.courseFormatField) this.courseFormatField.value = draft.course_format || "";
        if (this.categoryTagsField) this.categoryTagsField.value = draft.category_tags || "";
        if (this.skillLevelField) this.skillLevelField.value = draft.skill_level || "";
        if (this.courseDurationField) this.courseDurationField.value = draft.course_duration || "";
        if (this.instructorQualificationsField) this.instructorQualificationsField.value = draft.instructor_qualifications || "";

        this.updateAllCharCounters();
        this.lastSavedState = this.getFormState();
        this.updateDraftStatus();
        this.initAutoSave();
    },

    showDraftPrompt(draft) {
        if (!this.draftPrompt) return;

        const preview = document.getElementById("draftPreview");
        if (preview) {
            preview.innerHTML = `
                <p><strong>Title:</strong> ${escapeHtml(draft.title || "(empty)")}</p>
                <p><strong>Description:</strong> ${escapeHtml((draft.description || "").substring(0, 100))}${draft.description && draft.description.length > 100 ? "..." : ""}</p>
            `;
        }

        this.draftPrompt.style.display = "block";
    },

    loadDraft() {
        const draft = this.getDraft();
        if (!draft) return;

        if (this.titleField) this.titleField.value = draft.title || "";
        if (this.descriptionField) this.descriptionField.value = draft.description || "";
        if (this.learningObjectivesField) this.learningObjectivesField.value = draft.learningObjectives || "";
        if (this.prerequisitesField) this.prerequisitesField.value = draft.prerequisites || "";
        if (this.courseFormatField) this.courseFormatField.value = draft.courseFormat || "";
        if (this.categoryTagsField) this.categoryTagsField.value = draft.categoryTags || "";
        if (this.skillLevelField) this.skillLevelField.value = draft.skillLevel || "";
        if (this.courseDurationField) this.courseDurationField.value = draft.courseDuration || "";
        if (this.instructorQualificationsField) this.instructorQualificationsField.value = draft.instructorQualifications || "";

        this.updateAllCharCounters();
        this.lastSavedState = this.getFormState();
        this.updateDraftStatus();
        this.initAutoSave();

        if (this.draftPrompt) {
            this.draftPrompt.style.display = "none";
        }
    },

    discardDraft() {
        this.clearDraft();
        this.lastSavedState = null;
        this.updateDraftStatus();
        this.initAutoSave();

        if (this.draftPrompt) {
            this.draftPrompt.style.display = "none";
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

    getFormState() {
        return {
            title: this.titleField?.value || "",
            description: this.descriptionField?.value || "",
            learningObjectives: this.learningObjectivesField?.value || "",
            prerequisites: this.prerequisitesField?.value || "",
            courseFormat: this.courseFormatField?.value || "",
            categoryTags: this.categoryTagsField?.value || "",
            skillLevel: this.skillLevelField?.value || "",
            courseDuration: this.courseDurationField?.value || "",
            instructorQualifications: this.instructorQualificationsField?.value || ""
        };
    },

    // Convert camelCase field names to snake_case for API calls
    convertToSnakeCase(formData) {
        const mapping = {
            learningObjectives: "learning_objectives",
            courseFormat: "course_format",
            categoryTags: "category_tags",
            skillLevel: "skill_level",
            courseDuration: "course_duration",
            instructorQualifications: "instructor_qualifications"
        };

        const apiData = {};
        for (const [key, value] of Object.entries(formData)) {
            const apiKey = mapping[key] || key;
            apiData[apiKey] = value;
        }
        return apiData;
    },

    // Convert snake_case field names back to camelCase for error display
    convertToCamelCase(fieldName) {
        const mapping = {
            learning_objectives: "learningObjectives",
            course_format: "courseFormat",
            category_tags: "categoryTags",
            skill_level: "skillLevel",
            course_duration: "courseDuration",
            instructor_qualifications: "instructorQualifications"
        };
        return mapping[fieldName] || fieldName;
    },

    hasUnsavedChanges() {
        const currentState = this.getFormState();
        if (!this.lastSavedState) {
            return Object.values(currentState).some(val => val.length > 0);
        }
        return JSON.stringify(currentState) !== JSON.stringify(this.lastSavedState);
    },

    saveDraft(showMessage = false) {
        if (this.isSubmitting) return;

        const formData = this.getFormState();
        formData.savedAt = Date.now();
        if (this.currentDraftId) {
            formData.draftId = this.currentDraftId;
        }

        try {
            localStorage.setItem("courseDraft", JSON.stringify(formData));
            this.lastSavedState = { ...formData };
            delete this.lastSavedState.savedAt;

            if (!showMessage) {
                this.unsavedChanges = true;
            }

            this.updateDraftStatus();

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

        const formData = this.getFormState();
        // Trim all string values
        Object.keys(formData).forEach(key => {
            if (typeof formData[key] === 'string') {
                formData[key] = formData[key].trim();
            }
        });
        const apiData = this.convertToSnakeCase(formData);

        try {
            let draft;
            if (this.currentDraftId) {
                draft = await api.drafts.update(this.currentDraftId, apiData);
            } else {
                draft = await api.drafts.create(apiData);
                this.currentDraftId = draft.id;
            }

            this.unsavedChanges = false;
            const fullFormData = this.getFormState();
            fullFormData.savedAt = Date.now();
            fullFormData.draftId = this.currentDraftId;
            localStorage.setItem("courseDraft", JSON.stringify(fullFormData));
            this.lastSavedState = { ...fullFormData };
            delete this.lastSavedState.savedAt;
            delete this.lastSavedState.draftId;

            this.showMessage("Draft saved successfully!", "success");
            this.updateDraftStatus();
        } catch (error) {
            if (error.type === "validation" && error.fields) {
                Object.entries(error.fields).forEach(([field, message]) => {
                    // Convert snake_case field names back to camelCase for error display
                    const camelCaseField = this.convertToCamelCase(field);
                    this.showFieldError(camelCaseField, message);
                });
            } else {
                this.showMessage(error.message || "Failed to save draft to server", "error");
            }
        }
    },

    async handleSubmit(e) {
        e.preventDefault();
        this.clearAllFieldErrors();
        this.isSubmitting = true;
        this.stopAutoSave();

        const formData = this.getFormState();
        // Trim all string values
        Object.keys(formData).forEach(key => {
            if (typeof formData[key] === 'string') {
                formData[key] = formData[key].trim();
            }
        });
        const apiData = this.convertToSnakeCase(formData);

        try {
            if (this.currentDraftId) {
                await api.drafts.submit(this.currentDraftId, apiData);
            } else {
                await api.applications.create({ ...apiData, status: "pending" });
            }

            this.showMessage("Application submitted successfully!", "success");
            this.clearDraft();
            this.lastSavedState = null;
            this.currentDraftId = null;
            this.unsavedChanges = false;

            // Redirect to portal after short delay
            setTimeout(() => {
                window.location.href = "/apply/";
            }, 1500);
        } catch (error) {
            if (error.type === "validation" && error.fields) {
                Object.entries(error.fields).forEach(([field, message]) => {
                    // Convert snake_case field names back to camelCase for error display
                    const camelCaseField = this.convertToCamelCase(field);
                    this.showFieldError(camelCaseField, message);
                });
            } else {
                this.showMessage(error.message || "Failed to submit application", "error");
            }
        } finally {
            this.isSubmitting = false;
            this.initAutoSave();
        }
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

    scheduleAutoSave() {
        if (this.autoSaveDebounce) {
            clearTimeout(this.autoSaveDebounce);
        }

        this.autoSaveDebounce = setTimeout(() => {
            if (this.hasUnsavedChanges()) {
                this.saveDraft();
            }
        }, 2000);
    },

    initAutoSave() {
        this.stopAutoSave();
        this.autoSaveInterval = setInterval(() => {
            if (this.hasUnsavedChanges()) {
                this.saveDraft();
            }
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

    updateDraftStatus() {
        if (!this.draftStatus) return;

        const draft = this.getDraft();
        if (draft && draft.savedAt) {
            const savedTime = new Date(draft.savedAt);
            const timeStr = savedTime.toLocaleTimeString();
            this.draftStatus.innerHTML = `<span style="color: var(--color-success);">✓ Draft saved at ${timeStr}</span>`;
            this.draftStatus.style.display = "block";
        } else {
            this.draftStatus.style.display = "none";
        }
    },

    updateAllCharCounters() {
        this.updateCharCounter("title", 255);
        this.updateCharCounter("description");
        this.updateCharCounter("learningObjectives");
        this.updateCharCounter("prerequisites");
        this.updateCharCounter("categoryTags", 200);
        this.updateCharCounter("courseDuration", 100);
        this.updateCharCounter("instructorQualifications");
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
            const field = document.getElementById(fieldId);
            if (field) {
                field.classList.add("error");
            }
        }
    },

    clearFieldError(fieldId) {
        const errorDiv = document.getElementById(`${fieldId}Error`);
        if (errorDiv) {
            errorDiv.textContent = "";
            errorDiv.classList.remove("show");
            const field = document.getElementById(fieldId);
            if (field) {
                field.classList.remove("error");
            }
        }
    },

    clearAllFieldErrors() {
        const fieldIds = ["title", "description", "learningObjectives", "prerequisites", "courseFormat", "categoryTags", "skillLevel", "courseDuration", "instructorQualifications"];
        fieldIds.forEach(id => this.clearFieldError(id));
    },

    showMessage(message, type = "info") {
        const messageDiv = document.getElementById("formMessage");
        if (!messageDiv) return;

        messageDiv.textContent = message;
        messageDiv.className = `message ${type}`;
        messageDiv.style.display = "block";

        if (type === "success" || type === "error") {
            setTimeout(() => {
                messageDiv.style.display = "none";
            }, 5000);
        }
    }
};

