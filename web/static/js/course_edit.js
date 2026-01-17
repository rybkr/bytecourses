document.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById("course-form");
    if (!form) {
        return;
    }

    const courseId = Number(form.dataset.courseId);
    if (!Number.isFinite(courseId) || courseId <= 0) {
        console.warn("Missing or invalid course id");
        return;
    }

    const saveDelay = Number(form.dataset.autosaveDelay);

    const errorDiv = document.getElementById("error-message");
    const statusDiv = document.getElementById("save-status");
    const publishBtn = document.getElementById("publishBtn");
    const saveBtn = document.getElementById("saveBtn");

    const fieldIds = ["title", "summary", "target_audience", "learning_objectives", "assumed_prerequisites"];

    let saveTimer = null;
    let dirty = false;
    let saveInFlight = false;
    let lastSavedJson = null;

    function nowLabel() {
        const d = new Date();
        return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" });
    }

    function readPayload() {
        return {
            title: document.getElementById("title")?.value ?? "",
            summary: document.getElementById("summary")?.value ?? "",
            target_audience: document.getElementById("target_audience")?.value ?? "",
            learning_objectives: document.getElementById("learning_objectives")?.value ?? "",
            assumed_prerequisites: document.getElementById("assumed_prerequisites")?.value ?? "",
        };
    }

    async function patchCourse(payload) {
        const res = await fetch(`/api/courses/${courseId}`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload),
        });

        if (!res.ok) {
            const txt = await res.text();
            throw new Error(txt || "Save failed");
        }
    }

    function scheduleSave() {
        dirty = true;
        clearTimeout(saveTimer);
        saveTimer = setTimeout(() => {
            saveNow().catch(() => { });
        }, saveDelay);
    }

    async function saveNow() {
        clearTimeout(saveTimer);
        if (!dirty || saveInFlight) {
            return;
        }

        const payload = readPayload();
        const json = JSON.stringify(payload);
        if (json === lastSavedJson) {
            dirty = false;
            statusDiv.textContent = `Saved at ${nowLabel()}`;
            return;
        }

        saveInFlight = true;
        errorDiv.textContent = "";
        errorDiv.style.display = "none";

        try {
            await patchCourse(payload);
            lastSavedJson = json;
            dirty = false;
            statusDiv.textContent = `Saved at ${nowLabel()}`;
        } catch (e) {
            errorDiv.textContent = e.message || "Autosave failed";
            errorDiv.style.display = "block";
        } finally {
            saveInFlight = false;
        }
    }

    async function publish() {
        await saveNow();
        if (dirty) {
            return;
        }

        errorDiv.textContent = "";
        errorDiv.style.display = "none";
        publishBtn.disabled = true;

        try {
            const res = await fetch(`/api/courses/${courseId}/actions/publish`, {
                method: "POST",
            });

            if (!res.ok) {
                const txt = await res.text();
                errorDiv.textContent = txt || "Publish failed";
                errorDiv.style.display = "block";
                publishBtn.disabled = false;
                return;
            }

            window.location.href = `/courses/${courseId}`;
        } catch (error) {
            errorDiv.textContent = "Network error. Please try again.";
            errorDiv.style.display = "block";
            publishBtn.disabled = false;
        }
    }

    for (const id of fieldIds) {
        const el = document.getElementById(id);
        if (!el) {
            continue;
        }
        el.addEventListener("input", scheduleSave);
        el.addEventListener("blur", () => saveNow().catch(() => { }));
    }

    if (publishBtn) {
        publishBtn.addEventListener("click", (e) => {
            e.preventDefault();
            publish().catch(() => {
                errorDiv.textContent = "Publish failed";
                errorDiv.style.display = "block";
            });
        });
    }

    if (saveBtn) {
        saveBtn.addEventListener("click", (e) => {
            e.preventDefault();
            saveNow().catch(() => {
                errorDiv.textContent = "Save failed";
                errorDiv.style.display = "block";
            });
        });
    }

    document.addEventListener("visibilitychange", () => {
        if (document.visibilityState !== "hidden") {
            return;
        }
        if (dirty) {
            saveNow().catch(() => { });
        }
    });

    // Module management
    const modulesList = document.getElementById("modules-list");
    const addModuleBtn = document.getElementById("add-module-btn");
    const modulesError = document.getElementById("modules-error");
    let modules = [];
    let editingModuleId = null;

    async function loadModules() {
        try {
            const res = await fetch(`/api/courses/${courseId}/modules`);
            if (!res.ok) {
                throw new Error("Failed to load modules");
            }
            modules = await res.json();
            renderModules();
        } catch (e) {
            showModulesError(e.message || "Failed to load modules");
        }
    }

    function renderModules() {
        if (!modulesList) return;
        
        if (modules.length === 0) {
            modulesList.innerHTML = '<p style="color: var(--text-muted); margin: 1rem 0;">No modules yet. Click "Add Module" to create one.</p>';
            return;
        }

        modulesList.innerHTML = modules.map(module => {
            const isEditing = editingModuleId === module.id;
            return `
                <div class="module-item" data-module-id="${module.id}">
                    ${isEditing ? `
                        <input type="text" class="module-edit-input" value="${escapeHtml(module.title)}" 
                               data-module-id="${module.id}" />
                        <div class="module-item-actions">
                            <button type="button" class="btn btn-small btn-primary save-module-btn" data-module-id="${module.id}">Save</button>
                            <button type="button" class="btn btn-small btn-secondary cancel-edit-btn" data-module-id="${module.id}">Cancel</button>
                        </div>
                    ` : `
                        <div class="module-item-content">
                            <span class="module-position">${module.position}</span>
                            <span class="module-title">${escapeHtml(module.title)}</span>
                        </div>
                        <div class="module-item-actions">
                            <button type="button" class="btn btn-small btn-secondary edit-module-btn" data-module-id="${module.id}">Edit</button>
                            <button type="button" class="btn btn-small btn-danger delete-module-btn" data-module-id="${module.id}">Delete</button>
                        </div>
                    `}
                </div>
            `;
        }).join("");

        // Attach event listeners
        modulesList.querySelectorAll(".edit-module-btn").forEach(btn => {
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

        modulesList.querySelectorAll(".cancel-edit-btn").forEach(btn => {
            btn.addEventListener("click", () => {
                editingModuleId = null;
                renderModules();
            });
        });

        modulesList.querySelectorAll(".save-module-btn").forEach(btn => {
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

        modulesList.querySelectorAll(".module-edit-input").forEach(input => {
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

        modulesList.querySelectorAll(".delete-module-btn").forEach(btn => {
            btn.addEventListener("click", async () => {
                const moduleId = Number(btn.dataset.moduleId);
                if (!confirm(`Are you sure you want to delete "${modules.find(m => m.id === moduleId)?.title}"?`)) {
                    return;
                }
                await deleteModule(moduleId);
            });
        });
    }

    async function createModule(title) {
        try {
            const res = await fetch(`/api/courses/${courseId}/modules`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ title }),
            });

            if (!res.ok) {
                const txt = await res.text();
                throw new Error(txt || "Failed to create module");
            }

            const module = await res.json();
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
            const res = await fetch(`/api/courses/${courseId}/modules/${moduleId}`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ title }),
            });

            if (!res.ok) {
                const txt = await res.text();
                throw new Error(txt || "Failed to update module");
            }

            editingModuleId = null;
            await loadModules();
            clearModulesError();
        } catch (e) {
            showModulesError(e.message || "Failed to update module");
        }
    }

    async function deleteModule(moduleId) {
        try {
            const res = await fetch(`/api/courses/${courseId}/modules/${moduleId}`, {
                method: "DELETE",
            });

            if (!res.ok) {
                const txt = await res.text();
                throw new Error(txt || "Failed to delete module");
            }

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

    function escapeHtml(text) {
        const div = document.createElement("div");
        div.textContent = text;
        return div.innerHTML;
    }

    if (addModuleBtn) {
        addModuleBtn.addEventListener("click", () => {
            const title = prompt("Enter module title:");
            if (title && title.trim()) {
                createModule(title.trim());
            }
        });
    }

    // Load modules on page load
    if (modulesList) {
        loadModules();
    }
});
