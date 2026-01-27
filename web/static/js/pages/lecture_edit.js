import api from "../core/api.js";
import { debounce, showError, hideError } from "../core/utils.js";
import { $ } from "../core/dom.js";
import { createUnifiedEditor } from "../core/unified-editor.js";

document.addEventListener("DOMContentLoaded", () => {
    const { courseId, moduleId, readingId, order, format: initialFormat, initialContent } = window.LECTURE_DATA || {};
    if (!courseId || !moduleId || !readingId) return;

    const titleInput = $("#lecture-title");
    const saveStatus = $("#save-status");
    const saveBtn = $("#save-btn");
    const publishBtn = $("#publish-btn");
    const unpublishBtn = $("#unpublish-btn");
    const errorContainer = $("#lecture-error");
    const editorContainer = document.getElementById("lecture-editor");
    const formatLabel = document.getElementById("editor-format-label");

    let lastSavedTitle = titleInput.value;
    let lastSavedContent = initialContent || "";
    let lastSavedFormat = initialFormat || "markdown";
    let isSaving = false;
    let unifiedEditor = null;

    const apiUrl = `/api/courses/${courseId}/modules/${moduleId}/content/${readingId}`;

    function updateSaveStatus(type, text) {
        saveStatus.textContent = text;
        saveStatus.className = `save-status ${type}`;
        if (type === "saved") {
            setTimeout(() => {
                if (saveStatus.textContent === text) saveStatus.textContent = "";
            }, 3000);
        }
    }

    async function save() {
        if (isSaving || !unifiedEditor) return;

        const currentTitle = titleInput.value.trim();
        const currentContent = unifiedEditor.getValue();
        const currentFormat = unifiedEditor.getFormat();

        if (currentTitle === lastSavedTitle && currentContent === lastSavedContent && currentFormat === lastSavedFormat) {
            updateSaveStatus("saved", "Saved");
            return;
        }

        isSaving = true;
        updateSaveStatus("saving", "Saving...");
        hideError(errorContainer);

        try {
            let content = currentContent;
            if (currentFormat === "html" && typeof DOMPurify !== "undefined") {
                content = DOMPurify.sanitize(content);
            }

            await api.patch(apiUrl, {
                type: "reading",
                title: currentTitle,
                order: order ?? 0,
                format: currentFormat,
                content,
            });

            lastSavedTitle = currentTitle;
            lastSavedContent = currentContent;
            lastSavedFormat = currentFormat;
            updateSaveStatus("saved", "Saved");
        } catch (error) {
            updateSaveStatus("error", "Failed to save");
            showError(error.message || "Failed to save", errorContainer);
        } finally {
            isSaving = false;
        }
    }

    const scheduleAutosave = debounce(save, 2000);

    try {
        unifiedEditor = createUnifiedEditor(editorContainer, {
            initialValue: initialContent || "",
            initialFormat: initialFormat || "markdown",
            placeholder: "Write your content here...",
            onFormatChange: (newFormat) => {
                const names = { markdown: "Markdown", plain: "Plain Text", html: "Rich Text" };
                if (formatLabel) formatLabel.textContent = names[newFormat] || newFormat;
                scheduleAutosave();
            },
            onUpdate: scheduleAutosave,
        });
    } catch (error) {
        console.error("Failed to initialize editor:", error);
        showError("Failed to load editor. Please refresh the page.", errorContainer);
    }

    titleInput.addEventListener("input", scheduleAutosave);
    saveBtn.addEventListener("click", save);

    publishBtn?.addEventListener("click", async () => {
        await save();
        try {
            await api.post(`${apiUrl}/actions/publish`);
            window.location.reload();
        } catch (error) {
            showError(error.message || "Failed to publish", errorContainer);
        }
    });

    unpublishBtn?.addEventListener("click", async () => {
        await save();
        try {
            await api.post(`${apiUrl}/actions/unpublish`);
            window.location.reload();
        } catch (error) {
            showError(error.message || "Failed to unpublish", errorContainer);
        }
    });

    window.addEventListener("beforeunload", (e) => {
        if (!unifiedEditor) return;
        const hasChanges =
            titleInput.value.trim() !== lastSavedTitle ||
            unifiedEditor.getValue() !== lastSavedContent ||
            unifiedEditor.getFormat() !== lastSavedFormat;
        if (hasChanges) {
            e.preventDefault();
            e.returnValue = "";
        }
    });
});
