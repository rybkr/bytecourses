import api from "../core/api.js";
import { debounce, showError, hideError } from "../core/utils.js";
import { $ } from "../core/dom.js";
import { updateMarkdownPreview } from "../core/markdown.js";

document.addEventListener("DOMContentLoaded", () => {
    const { courseId, moduleId, readingId, order } = window.LECTURE_DATA || {};

    if (!courseId || !moduleId || !readingId) return;

    const titleInput = $("#lecture-title");
    const contentTextarea = $("#lecture-content");
    const previewDiv = $("#lecture-preview");
    const saveStatus = $("#save-status");
    const saveBtn = $("#save-btn");
    const publishBtn = $("#publish-btn");
    const unpublishBtn = $("#unpublish-btn");
    const errorContainer = $("#lecture-error");

    let lastSavedTitle = titleInput.value;
    let lastSavedContent = contentTextarea.value;
    let isSaving = false;

    const apiUrl = `/api/courses/${courseId}/modules/${moduleId}/content/${readingId}`;

    function updatePreview() {
        updateMarkdownPreview(contentTextarea.value, previewDiv, {
            wrapperClass: "proposal-content-value",
        });
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

        const currentTitle = title !== undefined ? title : titleInput.value.trim();
        const currentContent = content !== undefined ? content : contentTextarea.value;
        const hasChanges =
            currentTitle !== lastSavedTitle || currentContent !== lastSavedContent;

        if (!hasChanges) {
            updateSaveStatus("saved", "Saved");
            return;
        }

        isSaving = true;
        updateSaveStatus("saving", "Saving...");
        hideError(errorContainer);

        try {
            await api.patch(apiUrl, {
                type: "reading",
                title: currentTitle,
                order: order ?? 0,
                format: "markdown",
                content: currentContent,
            });

            lastSavedTitle = currentTitle;
            lastSavedContent = currentContent;
            updateSaveStatus("saved", "Saved");
        } catch (error) {
            updateSaveStatus("error", "Failed to save");
            showError(error.message || "Failed to save", errorContainer);
        } finally {
            isSaving = false;
        }
    }

    const scheduleAutosave = debounce(() => {
        const currentTitle = titleInput.value.trim();
        const currentContent = contentTextarea.value;

        if (
            currentTitle !== lastSavedTitle ||
            currentContent !== lastSavedContent
        ) {
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
                await api.post(`${apiUrl}/actions/publish`);
                window.location.reload();
            } catch (error) {
                showError(error.message || "Failed to publish", errorContainer);
            }
        });
    }

    if (unpublishBtn) {
        unpublishBtn.addEventListener("click", async () => {
            await save(titleInput.value.trim(), contentTextarea.value);

            try {
                await api.post(`${apiUrl}/actions/unpublish`);
                window.location.reload();
            } catch (error) {
                showError(error.message || "Failed to unpublish", errorContainer);
            }
        });
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
