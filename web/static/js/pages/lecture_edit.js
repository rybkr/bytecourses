import api from "../core/api.js";
import { debounce, showError, hideError } from "../core/utils.js";
import { $ } from "../core/dom.js";
import { updateMarkdownPreview } from "../core/markdown.js";
import {
    createMarkdownEditor,
    setupScrollSync,
    addCustomShortcut,
} from "../core/markdown-editor.js";
import { createResizer } from "../core/resizer.js";

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

    const debouncedUpdatePreview = debounce((content) => {
        updateMarkdownPreview(content, previewDiv, {
            wrapperClass: "proposal-content-value",
        });
    }, 300);

    function updatePreview(content) {
        debouncedUpdatePreview(content);
    }

    try {
        const editorContainer = document.getElementById("lecture-editor");
        markdownEditor = createMarkdownEditor(contentTextarea, {
            initialValue: initialContentFromData || "",
            placeholder: "Write your reading content here using Markdown...",
            lineNumbers: true,
            customPreviewElement: previewDiv,
            editorContainer: editorContainer,
            onUpdate: (content) => {
                updatePreview(content);
            },
        });

        addCustomShortcut(markdownEditor.editor, "Mod-s", () => {
            save(titleInput.value.trim(), markdownEditor.getValue());
        });

        setupScrollSync(markdownEditor.editor, previewDiv);

        const leftPane = document.querySelector(".lecture-editor-pane:first-child");
        const rightPane = document.querySelector(".lecture-editor-pane:last-child");
        const resizer = document.getElementById("editor-resizer");
        if (resizer && leftPane && rightPane) {
            createResizer(resizer, leftPane, rightPane, {
                storageKey: "markdown-editor-split",
                defaultRatio: 0.5,
                minLeftWidth: 200,
                minRightWidth: 200,
            });
        }

        const togglePreviewBtn = document.getElementById("toggle-preview-btn");
        const toggleMarkdownBtn = document.getElementById("toggle-markdown-btn");
        const storageKey = "markdown-editor-mode";

        function setEditorMode(mode) {
            if (!editorContainer) return;

            editorContainer.classList.remove("mode-markdown", "mode-preview", "mode-split");

            const isMobile = window.matchMedia("(max-width: 900px)").matches;

            if (isMobile) {
                if (mode === "preview") {
                    editorContainer.classList.add("mode-preview");
                } else {
                    editorContainer.classList.add("mode-markdown");
                }
            } else {
                editorContainer.classList.add("mode-split");
            }

            try {
                localStorage.setItem(storageKey, mode);
            } catch (e) {
                console.warn("Failed to save editor mode", e);
            }
        }

        function getEditorMode() {
            try {
                return localStorage.getItem(storageKey) || "markdown";
            } catch (e) {
                return "markdown";
            }
        }

        const savedMode = getEditorMode();
        setEditorMode(savedMode);

        if (togglePreviewBtn) {
            togglePreviewBtn.addEventListener("click", () => {
                setEditorMode("preview");
            });
        }

        if (toggleMarkdownBtn) {
            toggleMarkdownBtn.addEventListener("click", () => {
                setEditorMode("markdown");
            });
        }

        let resizeTimeout;
        window.addEventListener("resize", () => {
            clearTimeout(resizeTimeout);
            resizeTimeout = setTimeout(() => {
                const isMobile = window.matchMedia("(max-width: 900px)").matches;
                const currentMode = getEditorMode();
                setEditorMode(isMobile ? currentMode : "split");
            }, 100);
        });

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
