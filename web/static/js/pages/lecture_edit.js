import api from "../core/api.js";
import { debounce, showError, hideError } from "../core/utils.js";
import { $ } from "../core/dom.js";
import { updateMarkdownPreview } from "../core/markdown.js";
import {
    createMarkdownEditor,
    setupScrollSync,
    addCustomShortcut,
} from "../core/markdown-editor.js";

document.addEventListener("DOMContentLoaded", () => {
    const { courseId, moduleId, readingId, order, initialContent: initialContentFromData } = window.LECTURE_DATA || {};

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
    let markdownEditor = null;
    let lastSavedContent = initialContentFromData || "";
    let isSaving = false;

    const apiUrl = `/api/courses/${courseId}/modules/${moduleId}/content/${readingId}`;

    // Debounced preview update
    const debouncedUpdatePreview = debounce((content) => {
        updateMarkdownPreview(content, previewDiv, {
            wrapperClass: "proposal-content-value",
        });
    }, 300);

    function updatePreview(content) {
        debouncedUpdatePreview(content);
    }

    // Initialize EasyMDE editor
    try {
        markdownEditor = createMarkdownEditor(contentTextarea, {
            initialValue: initialContentFromData || "",
            placeholder: "Write your reading content here using Markdown...",
            lineNumbers: true,
            onUpdate: (content) => {
                updatePreview(content);
            },
        });

        // Add custom keyboard shortcut for Ctrl/Cmd+S to save
        addCustomShortcut(markdownEditor.editor, "Mod-s", () => {
            save(titleInput.value.trim(), markdownEditor.getValue());
        });

        // Setup scroll sync
        setupScrollSync(markdownEditor.editor, previewDiv);

        lastSavedContent = markdownEditor.getValue();
        updatePreview(lastSavedContent);
    } catch (error) {
        console.error("Failed to initialize markdown editor:", error);
        showError("Failed to load editor. Please refresh the page.", errorContainer);
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
        if (isSaving || !markdownEditor) return;

        const currentTitle = title !== undefined ? title : titleInput.value.trim();
        const currentContent =
            content !== undefined ? content : markdownEditor.getValue();
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
        if (!markdownEditor) return;
        const currentTitle = titleInput.value.trim();
        const currentContent = markdownEditor.getValue();

        if (
            currentTitle !== lastSavedTitle ||
            currentContent !== lastSavedContent
        ) {
            save(currentTitle, currentContent);
        }
    }, 2000);

    // Setup autosave on title changes
    titleInput.addEventListener("input", scheduleAutosave);

    saveBtn.addEventListener("click", () => {
        if (markdownEditor) {
            save(titleInput.value.trim(), markdownEditor.getValue());
        }
    });

    if (publishBtn) {
        publishBtn.addEventListener("click", async () => {
            if (markdownEditor) {
                await save(titleInput.value.trim(), markdownEditor.getValue());
            }

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
            if (markdownEditor) {
                await save(titleInput.value.trim(), markdownEditor.getValue());
            }

            try {
                await api.post(`${apiUrl}/actions/unpublish`);
                window.location.reload();
            } catch (error) {
                showError(error.message || "Failed to unpublish", errorContainer);
            }
        });
    }

    window.addEventListener("beforeunload", (e) => {
        if (!markdownEditor) return;
        const currentContent = markdownEditor.getValue();
        const hasUnsavedChanges =
            titleInput.value.trim() !== lastSavedTitle ||
            currentContent !== lastSavedContent;

        if (hasUnsavedChanges) {
            e.preventDefault();
            e.returnValue = "";
        }
    });
});
