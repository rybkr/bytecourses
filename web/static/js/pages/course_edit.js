import api from "../core/api.js";
import FormHandler from "../components/FormHandler.js";
import { escapeHtml, confirmAction } from "../core/utils.js";
import { $, on } from "../core/dom.js";
import { showError, hideError } from "../core/utils.js";

document.addEventListener("DOMContentLoaded", () => {
    const form = $("#course-form");
    if (!form) return;

    const courseId = Number(form.dataset.courseId);
    if (!Number.isFinite(courseId) || courseId <= 0) return;

    const saveDelay = Number(form.dataset.autosaveDelay) || 2000;
    const errorDiv = $("#error-message");
    const publishBtn = $("#publishBtn");
    const saveBtn = $("#saveBtn");

    const fieldIds = ["title", "summary", "target_audience", "learning_objectives", "assumed_prerequisites"];

    form.addEventListener("submit", (e) => e.preventDefault());

    const handler = new FormHandler("#course-form", {
        apiPath: "/api/courses",
        entityId: courseId,
        autosaveDelay: saveDelay,
        fieldIds: fieldIds,
        errorContainer: "#error-message",
        statusContainer: "#save-status",
    });

    async function publish() {
        await handler.saveNow();

        hideError(errorDiv);
        publishBtn.disabled = true;

        try {
            await api.post(`/api/courses/${courseId}/publish`);
            window.location.href = `/courses/${courseId}`;
        } catch (error) {
            showError(error.message || "Publish failed", errorDiv);
            publishBtn.disabled = false;
        }
    }

    if (publishBtn) {
        on(publishBtn, "click", (e) => {
            e.preventDefault();
            publish().catch(() => {
                showError("Publish failed", errorDiv);
                publishBtn.disabled = false;
            });
        });
    }

    if (saveBtn) {
        on(saveBtn, "click", (e) => {
            e.preventDefault();
            handler.saveNow().catch(() => {
                showError("Save failed", errorDiv);
            });
        });
    }

    document.addEventListener("visibilitychange", () => {
        if (document.visibilityState === "hidden") {
            handler.saveNow().catch(() => {});
        }
    });

    const modulesList = $("#modules-list");
    const addModuleBtn = $("#add-module-btn");
    const modulesError = $("#modules-error");
    let modules = [];
    let editingModuleId = null;

    async function loadModules() {
        try {
            const response = await api.get(`/api/courses/${courseId}/modules`);
            if (!response) return;

            modules = await response.json();
            renderModules();
        } catch (e) {
            showModulesError(e.message || "Failed to load modules");
        }
    }

    function renderModules() {
        if (!modulesList) return;

        if (modules.length === 0) {
            modulesList.innerHTML =
                '<p style="color: var(--text-muted); margin: 1rem 0;">No modules yet. Click "Add Module" to create one.</p>';
            return;
        }

        modulesList.innerHTML = modules
            .map((module) => {
                const isEditing = editingModuleId === module.id;
                return `
                <div class="module-item" data-module-id="${module.id}">
                    ${
                        isEditing
                            ? `
                        <input type="text" class="module-edit-input" value="${escapeHtml(module.title)}"
                               data-module-id="${module.id}" />
                        <div class="module-item-actions">
                            <button type="button" class="btn btn-small btn-primary save-module-btn" data-module-id="${module.id}">Save</button>
                            <button type="button" class="btn btn-small btn-secondary cancel-edit-btn" data-module-id="${module.id}">Cancel</button>
                        </div>
                    `
                            : `
                        <div class="module-item-content">
                            <span class="module-position">${module.position}</span>
                            <span class="module-title">${escapeHtml(module.title)}</span>
                        </div>
                        <div class="module-item-actions">
                            <button type="button" class="btn btn-small btn-secondary edit-module-btn" data-module-id="${module.id}">Edit</button>
                            <button type="button" class="btn btn-small btn-danger delete-module-btn" data-module-id="${module.id}">Delete</button>
                        </div>
                    `
                    }
                </div>
            `;
            })
            .join("");

        modulesList.querySelectorAll(".edit-module-btn").forEach((btn) => {
            btn.addEventListener("click", () => {
                const moduleId = Number(btn.dataset.moduleId);
                editingModuleId = moduleId;
                renderModules();
                const input = modulesList.querySelector(`.module-edit-input[data-module-id="${moduleId}"]`);
                if (input) {
                    input.focus();
                    input.select();
                }
            });
        });

        modulesList.querySelectorAll(".cancel-edit-btn").forEach((btn) => {
            btn.addEventListener("click", () => {
                editingModuleId = null;
                renderModules();
            });
        });

        modulesList.querySelectorAll(".save-module-btn").forEach((btn) => {
            btn.addEventListener("click", async () => {
                const moduleId = Number(btn.dataset.moduleId);
                const input = modulesList.querySelector(`.module-edit-input[data-module-id="${moduleId}"]`);
                if (!input) return;

                const newTitle = input.value.trim();
                if (!newTitle) {
                    showModulesError("Module title cannot be empty");
                    return;
                }

                await updateModule(moduleId, newTitle);
            });
        });

        modulesList.querySelectorAll(".module-edit-input").forEach((input) => {
            input.addEventListener("keydown", (e) => {
                if (e.key === "Enter") {
                    e.preventDefault();
                    const moduleId = Number(input.dataset.moduleId);
                    const saveBtn = modulesList.querySelector(`.save-module-btn[data-module-id="${moduleId}"]`);
                    saveBtn?.click();
                } else if (e.key === "Escape") {
                    editingModuleId = null;
                    renderModules();
                }
            });
        });

        modulesList.querySelectorAll(".delete-module-btn").forEach((btn) => {
            btn.addEventListener("click", async () => {
                const moduleId = Number(btn.dataset.moduleId);
                const moduleTitle = modules.find((m) => m.id === moduleId)?.title || "this module";
                if (!confirmAction(`Are you sure you want to delete "${moduleTitle}"?`)) {
                    return;
                }
                await deleteModule(moduleId);
            });
        });
    }

    async function createModule(title) {
        try {
            const response = await api.post(`/api/courses/${courseId}/modules`, { title });
            if (!response) return;

            const module = await response.json();
            await loadModules();
            editingModuleId = module.id;
            renderModules();
            const input = modulesList.querySelector(`.module-edit-input[data-module-id="${module.id}"]`);
            if (input) {
                input.focus();
                input.select();
            }
            clearModulesError();
        } catch (e) {
            showModulesError(e.message || "Failed to create module");
        }
    }

    async function updateModule(moduleId, title) {
        try {
            await api.patch(`/api/courses/${courseId}/modules/${moduleId}`, { title });
            editingModuleId = null;
            await loadModules();
            clearModulesError();
        } catch (e) {
            showModulesError(e.message || "Failed to update module");
        }
    }

    async function deleteModule(moduleId) {
        try {
            await api.delete(`/api/courses/${courseId}/modules/${moduleId}`);
            await loadModules();
            clearModulesError();
        } catch (e) {
            showModulesError(e.message || "Failed to delete module");
        }
    }

    function showModulesError(message) {
        if (modulesError) {
            modulesError.textContent = message;
            modulesError.style.display = "block";
        }
    }

    function clearModulesError() {
        if (modulesError) {
            modulesError.style.display = "none";
        }
    }

    if (addModuleBtn) {
        addModuleBtn.addEventListener("click", () => {
            const title = prompt("Enter module title:");
            if (title && title.trim()) {
                createModule(title.trim());
            }
        });
    }

    if (modulesList) {
        loadModules();
    }
});
