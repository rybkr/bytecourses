document.addEventListener("DOMContentLoaded", () => {
    // Extract course ID from URL (e.g., /courses/123)
    const pathMatch = window.location.pathname.match(/^\/courses\/(\d+)/);
    const courseId = pathMatch ? Number(pathMatch[1]) : null;
    if (!courseId || !Number.isFinite(courseId)) {
        console.warn("Could not extract course ID from URL");
        return;
    }

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

    // Save content buttons (placeholder - saves to localStorage for now)
    document.querySelectorAll(".save-content-btn").forEach((btn) => {
        btn.addEventListener("click", async (e) => {
            e.stopPropagation();
            const moduleId = Number(btn.dataset.moduleId);
            const textarea = document.querySelector(`.module-content-input[data-module-id="${moduleId}"]`);
            const statusSpan = document.querySelector(`.content-save-status[data-module-id="${moduleId}"]`);
            
            if (!textarea) return;

            const content = textarea.value.trim();
            
            // For MVP: Save to localStorage (could be enhanced with backend API later)
            const storageKey = `module-content-${courseId}-${moduleId}`;
            try {
                localStorage.setItem(storageKey, content);
                if (statusSpan) {
                    statusSpan.textContent = "Saved";
                    statusSpan.style.color = "var(--success-color)";
                    setTimeout(() => {
                        statusSpan.textContent = "";
                    }, 2000);
                }
            } catch (error) {
                if (statusSpan) {
                    statusSpan.textContent = "Failed to save";
                    statusSpan.style.color = "var(--danger-color)";
                }
            }
        });
    });

    // Load saved content from localStorage on page load
    document.querySelectorAll(".module-content-input").forEach((textarea) => {
        const moduleId = Number(textarea.dataset.moduleId);
        const storageKey = `module-content-${courseId}-${moduleId}`;
        try {
            const saved = localStorage.getItem(storageKey);
            if (saved) {
                textarea.value = saved;
            }
        } catch (error) {
            // Ignore localStorage errors
        }
    });

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
