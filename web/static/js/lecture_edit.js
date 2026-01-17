document.addEventListener("DOMContentLoaded", () => {
    const { courseId, moduleId, contentId, status } = window.LECTURE_DATA || {};

    if (!courseId || !moduleId || !contentId) {
        console.error("Missing lecture data");
        return;
    }

    const titleInput = document.getElementById("lecture-title");
    const contentTextarea = document.getElementById("lecture-content");
    const previewDiv = document.getElementById("lecture-preview");
    const saveStatus = document.getElementById("save-status");
    const saveBtn = document.getElementById("save-btn");
    const publishBtn = document.getElementById("publish-btn");
    const unpublishBtn = document.getElementById("unpublish-btn");

    let saveTimeout = null;
    let lastSavedTitle = titleInput.value;
    let lastSavedContent = contentTextarea.value;
    let isSaving = false;

    const apiUrl = `/api/courses/${courseId}/modules/${moduleId}/content/${contentId}`;

    // Update preview with marked
    function updatePreview() {
        if (typeof marked !== 'undefined') {
            const html = marked.parse(contentTextarea.value || '');
            previewDiv.innerHTML = `<div class="proposal-content-value">${html}</div>`;
        }
    }

    // Initial preview
    updatePreview();

    // Live preview on input
    contentTextarea.addEventListener("input", () => {
        updatePreview();
        scheduleAutosave();
    });

    titleInput.addEventListener("input", () => {
        scheduleAutosave();
    });

    function scheduleAutosave() {
        if (saveTimeout) {
            clearTimeout(saveTimeout);
        }
        saveTimeout = setTimeout(() => {
            autosave();
        }, 2000);
    }

    async function autosave() {
        const currentTitle = titleInput.value.trim();
        const currentContent = contentTextarea.value;

        // Don't save if nothing changed
        if (currentTitle === lastSavedTitle && currentContent === lastSavedContent) {
            return;
        }

        await save(currentTitle, currentContent);
    }

    async function save(title, content) {
        if (isSaving) return;

        isSaving = true;
        updateSaveStatus("saving", "Saving...");

        try {
            const body = {};
            if (title !== undefined && title !== lastSavedTitle) {
                body.title = title;
            }
            if (content !== undefined && content !== lastSavedContent) {
                body.content = content;
            }

            if (Object.keys(body).length === 0) {
                updateSaveStatus("saved", "Saved");
                isSaving = false;
                return;
            }

            const res = await fetch(apiUrl, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(body),
            });

            if (!res.ok) {
                const txt = await res.text();
                throw new Error(txt || "Failed to save");
            }

            lastSavedTitle = title !== undefined ? title : lastSavedTitle;
            lastSavedContent = content !== undefined ? content : lastSavedContent;

            updateSaveStatus("saved", "Saved");
        } catch (error) {
            updateSaveStatus("error", "Failed to save");
            console.error("Save error:", error);
        } finally {
            isSaving = false;
        }
    }

    function updateSaveStatus(type, text) {
        saveStatus.textContent = text;
        saveStatus.className = `save-status ${type}`;

        if (type === "saved") {
            setTimeout(() => {
                if (saveStatus.textContent === text) {
                    saveStatus.textContent = "";
                }
            }, 3000);
        }
    }

    // Manual save button
    saveBtn.addEventListener("click", async () => {
        if (saveTimeout) {
            clearTimeout(saveTimeout);
        }
        await save(titleInput.value.trim(), contentTextarea.value);
    });

    // Publish button
    if (publishBtn) {
        publishBtn.addEventListener("click", async () => {
            // Save first
            await save(titleInput.value.trim(), contentTextarea.value);

            try {
                const res = await fetch(`${apiUrl}/publish`, {
                    method: "POST",
                });

                if (!res.ok) {
                    const txt = await res.text();
                    alert(txt || "Failed to publish");
                    return;
                }

                // Reload to show updated status
                window.location.reload();
            } catch (error) {
                alert("Network error. Please try again.");
            }
        });
    }

    // Unpublish button
    if (unpublishBtn) {
        unpublishBtn.addEventListener("click", async () => {
            try {
                const res = await fetch(`${apiUrl}/unpublish`, {
                    method: "POST",
                });

                if (!res.ok) {
                    const txt = await res.text();
                    alert(txt || "Failed to unpublish");
                    return;
                }

                // Reload to show updated status
                window.location.reload();
            } catch (error) {
                alert("Network error. Please try again.");
            }
        });
    }

    // Warn before leaving with unsaved changes
    window.addEventListener("beforeunload", (e) => {
        const hasUnsavedChanges =
            titleInput.value.trim() !== lastSavedTitle ||
            contentTextarea.value !== lastSavedContent;

        if (hasUnsavedChanges) {
            e.preventDefault();
            e.returnValue = "";
        }
    });

    // Keyboard shortcuts
    document.addEventListener("keydown", (e) => {
        // Ctrl/Cmd + S to save
        if ((e.ctrlKey || e.metaKey) && e.key === "s") {
            e.preventDefault();
            if (saveTimeout) {
                clearTimeout(saveTimeout);
            }
            save(titleInput.value.trim(), contentTextarea.value);
        }
    });
});
