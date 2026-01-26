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

    const storageKey = `course-${courseId}-expanded-modules`;

    function getStoredExpanded() {
        try {
            const stored = localStorage.getItem(storageKey);
            return stored ? new Set(JSON.parse(stored)) : new Set();
        } catch (e) {
            return new Set();
        }
    }

    function saveExpanded(expanded) {
        try {
            localStorage.setItem(
                storageKey,
                JSON.stringify(Array.from(expanded)),
            );
        } catch (e) {

        }
    }

    function setModuleExpanded(moduleEl, expanded) {
        const toggle = moduleEl?.querySelector(".course-content-module-toggle");
        const content = moduleEl?.querySelector(".course-content-module-content");
        if (!toggle || !content) return;
        if (expanded) {
            moduleEl.classList.add("expanded");
            toggle.setAttribute("aria-expanded", "true");
        } else {
            moduleEl.classList.remove("expanded");
            toggle.setAttribute("aria-expanded", "false");
        }
    }

    function initExpandableModules() {
        const modules = document.querySelectorAll(".course-content-module");

        let expandedModules = getStoredExpanded();

        modules.forEach((module) => {
            const moduleId = module.dataset.moduleId;
            const toggle = module.querySelector(".course-content-module-toggle");
            const content = module.querySelector(".course-content-module-content");

            if (!toggle || !content) return;

            const isExpanded = expandedModules.has(moduleId);
            if (isExpanded) {
                module.classList.add("expanded");
                toggle.setAttribute("aria-expanded", "true");
            }

            on(toggle, "click", () => {
                const nowExpanded = !module.classList.contains("expanded");
                setModuleExpanded(module, nowExpanded);

                if (nowExpanded) {
                    expandedModules.add(moduleId);
                    module.scrollIntoView({
                        behavior: "smooth",
                        block: "nearest",
                    });
                } else {
                    expandedModules.delete(moduleId);
                }
                saveExpanded(expandedModules);
            });
        });
    }

    initExpandableModules();

    const expandAllBtn = $("#expand-all-btn");
    const collapseAllBtn = $("#collapse-all-btn");
    if (expandAllBtn && collapseAllBtn) {
        on(expandAllBtn, "click", () => {
            const modules = document.querySelectorAll(".course-content-module");
            const expanded = new Set();
            modules.forEach((module) => {
                const id = module.dataset.moduleId;
                if (id) {
                    expanded.add(id);
                    setModuleExpanded(module, true);
                }
            });
            saveExpanded(expanded);
            const first = modules[0];
            if (first) {
                first.scrollIntoView({ behavior: "smooth", block: "nearest" });
            }
        });

        on(collapseAllBtn, "click", () => {
            const modules = document.querySelectorAll(".course-content-module");
            modules.forEach((module) => setModuleExpanded(module, false));
            saveExpanded(new Set());
        });
    }

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

    document.addEventListener("click", async (e) => {
        const modPub = e.target.closest(".publish-module-btn");
        if (modPub) {
            e.preventDefault();
            e.stopPropagation();
            const cid = modPub.dataset.courseId;
            const mid = modPub.dataset.moduleId;
            if (!cid || !mid) return;
            try {
                await api.post(
                    `/api/courses/${cid}/modules/${mid}/actions/publish`,
                );
                window.location.reload();
            } catch (err) {
                showToast(err.message || "Failed to publish module", "error");
            }
            return;
        }

        const pub = e.target.closest(".sidebar-publish-btn");
        const unpub = e.target.closest(".sidebar-unpublish-btn");
        const row = e.target.closest(".course-content-reading-row");
        if (!row) return;
        const cid = row.dataset.courseId;
        const mid = row.dataset.moduleId;
        const rid = row.dataset.readingId;
        if (!cid || !mid || !rid) return;

        const base = `/api/courses/${cid}/modules/${mid}/content/${rid}`;
        if (pub) {
            e.preventDefault();
            e.stopPropagation();
            try {
                await api.post(`${base}/actions/publish`);
                window.location.reload();
            } catch (err) {
                showToast(err.message || "Failed to publish", "error");
            }
            return;
        }
        if (unpub) {
            e.preventDefault();
            e.stopPropagation();
            try {
                await api.post(`${base}/actions/unpublish`);
                window.location.reload();
            } catch (err) {
                showToast(err.message || "Failed to unpublish", "error");
            }
        }
    });
});
