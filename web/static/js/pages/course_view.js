import api from "../core/api.js";
import FormHandler from "../components/FormHandler.js";
import { escapeHtml, confirmAction, showError, hideError } from "../core/utils.js";
import { $, on } from "../core/dom.js";

document.addEventListener("DOMContentLoaded", () => {
    const pathMatch = window.location.pathname.match(/^\/courses\/(\d+)/);
    const courseId = pathMatch ? Number(pathMatch[1]) : null;
    if (!courseId || !Number.isFinite(courseId)) return;

    const editToggleBtn = $("#edit-toggle-btn");
    const courseForm = $("#course-form");
    const viewModeElements = document.querySelectorAll(".view-mode");
    const editModeElements = document.querySelectorAll(".edit-mode");
    const courseViewMain = $(".course-view-main");
    const errorDiv = $("#error-message");
    const saveBtn = $("#saveBtn");
    const publishBtn = $("#publishBtn");

    let isEditMode = false;
    let formHandler = null;

    if (!editToggleBtn) return;

    const fieldIds = ["title", "summary", "target_audience", "learning_objectives", "assumed_prerequisites"];

    function initializeFormHandler() {
        if (!courseForm) return null;

        courseForm.addEventListener("submit", (e) => e.preventDefault());

        return new FormHandler("#course-form", {
            apiPath: "/api/courses",
            entityId: courseId,
            autosaveDelay: 2000,
            fieldIds: fieldIds,
            errorContainer: "#error-message",
            statusContainer: "#save-status",
        });
    }

    function enterEditMode() {
        isEditMode = true;
        editToggleBtn.textContent = "Cancel";

        const viewTitle = $("#view-title");
        const titleInput = $("#title");
        const viewSummary = $("#view-summary");
        const summaryInput = $("#summary");

        if (viewTitle && titleInput) {
            viewTitle.style.display = "none";
            titleInput.style.display = "block";
            titleInput.value = viewTitle.textContent.trim();
        }

        if (viewSummary && summaryInput) {
            viewSummary.style.display = "none";
            summaryInput.style.display = "block";
            summaryInput.value = viewSummary.textContent.trim();
        }

        viewModeElements.forEach((el) => {
            if (el && el.id !== "view-title" && el.id !== "view-summary") {
                el.style.display = "none";
            }
        });
        editModeElements.forEach((el) => {
            if (el && el.id !== "title" && el.id !== "summary") {
                el.style.display = "";
            }
        });

        const formContainer = $(".form-container");
        if (formContainer) {
            formContainer.style.display = "block";
        }
        if (courseViewMain) {
            courseViewMain.style.display = "none";
        }

        if (!formHandler) {
            formHandler = initializeFormHandler();
        }
    }

    function exitEditMode() {
        isEditMode = false;
        editToggleBtn.textContent = "Edit";

        const viewTitle = $("#view-title");
        const titleInput = $("#title");
        const viewSummary = $("#view-summary");
        const summaryInput = $("#summary");

        if (viewTitle && titleInput) {
            viewTitle.style.display = "";
            titleInput.style.display = "none";
        }

        if (viewSummary && summaryInput) {
            viewSummary.style.display = "";
            summaryInput.style.display = "none";
        }

        viewModeElements.forEach((el) => {
            if (el && el.id !== "view-title" && el.id !== "view-summary") {
                el.style.display = "";
            }
        });
        editModeElements.forEach((el) => {
            if (el && el.id !== "title" && el.id !== "summary") {
                el.style.display = "none";
            }
        });

        const formContainer = $(".form-container");
        if (formContainer) {
            formContainer.style.display = "none";
        }
        if (courseViewMain) {
            courseViewMain.style.display = "";
        }

        if (formHandler) {
            formHandler.saveNow().then(() => {
                window.location.reload();
            }).catch(() => {
                if (confirm("You have unsaved changes. Reload anyway?")) {
                    window.location.reload();
                }
            });
        } else {
            window.location.reload();
        }
    }

    if (editToggleBtn) {
        editToggleBtn.addEventListener("click", () => {
            if (isEditMode) {
                exitEditMode();
            } else {
                enterEditMode();
            }
        });
    }

    if (saveBtn) {
        on(saveBtn, "click", async (e) => {
            e.preventDefault();
            if (!formHandler) {
                formHandler = initializeFormHandler();
            }
            if (formHandler) {
                try {
                    await formHandler.saveNow();
                    exitEditMode();
                } catch (err) {
                    showError("Save failed", errorDiv);
                }
            }
        });
    }

    async function publish() {
        if (formHandler) {
            await formHandler.saveNow();
        }

        hideError(errorDiv);
        if (publishBtn) publishBtn.disabled = true;

        try {
            await api.post(`/api/courses/${courseId}/publish`);
            window.location.reload();
        } catch (error) {
            showError(error.message || "Publish failed", errorDiv);
            if (publishBtn) publishBtn.disabled = false;
        }
    }

    if (publishBtn) {
        on(publishBtn, "click", (e) => {
            e.preventDefault();
            publish().catch(() => {
                showError("Publish failed", errorDiv);
                if (publishBtn) publishBtn.disabled = false;
            });
        });
    }

    document.addEventListener("visibilitychange", () => {
        if (document.visibilityState === "hidden" && isEditMode && formHandler) {
            formHandler.saveNow().catch(() => {});
        }
    });

    async function getNextModuleOrder() {
        const response = await api.get(`/api/courses/${courseId}/modules`);
        const modules = await response.json();
        if (!modules || modules.length === 0) return 0;
        const maxOrder = Math.max(...modules.map(m => m.order || 0));
        return maxOrder + 1;
    }

    async function getNextReadingOrder(moduleId) {
        const response = await api.get(`/api/modules/${moduleId}/readings`);
        const readings = await response.json();
        if (!readings || readings.length === 0) return 0;
        const maxOrder = Math.max(...readings.map(r => r.order || 0));
        return maxOrder + 1;
    }

    function openModal(modalId) {
        const modal = $(modalId);
        if (modal) {
            modal.style.display = "flex";
            document.body.style.overflow = "hidden";
        }
    }

    function closeModal(modalId) {
        const modal = $(modalId);
        if (modal) {
            modal.style.display = "none";
            document.body.style.overflow = "";
        }
    }

    function setupModalClose(modalId) {
        const modal = $(modalId);
        if (!modal) return;

        const closeBtns = modal.querySelectorAll(".modal-close, [data-modal-close]");
        closeBtns.forEach(btn => {
            on(btn, "click", () => closeModal(modalId));
        });

        const overlay = modal.querySelector(".modal-overlay");
        if (overlay) {
            on(overlay, "click", () => closeModal(modalId));
        }

        document.addEventListener("keydown", (e) => {
            if (e.key === "Escape" && modal.style.display === "flex") {
                closeModal(modalId);
            }
        });
    }

    const moduleCreateModal = $("#module-form-modal");
    const moduleEditModal = $("#module-edit-modal");

    if (moduleCreateModal) {
        setupModalClose("#module-form-modal");
    }

    if (moduleEditModal) {
        setupModalClose("#module-edit-modal");
    }

    const addModuleBtn = $("#add-module-btn");
    if (addModuleBtn) {
        on(addModuleBtn, "click", () => {
            const form = $("#module-create-form");
            const titleInput = $("#module-title");
            const descInput = $("#module-description");
            if (form && titleInput && descInput) {
                titleInput.value = "";
                descInput.value = "";
                hideError($("#module-error-message"));
                openModal("#module-form-modal");
            }
        });
    }

    const moduleCreateForm = $("#module-create-form");
    if (moduleCreateForm) {
        on(moduleCreateForm, "submit", async (e) => {
            e.preventDefault();
            const titleInput = $("#module-title");
            const descInput = $("#module-description");
            const errorDiv = $("#module-error-message");

            if (!titleInput || !titleInput.value.trim()) {
                showError("Title is required", errorDiv);
                return;
            }

            hideError(errorDiv);

            try {
                const order = await getNextModuleOrder();
                await api.post(`/api/courses/${courseId}/modules`, {
                    title: titleInput.value.trim(),
                    description: descInput.value.trim(),
                    order: order,
                });
                closeModal("#module-form-modal");
                window.location.reload();
            } catch (error) {
                showError(error.message || "Failed to create module", errorDiv);
            }
        });
    }

    document.querySelectorAll(".btn-edit-module").forEach((btn) => {
        on(btn, "click", () => {
            const moduleId = Number(btn.dataset.moduleId);
            if (!moduleId) return;

            const moduleItem = document.querySelector(`[data-module-id="${moduleId}"]`);
            if (!moduleItem) return;

            const titleEl = moduleItem.querySelector("h3");
            const descEl = moduleItem.querySelector(".module-description");

            const currentTitle = titleEl ? titleEl.textContent.trim() : "";
            const currentDesc = descEl ? descEl.textContent.trim() : "";

            const editIdInput = $("#module-edit-id");
            const editTitleInput = $("#module-edit-title");
            const editDescInput = $("#module-edit-description");

            if (editIdInput && editTitleInput && editDescInput) {
                editIdInput.value = moduleId;
                editTitleInput.value = currentTitle;
                editDescInput.value = currentDesc;
                hideError($("#module-edit-error-message"));
                openModal("#module-edit-modal");
            }
        });
    });

    const moduleEditForm = $("#module-edit-form");
    if (moduleEditForm) {
        on(moduleEditForm, "submit", async (e) => {
            e.preventDefault();
            const moduleIdInput = $("#module-edit-id");
            const titleInput = $("#module-edit-title");
            const descInput = $("#module-edit-description");
            const errorDiv = $("#module-edit-error-message");

            if (!moduleIdInput || !titleInput || !titleInput.value.trim()) {
                showError("Title is required", errorDiv);
                return;
            }

            const moduleId = Number(moduleIdInput.value);
            if (!moduleId) return;

            hideError(errorDiv);

            try {
                const moduleItem = document.querySelector(`[data-module-id="${moduleId}"]`);
                const currentOrder = moduleItem ? Number(moduleItem.dataset.order) : 0;

                await api.patch(`/api/courses/${courseId}/modules/${moduleId}`, {
                    title: titleInput.value.trim(),
                    description: descInput.value.trim(),
                    order: currentOrder,
                });
                closeModal("#module-edit-modal");
                window.location.reload();
            } catch (error) {
                showError(error.message || "Failed to update module", errorDiv);
            }
        });
    }

    document.querySelectorAll(".btn-delete-module").forEach((btn) => {
        on(btn, "click", async () => {
            const moduleId = Number(btn.dataset.moduleId);
            if (!moduleId) return;

            if (!confirm("Are you sure you want to delete this module? This will also delete all readings in this module.")) {
                return;
            }

            try {
                await api.delete(`/api/courses/${courseId}/modules/${moduleId}`);
                window.location.reload();
            } catch (error) {
                alert(error.message || "Failed to delete module");
            }
        });
    });

    const readingCreateModal = $("#reading-form-modal");
    if (readingCreateModal) {
        setupModalClose("#reading-form-modal");
    }

    document.querySelectorAll(".btn-add-reading").forEach((btn) => {
        on(btn, "click", () => {
            const moduleId = Number(btn.dataset.moduleId);
            if (!moduleId) return;

            const moduleIdInput = $("#reading-module-id");
            const titleInput = $("#reading-title");
            if (moduleIdInput && titleInput) {
                moduleIdInput.value = moduleId;
                titleInput.value = "";
                hideError($("#reading-error-message"));
                openModal("#reading-form-modal");
            }
        });
    });

    const readingCreateForm = $("#reading-create-form");
    if (readingCreateForm) {
        on(readingCreateForm, "submit", async (e) => {
            e.preventDefault();
            const moduleIdInput = $("#reading-module-id");
            const titleInput = $("#reading-title");
            const errorDiv = $("#reading-error-message");

            if (!moduleIdInput || !titleInput || !titleInput.value.trim()) {
                showError("Title is required", errorDiv);
                return;
            }

            const moduleId = Number(moduleIdInput.value);
            if (!moduleId) return;

            hideError(errorDiv);

            try {
                const order = await getNextReadingOrder(moduleId);
                const response = await api.post(`/api/modules/${moduleId}/readings`, {
                    title: titleInput.value.trim(),
                    order: order,
                    format: "markdown",
                    content: "",
                });
                const reading = await response.json();
                closeModal("#reading-form-modal");
                window.location.href = `/modules/${moduleId}/readings/${reading.id}/edit`;
            } catch (error) {
                showError(error.message || "Failed to create reading", errorDiv);
            }
        });
    }
});
