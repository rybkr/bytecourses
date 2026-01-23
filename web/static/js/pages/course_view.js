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

    // Module Management
    const addModuleBtn = $("#add-module-btn");
    if (addModuleBtn) {
        on(addModuleBtn, "click", async () => {
            const title = prompt("Module title:");
            if (!title || !title.trim()) return;

            const description = prompt("Module description (optional):") || "";
            const orderStr = prompt("Order (number, optional):") || "0";
            const order = parseInt(orderStr, 10) || 0;

            try {
                await api.post(`/api/courses/${courseId}/modules`, {
                    title: title.trim(),
                    description: description.trim(),
                    order: order,
                });
                window.location.reload();
            } catch (error) {
                alert(error.message || "Failed to create module");
            }
        });
    }

    // Edit Module
    document.querySelectorAll(".btn-edit-module").forEach((btn) => {
        on(btn, "click", async (e) => {
            const moduleId = Number(btn.dataset.moduleId);
            if (!moduleId) return;

            const moduleItem = document.querySelector(`[data-module-id="${moduleId}"]`);
            if (!moduleItem) return;

            const titleEl = moduleItem.querySelector("h3");
            const descEl = moduleItem.querySelector(".module-description");

            const currentTitle = titleEl ? titleEl.textContent.trim() : "";
            const currentDesc = descEl ? descEl.textContent.trim() : "";

            const newTitle = prompt("Module title:", currentTitle);
            if (newTitle === null) return;

            const newDesc = prompt("Module description:", currentDesc) || "";
            const orderStr = prompt("Order (number):", "0") || "0";
            const order = parseInt(orderStr, 10) || 0;

            try {
                await api.patch(`/api/courses/${courseId}/modules/${moduleId}`, {
                    title: newTitle.trim(),
                    description: newDesc.trim(),
                    order: order,
                });
                window.location.reload();
            } catch (error) {
                alert(error.message || "Failed to update module");
            }
        });
    });

    // Delete Module
    document.querySelectorAll(".btn-delete-module").forEach((btn) => {
        on(btn, "click", async (e) => {
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

    // Add Reading
    document.querySelectorAll(".btn-add-reading").forEach((btn) => {
        on(btn, "click", async (e) => {
            const moduleId = Number(btn.dataset.moduleId);
            if (!moduleId) return;

            const title = prompt("Reading title:");
            if (!title || !title.trim()) return;

            const orderStr = prompt("Order (number, optional):") || "0";
            const order = parseInt(orderStr, 10) || 0;

            const content = prompt("Reading content (Markdown):") || "";

            try {
                const response = await api.post(`/api/modules/${moduleId}/readings`, {
                    title: title.trim(),
                    order: order,
                    format: "markdown",
                    content: content,
                });
                const reading = await response.json();
                window.location.href = `/modules/${moduleId}/readings/${reading.id}/edit`;
            } catch (error) {
                alert(error.message || "Failed to create reading");
            }
        });
    });
});
