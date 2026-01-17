document.addEventListener("DOMContentLoaded", () => {
    // Extract course ID from URL (e.g., /courses/123)
    const pathMatch = window.location.pathname.match(/^\/courses\/(\d+)/);
    const courseId = pathMatch ? Number(pathMatch[1]) : null;
    if (!courseId || !Number.isFinite(courseId)) {
        console.warn("Could not extract course ID from URL");
        return;
    }

    const curriculumSection = document.querySelector(".curriculum-section");
    const isInstructor = curriculumSection?.dataset.isInstructor === "true";

    // Accordion functionality
    const accordions = document.querySelectorAll(".module-accordion");

    accordions.forEach((accordion) => {
        const header = accordion.querySelector(".module-accordion-header");
        if (!header) return;

        // Don't toggle accordion when clicking action buttons
        const toggleOnly = (e) => {
            if (e.target.closest(".module-actions, .module-accordion-toggle")) {
                return;
            }
            const isExpanded = accordion.classList.contains("expanded");

            if (isExpanded) {
                accordion.classList.remove("expanded");
                accordion.setAttribute("aria-expanded", "false");
            } else {
                accordion.classList.add("expanded");
                accordion.setAttribute("aria-expanded", "true");
            }
        };

        header.addEventListener("click", toggleOnly);

        header.addEventListener("keydown", (e) => {
            if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                toggleOnly(e);
            }
        });

        // Set initial aria-expanded state
        const isInitiallyExpanded = accordion.classList.contains("expanded");
        accordion.setAttribute("aria-expanded", isInitiallyExpanded ? "true" : "false");

        // Load content for this module
        const moduleId = Number(accordion.dataset.moduleId);
        if (moduleId) {
            loadModuleContent(moduleId);
        }
    });

    // Instructor-only functionality
    const addModuleBtn = document.getElementById("add-module-btn");
    if (addModuleBtn) {
        addModuleBtn.addEventListener("click", async () => {
            const title = prompt("Enter module title:");
            if (!title || !title.trim()) {
                return;
            }

            try {
                const res = await fetch(`/api/courses/${courseId}/modules`, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ title: title.trim() }),
                });

                if (!res.ok) {
                    const txt = await res.text();
                    alert(txt || "Failed to create module");
                    return;
                }

                // Reload page to show new module
                window.location.reload();
            } catch (error) {
                alert("Network error. Please try again.");
            }
        });
    }

    // Edit module buttons
    document.querySelectorAll(".edit-module-btn").forEach((btn) => {
        btn.addEventListener("click", (e) => {
            e.stopPropagation();
            const moduleId = Number(btn.dataset.moduleId);
            const currentTitle = btn.dataset.moduleTitle || "";
            const newTitle = prompt("Enter new module title:", currentTitle);

            if (!newTitle || !newTitle.trim() || newTitle.trim() === currentTitle) {
                return;
            }

            updateModuleTitle(courseId, moduleId, newTitle.trim());
        });
    });

    // Delete module buttons
    document.querySelectorAll(".delete-module-btn").forEach((btn) => {
        btn.addEventListener("click", async (e) => {
            e.stopPropagation();
            const moduleId = Number(btn.dataset.moduleId);
            const moduleTitle = btn.dataset.moduleTitle || "this module";

            if (!confirm(`Are you sure you want to delete "${moduleTitle}"? This action cannot be undone.`)) {
                return;
            }

            try {
                const res = await fetch(`/api/courses/${courseId}/modules/${moduleId}`, {
                    method: "DELETE",
                });

                if (!res.ok) {
                    const txt = await res.text();
                    alert(txt || "Failed to delete module");
                    return;
                }

                // Reload page to reflect deletion
                window.location.reload();
            } catch (error) {
                alert("Network error. Please try again.");
            }
        });
    });

    // Add lecture buttons
    document.querySelectorAll(".add-lecture-btn").forEach((btn) => {
        btn.addEventListener("click", async (e) => {
            e.stopPropagation();
            const moduleId = Number(btn.dataset.moduleId);

            const title = prompt("Enter lecture title:");
            if (!title || !title.trim()) {
                return;
            }

            try {
                const res = await fetch(`/api/courses/${courseId}/modules/${moduleId}/content`, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ title: title.trim() }),
                });

                if (!res.ok) {
                    const txt = await res.text();
                    alert(txt || "Failed to create lecture");
                    return;
                }

                const item = await res.json();
                // Redirect to edit page
                window.location.href = `/courses/${courseId}/modules/${moduleId}/content/${item.id}/edit`;
            } catch (error) {
                alert("Network error. Please try again.");
            }
        });
    });

    async function loadModuleContent(moduleId) {
        const contentList = document.querySelector(`.content-list[data-module-id="${moduleId}"]`);
        if (!contentList) return;

        try {
            const res = await fetch(`/api/courses/${courseId}/modules/${moduleId}/content`);
            if (!res.ok) {
                contentList.innerHTML = '<div class="content-list-error">Failed to load content</div>';
                return;
            }

            const data = await res.json();
            renderContentList(moduleId, data.items || [], data.lectures || {});
        } catch (error) {
            contentList.innerHTML = '<div class="content-list-error">Failed to load content</div>';
        }
    }

    function renderContentList(moduleId, items, lectures) {
        const contentList = document.querySelector(`.content-list[data-module-id="${moduleId}"]`);
        if (!contentList) return;

        if (!items || items.length === 0) {
            contentList.innerHTML = '<div class="content-list-empty">No content yet.</div>';
            return;
        }

        const html = items.map(item => {
            const statusBadge = isInstructor ?
                `<span class="content-status-badge content-status-${item.status}">${item.status}</span>` : '';

            const viewUrl = `/courses/${courseId}/modules/${moduleId}/content/${item.id}`;
            const editUrl = `/courses/${courseId}/modules/${moduleId}/content/${item.id}/edit`;

            const deleteBtn = isInstructor ?
                `<button type="button" class="btn btn-small btn-danger delete-content-btn" data-content-id="${item.id}" data-content-title="${item.title}">Delete</button>` : '';

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
        }).join('');

        contentList.innerHTML = html;

        // Attach delete handlers
        contentList.querySelectorAll(".delete-content-btn").forEach(btn => {
            btn.addEventListener("click", async (e) => {
                e.stopPropagation();
                e.preventDefault();
                const contentId = Number(btn.dataset.contentId);
                const contentTitle = btn.dataset.contentTitle || "this content";

                if (!confirm(`Are you sure you want to delete "${contentTitle}"? This action cannot be undone.`)) {
                    return;
                }

                try {
                    const res = await fetch(`/api/courses/${courseId}/modules/${moduleId}/content/${contentId}`, {
                        method: "DELETE",
                    });

                    if (!res.ok) {
                        const txt = await res.text();
                        alert(txt || "Failed to delete content");
                        return;
                    }

                    // Reload content list
                    loadModuleContent(moduleId);
                } catch (error) {
                    alert("Network error. Please try again.");
                }
            });
        });
    }

    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    async function updateModuleTitle(courseId, moduleId, title) {
        try {
            const res = await fetch(`/api/courses/${courseId}/modules/${moduleId}`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ title }),
            });

            if (!res.ok) {
                const txt = await res.text();
                alert(txt || "Failed to update module");
                return;
            }

            // Reload page to show updated title
            window.location.reload();
        } catch (error) {
            alert("Network error. Please try again.");
        }
    }
});
