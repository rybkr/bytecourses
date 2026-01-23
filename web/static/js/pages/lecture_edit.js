import api from "../core/api.js";
import { debounce } from "../core/utils.js";
import { $ } from "../core/dom.js";

document.addEventListener("DOMContentLoaded", () => {
    const { courseId, moduleId, readingId } = window.LECTURE_DATA || {};

    if (!courseId || !moduleId || !readingId) return;

    const titleInput = $("#lecture-title");
    const contentTextarea = $("#lecture-content");
    const previewDiv = $("#lecture-preview");
    const saveStatus = $("#save-status");
    const saveBtn = $("#save-btn");
    const publishBtn = $("#publish-btn");
    const unpublishBtn = $("#unpublish-btn");

    let lastSavedTitle = titleInput.value;
    let lastSavedContent = contentTextarea.value;
    let isSaving = false;

    const apiUrl = `/api/modules/${moduleId}/readings/${readingId}`;

    function updatePreview() {
        if (typeof marked !== "undefined") {
            const html = marked.parse(contentTextarea.value || "");
            previewDiv.innerHTML = `<div class="proposal-content-value">${html}</div>`;
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

            await api.patch(apiUrl, body);

            lastSavedTitle = title !== undefined ? title : lastSavedTitle;
            lastSavedContent = content !== undefined ? content : lastSavedContent;

            updateSaveStatus("saved", "Saved");
        } catch (error) {
            updateSaveStatus("error", "Failed to save");
        } finally {
            isSaving = false;
        }
    }

    const scheduleAutosave = debounce(() => {
        const currentTitle = titleInput.value.trim();
        const currentContent = contentTextarea.value;

        if (currentTitle !== lastSavedTitle || currentContent !== lastSavedContent) {
            save(currentTitle, currentContent);
        }
    }, 2000);

    updatePreview();

    contentTextarea.addEventListener("input", () => {
        updatePreview();
        scheduleAutosave();
    });

    titleInput.addEventListener("input", scheduleAutosave);

    saveBtn.addEventListener("click", () => {
        save(titleInput.value.trim(), contentTextarea.value);
    });

    if (publishBtn) {
        publishBtn.addEventListener("click", async () => {
            await save(titleInput.value.trim(), contentTextarea.value);

            try {
                await api.post(`${apiUrl}/publish`);
                window.location.reload();
            } catch (error) {
                alert(error.message || "Failed to publish");
            }
        });
    }

    // Unpublish functionality not yet implemented in service layer
    if (unpublishBtn) {
        unpublishBtn.style.display = "none";
    }

    window.addEventListener("beforeunload", (e) => {
        const hasUnsavedChanges =
            titleInput.value.trim() !== lastSavedTitle ||
            contentTextarea.value !== lastSavedContent;

        if (hasUnsavedChanges) {
            e.preventDefault();
            e.returnValue = "";
        }
    });

    document.addEventListener("keydown", (e) => {
        if ((e.ctrlKey || e.metaKey) && e.key === "s") {
            e.preventDefault();
            save(titleInput.value.trim(), contentTextarea.value);
        }
    });
});
