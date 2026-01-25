import api from "../core/api.js";
import { $, on } from "../core/dom.js";

document.addEventListener("DOMContentLoaded", () => {
    const pathMatch = window.location.pathname.match(
        /\/courses\/(\d+)\/modules$/,
    );
    const courseId = pathMatch ? Number(pathMatch[1]) : null;

    if (!courseId || !Number.isFinite(courseId) || courseId <= 0) {
        return;
    }

    async function getNextModuleOrder() {
        const response = await api.get(`/api/courses/${courseId}/modules`);
        const modules = await response.json();
        if (!modules || modules.length === 0) return 0;
        return Math.max(...modules.map((m) => m.order || 0)) + 1;
    }

    const addModuleBtn = $("#add-module-btn");
    const newModuleForm = $("#new-module-form");

    if (addModuleBtn && newModuleForm) {
        on(addModuleBtn, "click", () => {
            addModuleBtn.style.display = "none";
            newModuleForm.style.display = "block";
            newModuleForm.querySelector(".inline-input").focus();
        });

        const createBtn = newModuleForm.querySelector(".btn-create-module");
        if (createBtn) {
            on(createBtn, "click", async () => {
                const titleInput = newModuleForm.querySelector(".inline-input");
                const descInput =
                    newModuleForm.querySelector(".inline-textarea");
                const title = titleInput.value.trim();

                if (!title) {
                    titleInput.focus();
                    return;
                }

                createBtn.disabled = true;
                createBtn.textContent = "Creating...";

                try {
                    const order = await getNextModuleOrder();
                    await api.post(`/api/courses/${courseId}/modules`, {
                        title: title,
                        description: descInput.value.trim(),
                        order: order,
                    });
                    window.location.reload();
                } catch (error) {
                    showToast(
                        error.message || "Failed to create module",
                        "error",
                    );
                    createBtn.disabled = false;
                    createBtn.textContent = "Create";
                }
            });
        }

        const cancelBtn = newModuleForm.querySelector(".btn-cancel-new-module");
        if (cancelBtn) {
            on(cancelBtn, "click", () => {
                newModuleForm.style.display = "none";
                addModuleBtn.style.display = "";
                newModuleForm.querySelector(".inline-input").value = "";
                newModuleForm.querySelector(".inline-textarea").value = "";
            });
        }
    }

    document.addEventListener("keydown", (e) => {
        if (e.key === "Escape") {
            if (newModuleForm && newModuleForm.style.display !== "none") {
                newModuleForm.style.display = "none";
                if (addModuleBtn) addModuleBtn.style.display = "";
            }
        }
    });

    // Expandable module functionality
    function initExpandableModules() {
        const modules = document.querySelectorAll(".course-content-module");
        const storageKey = `course-${courseId}-expanded-modules`;

        // Load expanded state from localStorage
        let expandedModules = new Set();
        try {
            const stored = localStorage.getItem(storageKey);
            if (stored) {
                expandedModules = new Set(JSON.parse(stored));
            }
        } catch (e) {
            // Ignore localStorage errors
        }

        // Check if current URL is a reading page and auto-expand that module
        const pathMatch = window.location.pathname.match(
            /\/courses\/\d+\/modules\/(\d+)\/content\/\d+/,
        );
        if (pathMatch) {
            const currentModuleId = pathMatch[1];
            expandedModules.add(currentModuleId);
        }

        modules.forEach((module) => {
            const moduleId = module.dataset.moduleId;
            const header = module.querySelector(".course-content-module-header");
            const content = module.querySelector(".course-content-module-content");

            if (!header || !content) return;

            // Set initial state
            const isExpanded = expandedModules.has(moduleId);
            if (isExpanded) {
                module.classList.add("expanded");
                header.setAttribute("aria-expanded", "true");
            }

            // Toggle on click
            on(header, "click", (e) => {
                // Don't toggle if clicking the link
                if (e.target.closest(".module-card-link")) {
                    return;
                }

                const wasExpanded = module.classList.contains("expanded");
                module.classList.toggle("expanded");
                const nowExpanded = module.classList.contains("expanded");
                header.setAttribute("aria-expanded", nowExpanded ? "true" : "false");

                // Update localStorage
                if (nowExpanded) {
                    expandedModules.add(moduleId);
                } else {
                    expandedModules.delete(moduleId);
                }

                try {
                    localStorage.setItem(
                        storageKey,
                        JSON.stringify(Array.from(expandedModules)),
                    );
                } catch (e) {
                    // Ignore localStorage errors
                }
            });
        });
    }

    initExpandableModules();

    function showToast(message, type = "info") {
        const existing = document.querySelector(".toast");
        if (existing) existing.remove();

        const toast = document.createElement("div");
        toast.className = `toast toast-${type}`;
        toast.textContent = message;
        document.body.appendChild(toast);

        requestAnimationFrame(() => {
            toast.classList.add("show");
        });

        setTimeout(() => {
            toast.classList.remove("show");
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }
});
