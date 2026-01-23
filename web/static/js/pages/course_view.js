import api from "../core/api.js";
import { escapeHtml, confirmAction } from "../core/utils.js";
import { $, $$ } from "../core/dom.js";

document.addEventListener("DOMContentLoaded", () => {
    const pathMatch = window.location.pathname.match(/^\/courses\/(\d+)/);
    const courseId = pathMatch ? Number(pathMatch[1]) : null;
    if (!courseId || !Number.isFinite(courseId)) return;

    const curriculumSection = $(".curriculum-section");
    const isInstructor = curriculumSection?.dataset.isInstructor === "true";

    function initAccordions() {
        $$(".module-accordion").forEach((accordion) => {
            const header = accordion.querySelector(".module-accordion-header");
            if (!header) return;

            const toggleAccordion = (e) => {
                if (e.target.closest(".module-actions, .module-accordion-toggle")) return;

                const isExpanded = accordion.classList.contains("expanded");
                if (isExpanded) {
                    accordion.classList.remove("expanded");
                    accordion.setAttribute("aria-expanded", "false");
                } else {
                    accordion.classList.add("expanded");
                    accordion.setAttribute("aria-expanded", "true");
                }
            };

            header.addEventListener("click", toggleAccordion);
            header.addEventListener("keydown", (e) => {
                if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    toggleAccordion(e);
                }
            });

            const isInitiallyExpanded = accordion.classList.contains("expanded");
            accordion.setAttribute("aria-expanded", isInitiallyExpanded ? "true" : "false");

            const moduleId = Number(accordion.dataset.moduleId);
            if (moduleId) {
                loadModuleContent(moduleId);
            }
        });
    }

    async function loadModuleContent(moduleId) {
        const contentList = $(`.content-list[data-module-id="${moduleId}"]`);
        if (!contentList) return;

        try {
            const response = await api.get(`/api/courses/${courseId}/modules/${moduleId}/content`);
            if (!response) return;

            const data = await response.json();
            renderContentList(moduleId, data.items || []);
        } catch (error) {
            contentList.innerHTML = '<div class="content-list-error">Failed to load content</div>';
        }
    }

    function renderContentList(moduleId, items) {
        const contentList = $(`.content-list[data-module-id="${moduleId}"]`);
        if (!contentList) return;

        if (!items || items.length === 0) {
            contentList.innerHTML = '<div class="content-list-empty">No content yet.</div>';
            return;
        }

        const html = items
            .map((item) => {
                const statusBadge = isInstructor
                    ? `<span class="content-status-badge content-status-${item.status}">${item.status}</span>`
                    : "";

                const viewUrl = `/courses/${courseId}/modules/${moduleId}/content/${item.id}`;
                const editUrl = `/courses/${courseId}/modules/${moduleId}/content/${item.id}/edit`;

                const deleteBtn = isInstructor
                    ? `<button type="button" class="btn btn-small btn-danger delete-content-btn" data-content-id="${item.id}" data-content-title="${escapeHtml(item.title)}">Delete</button>`
                    : "";

                return `
                    <div class="content-item" data-content-id="${item.id}">
                        <div class="content-item-icon">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                                <polyline points="14 2 14 8 20 8"></polyline>
                                <line x1="16" y1="13" x2="8" y2="13"></line>
                                <line x1="16" y1="17" x2="8" y2="17"></line>
                                <polyline points="10 9 9 9 8 9"></polyline>
                            </svg>
                        </div>
                        <div class="content-item-info">
                            <a href="${isInstructor ? editUrl : viewUrl}" class="content-item-title">${escapeHtml(item.title)}</a>
                            ${statusBadge}
                        </div>
                        <div class="content-item-actions">
                            ${deleteBtn}
                        </div>
                    </div>
                `;
            })
            .join("");

        contentList.innerHTML = html;

        contentList.querySelectorAll(".delete-content-btn").forEach((btn) => {
            btn.addEventListener("click", async (e) => {
                e.stopPropagation();
                e.preventDefault();

                const contentId = Number(btn.dataset.contentId);
                const contentTitle = btn.dataset.contentTitle || "this content";

                if (!confirmAction(`Are you sure you want to delete "${contentTitle}"? This action cannot be undone.`)) {
                    return;
                }

                try {
                    await api.delete(`/api/courses/${courseId}/modules/${moduleId}/content/${contentId}`);
                    loadModuleContent(moduleId);
                } catch (error) {
                    alert(error.message || "Failed to delete content");
                }
            });
        });
    }

    async function updateModuleTitle(moduleId, title) {
        try {
            await api.patch(`/api/courses/${courseId}/modules/${moduleId}`, { title });
            window.location.reload();
        } catch (error) {
            alert(error.message || "Failed to update module");
        }
    }

    const addModuleBtn = $("#add-module-btn");
    if (addModuleBtn) {
        addModuleBtn.addEventListener("click", async () => {
            const title = prompt("Enter module title:");
            if (!title || !title.trim()) return;

            try {
                await api.post(`/api/courses/${courseId}/modules`, { title: title.trim() });
                window.location.reload();
            } catch (error) {
                alert(error.message || "Failed to create module");
            }
        });
    }

    $$(".edit-module-btn").forEach((btn) => {
        btn.addEventListener("click", (e) => {
            e.stopPropagation();
            const moduleId = Number(btn.dataset.moduleId);
            const currentTitle = btn.dataset.moduleTitle || "";
            const newTitle = prompt("Enter new module title:", currentTitle);

            if (!newTitle || !newTitle.trim() || newTitle.trim() === currentTitle) return;

            updateModuleTitle(moduleId, newTitle.trim());
        });
    });

    $$(".delete-module-btn").forEach((btn) => {
        btn.addEventListener("click", async (e) => {
            e.stopPropagation();
            const moduleId = Number(btn.dataset.moduleId);
            const moduleTitle = btn.dataset.moduleTitle || "this module";

            if (!confirmAction(`Are you sure you want to delete "${moduleTitle}"? This action cannot be undone.`)) {
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

    $$(".add-lecture-btn").forEach((btn) => {
        btn.addEventListener("click", async (e) => {
            e.stopPropagation();
            const moduleId = Number(btn.dataset.moduleId);

            const title = prompt("Enter lecture title:");
            if (!title || !title.trim()) return;

            try {
                const response = await api.post(`/api/courses/${courseId}/modules/${moduleId}/content`, {
                    title: title.trim(),
                });
                if (!response) return;

                const item = await response.json();
                window.location.href = `/courses/${courseId}/modules/${moduleId}/content/${item.id}/edit`;
            } catch (error) {
                alert(error.message || "Failed to create lecture");
            }
        });
    });

    initAccordions();
});
