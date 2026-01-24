import api from "../core/api.js";
import { $, on } from "../core/dom.js";

document.addEventListener("DOMContentLoaded", () => {
    const pathMatch = window.location.pathname.match(/\/courses\/(\d+)\/content/);
    const courseId = pathMatch ? Number(pathMatch[1]) : null;

    if (!courseId || !Number.isFinite(courseId) || courseId <= 0) {
        return;
    }

    // ========================
    // Module Accordion
    // ========================

    document.querySelectorAll(".module-toggle").forEach((btn) => {
        on(btn, "click", () => {
            const module = btn.closest(".course-content-module");
            if (module) {
                module.classList.toggle("expanded");
            }
        });
    });

    // Auto-expand module with active content
    const activeContent = $(".content-item-link.active");
    if (activeContent) {
        const module = activeContent.closest(".course-content-module");
        if (module) {
            module.classList.add("expanded");
            setTimeout(() => {
                activeContent.scrollIntoView({ behavior: "smooth", block: "center" });
            }, 100);
        }
    }

    // ========================
    // Helpers
    // ========================

    async function getNextModuleOrder() {
        const response = await api.get(`/api/courses/${courseId}/modules`);
        const modules = await response.json();
        if (!modules || modules.length === 0) return 0;
        return Math.max(...modules.map(m => m.order || 0)) + 1;
    }

    async function getNextReadingOrder(moduleId) {
        const response = await api.get(`/api/courses/${courseId}/modules/${moduleId}/content`);
        const readings = await response.json();
        if (!readings || readings.length === 0) return 0;
        return Math.max(...readings.map(r => r.order || 0)) + 1;
    }

    function hideAllForms() {
        document.querySelectorAll(".inline-edit-form, .delete-confirm, .add-content-menu, .delete-confirm-inline").forEach(el => {
            el.style.display = "none";
        });
        document.querySelectorAll(".module-header, .add-content-trigger, .content-item").forEach(el => {
            el.style.display = "";
        });
        document.querySelectorAll(".btn-delete-content").forEach(el => {
            el.style.display = "";
        });
    }

    // ========================
    // Module Edit
    // ========================

    document.querySelectorAll(".btn-edit-module").forEach((btn) => {
        on(btn, "click", (e) => {
            e.stopPropagation();
            hideAllForms();

            const moduleId = btn.dataset.moduleId;
            const moduleEl = document.querySelector(`.course-content-module[data-module-id="${moduleId}"]`);
            if (!moduleEl) return;

            const header = moduleEl.querySelector(".module-header");
            const editForm = moduleEl.querySelector(".inline-edit-form");

            if (header && editForm) {
                header.style.display = "none";
                editForm.style.display = "block";
                editForm.querySelector(".inline-input").focus();
            }
        });
    });

    document.querySelectorAll(".btn-save-module").forEach((btn) => {
        on(btn, "click", async () => {
            const form = btn.closest(".inline-edit-form");
            if (!form) return;

            const moduleId = Number(form.dataset.moduleId);
            const moduleEl = form.closest(".course-content-module");
            const titleInput = form.querySelector(".inline-input");
            const descInput = form.querySelector(".inline-textarea");
            const title = titleInput.value.trim();

            if (!title) {
                titleInput.focus();
                return;
            }

            const currentOrder = moduleEl ? Number(moduleEl.dataset.order) : 0;

            try {
                await api.patch(`/api/courses/${courseId}/modules/${moduleId}`, {
                    title: title,
                    description: descInput.value.trim(),
                    order: currentOrder,
                });

                const titleSpan = moduleEl.querySelector(".module-title");
                if (titleSpan) titleSpan.textContent = title;

                hideAllForms();
            } catch (error) {
                showToast(error.message || "Failed to update module", "error");
            }
        });
    });

    document.querySelectorAll(".btn-cancel-edit").forEach((btn) => {
        on(btn, "click", () => hideAllForms());
    });

    // ========================
    // Module Delete
    // ========================

    document.querySelectorAll(".btn-delete-module").forEach((btn) => {
        on(btn, "click", (e) => {
            e.stopPropagation();
            hideAllForms();

            const moduleId = btn.dataset.moduleId;
            const moduleEl = document.querySelector(`.course-content-module[data-module-id="${moduleId}"]`);
            if (!moduleEl) return;

            const header = moduleEl.querySelector(".module-header");
            const confirm = moduleEl.querySelector(".delete-confirm");

            if (header && confirm) {
                header.style.display = "none";
                confirm.style.display = "flex";
            }
        });
    });

    document.querySelectorAll(".btn-confirm-delete-module").forEach((btn) => {
        on(btn, "click", async () => {
            const confirm = btn.closest(".delete-confirm");
            if (!confirm) return;

            const moduleId = Number(confirm.dataset.moduleId);

            try {
                await api.delete(`/api/courses/${courseId}/modules/${moduleId}`);
                const moduleEl = document.querySelector(`.course-content-module[data-module-id="${moduleId}"]`);
                if (moduleEl) {
                    moduleEl.style.opacity = "0";
                    moduleEl.style.transform = "translateX(-10px)";
                    setTimeout(() => moduleEl.remove(), 200);
                }
            } catch (error) {
                showToast(error.message || "Failed to delete module", "error");
                hideAllForms();
            }
        });
    });

    document.querySelectorAll(".btn-cancel-delete").forEach((btn) => {
        on(btn, "click", () => hideAllForms());
    });

    // ========================
    // New Module
    // ========================

    const addModuleBtn = $("#add-module-btn");
    const newModuleForm = $("#new-module-form");

    if (addModuleBtn && newModuleForm) {
        on(addModuleBtn, "click", () => {
            hideAllForms();
            addModuleBtn.style.display = "none";
            newModuleForm.style.display = "block";
            newModuleForm.querySelector(".inline-input").focus();
        });

        const createBtn = newModuleForm.querySelector(".btn-create-module");
        if (createBtn) {
            on(createBtn, "click", async () => {
                const titleInput = newModuleForm.querySelector(".inline-input");
                const descInput = newModuleForm.querySelector(".inline-textarea");
                const title = titleInput.value.trim();

                if (!title) {
                    titleInput.focus();
                    return;
                }

                try {
                    const order = await getNextModuleOrder();
                    await api.post(`/api/courses/${courseId}/modules`, {
                        title: title,
                        description: descInput.value.trim(),
                        order: order,
                    });
                    window.location.reload();
                } catch (error) {
                    showToast(error.message || "Failed to create module", "error");
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

    // ========================
    // Add Content Menu
    // ========================

    document.querySelectorAll(".add-content-trigger").forEach((trigger) => {
        on(trigger, "click", () => {
            hideAllForms();
            const row = trigger.closest(".add-content-row");
            if (!row) return;

            const menu = row.querySelector(".add-content-menu");
            if (menu) {
                trigger.style.display = "none";
                menu.style.display = "flex";
            }
        });
    });

    document.querySelectorAll(".add-content-option").forEach((option) => {
        on(option, "click", () => {
            const row = option.closest(".add-content-row");
            if (!row) return;

            const moduleId = Number(row.dataset.moduleId);
            window.location.href = `/courses/${courseId}/modules/${moduleId}/content/new`;
        });
    });


    // ========================
    // Delete Content
    // ========================

    document.querySelectorAll(".btn-delete-content").forEach((btn) => {
        on(btn, "click", (e) => {
            e.preventDefault();
            e.stopPropagation();
            hideAllForms();

            const contentItem = btn.closest(".content-item");
            if (!contentItem) return;

            const confirm = contentItem.querySelector(".delete-confirm-inline");
            if (confirm) {
                btn.style.display = "none";
                confirm.style.display = "flex";
            }
        });
    });

    document.querySelectorAll(".btn-confirm-delete-content").forEach((btn) => {
        on(btn, "click", async (e) => {
            e.preventDefault();
            e.stopPropagation();

            const contentItem = btn.closest(".content-item");
            if (!contentItem) return;

            const readingId = Number(contentItem.dataset.readingId);
            const moduleEl = contentItem.closest(".course-content-module");
            if (!moduleEl) return;

            const moduleId = Number(moduleEl.dataset.moduleId);

            try {
                await api.delete(`/api/courses/${courseId}/modules/${moduleId}/content/${readingId}`);
                contentItem.style.opacity = "0";
                contentItem.style.transform = "translateX(-10px)";
                setTimeout(() => contentItem.remove(), 200);
            } catch (error) {
                showToast(error.message || "Failed to delete content", "error");
                hideAllForms();
            }
        });
    });

    document.querySelectorAll(".btn-cancel-delete-content").forEach((btn) => {
        on(btn, "click", (e) => {
            e.preventDefault();
            e.stopPropagation();
            hideAllForms();
        });
    });

    // ========================
    // Toast Notifications
    // ========================

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

    // Close forms on escape key
    document.addEventListener("keydown", (e) => {
        if (e.key === "Escape") {
            hideAllForms();
            const addModuleBtn = $("#add-module-btn");
            const newModuleForm = $("#new-module-form");
            if (newModuleForm && newModuleForm.style.display !== "none") {
                newModuleForm.style.display = "none";
                if (addModuleBtn) addModuleBtn.style.display = "";
            }
        }
    });

    // Close menus when clicking outside
    document.addEventListener("click", (e) => {
        if (!e.target.closest(".add-content-row")) {
            document.querySelectorAll(".add-content-menu").forEach(menu => {
                if (menu.style.display !== "none") {
                    menu.style.display = "none";
                    const row = menu.closest(".add-content-row");
                    if (row) {
                        const trigger = row.querySelector(".add-content-trigger");
                        if (trigger) trigger.style.display = "";
                    }
                }
            });
        }
    });
});
