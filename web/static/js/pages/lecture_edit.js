import api from "../core/api.js";
import { debounce, showError, hideError } from "../core/utils.js";
import { $ } from "../core/dom.js";
import { updateMarkdownPreview } from "../core/markdown.js";
import { createUnifiedEditor } from "../core/unified-editor.js";
import { createResizer } from "../core/resizer.js";

document.addEventListener("DOMContentLoaded", () => {
    const {
        courseId,
        moduleId,
        readingId,
        order,
        format: initialFormat,
        initialContent: initialContentFromData,
    } = window.LECTURE_DATA || {};

    if (!courseId || !moduleId || !readingId) return;

    const titleInput = $("#lecture-title");
    const formatSelect = $("#lecture-format");
    const previewDiv = $("#lecture-preview");
    const saveStatus = $("#save-status");
    const saveBtn = $("#save-btn");
    const publishBtn = $("#publish-btn");
    const unpublishBtn = $("#unpublish-btn");
    const errorContainer = $("#lecture-error");
    const editorContainer = document.getElementById("lecture-editor");
    const formatLabel = document.getElementById("editor-format-label");

    let lastSavedTitle = titleInput.value;
    let unifiedEditor = null;
    let lastSavedContent = initialContentFromData || "";
    let lastSavedFormat = initialFormat || "markdown";
    let isSaving = false;
    let currentFormat = initialFormat || "markdown";

    const apiUrl = `/api/courses/${courseId}/modules/${moduleId}/content/${readingId}`;

    function updatePreviewForFormat(content, format) {
        const valueEl = previewDiv?.querySelector(".proposal-content-value");
        if (!valueEl) return;

        if (!content || content.trim() === "") {
            valueEl.innerHTML = "";
            return;
        }

        if (format === "markdown") {
            updateMarkdownPreview(content, previewDiv, {
                wrapperClass: "proposal-content-value",
            });
        } else if (format === "html") {
            const sanitized = typeof DOMPurify !== "undefined" ? DOMPurify.sanitize(content) : content;
            valueEl.innerHTML = sanitized;
        } else {
            const pre = document.createElement("pre");
            pre.style.whiteSpace = "pre-wrap";
            pre.textContent = content;
            valueEl.innerHTML = "";
            valueEl.appendChild(pre);
        }
    }

    const debouncedUpdatePreview = debounce((content, format) => {
        updatePreviewForFormat(content, format);
    }, 300);

    function updatePreview(content, format) {
        debouncedUpdatePreview(content, format || currentFormat);
    }

    try {
        unifiedEditor = createUnifiedEditor(editorContainer, {
            initialValue: initialContentFromData || "",
            initialFormat: currentFormat,
            placeholder: "Write your content here...",
            previewElement: previewDiv,
            editorContainer: editorContainer,
            onFormatChange: async (newFormat, content) => {
                currentFormat = newFormat;
                const formatNames = {
                    markdown: "Markdown",
                    plain: "Plain Text",
                    html: "Rich Text",
                };
                if (formatLabel) {
                    formatLabel.textContent = formatNames[newFormat] || newFormat;
                }
                updatePreview(content, newFormat);
            },
            onUpdate: (content) => {
                updatePreview(content, currentFormat);
            },
        });

        const leftPane = document.querySelector(".lecture-editor-pane:first-child");
        const rightPane = document.querySelector(".lecture-editor-pane:last-child");
        const resizer = document.getElementById("editor-resizer");
        if (resizer && leftPane && rightPane) {
            createResizer(resizer, leftPane, rightPane, {
                storageKey: "unified-editor-split",
                defaultRatio: 0.5,
                minLeftWidth: 200,
                minRightWidth: 200,
            });
        }

        const togglePreviewBtn = document.getElementById("toggle-preview-btn");
        const toggleMarkdownBtn = document.getElementById("toggle-markdown-btn");
        const storageKey = "unified-editor-mode";

        function setEditorMode(mode) {
            if (!editorContainer) return;
            if (currentFormat !== "markdown") return;

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
        if (currentFormat === "markdown") {
            setEditorMode(savedMode);
        } else {
            const previewPane = document.getElementById("preview-pane");
            const togglePreviewBtn = document.getElementById("toggle-preview-btn");
            const toggleMarkdownBtn = document.getElementById("toggle-markdown-btn");
            if (previewPane) previewPane.style.display = "none";
            if (togglePreviewBtn) togglePreviewBtn.style.display = "none";
            if (toggleMarkdownBtn) toggleMarkdownBtn.style.display = "none";
        }

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

        if (formatSelect) {
            formatSelect.addEventListener("change", async (e) => {
                const newFormat = e.target.value;
                await unifiedEditor.setFormat(newFormat);
                currentFormat = newFormat;
                const previewPane = document.getElementById("preview-pane");
                const togglePreviewBtn = document.getElementById("toggle-preview-btn");
                const toggleMarkdownBtn = document.getElementById("toggle-markdown-btn");

                if (newFormat === "markdown") {
                    if (previewPane) previewPane.style.display = "";
                    if (togglePreviewBtn) togglePreviewBtn.style.display = "";
                    if (toggleMarkdownBtn) toggleMarkdownBtn.style.display = "";
                } else {
                    if (previewPane) previewPane.style.display = "none";
                    if (togglePreviewBtn) togglePreviewBtn.style.display = "none";
                    if (toggleMarkdownBtn) toggleMarkdownBtn.style.display = "none";
                    setEditorMode("markdown");
                }
            });
        }

        lastSavedContent = unifiedEditor.getValue();
        updatePreview(lastSavedContent, currentFormat);
    } catch (error) {
        console.error("Failed to initialize editor:", error);
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

    async function save(title, content, format) {
        if (isSaving || !unifiedEditor) return;

        const currentTitle = title !== undefined ? title : titleInput.value.trim();
        const currentContent =
            content !== undefined ? content : unifiedEditor.getValue();
        const currentFormatValue = format !== undefined ? format : unifiedEditor.getFormat();
        const hasChanges =
            currentTitle !== lastSavedTitle ||
            currentContent !== lastSavedContent ||
            currentFormatValue !== lastSavedFormat;

        if (!hasChanges) {
            updateSaveStatus("saved", "Saved");
            return;
        }

        isSaving = true;
        updateSaveStatus("saving", "Saving...");
        hideError(errorContainer);

        try {
            let sanitizedContent = currentContent;
            if (currentFormatValue === "html" && typeof DOMPurify !== "undefined") {
                sanitizedContent = DOMPurify.sanitize(currentContent);
            }

            await api.patch(apiUrl, {
                type: "reading",
                title: currentTitle,
                order: order ?? 0,
                format: currentFormatValue,
                content: sanitizedContent,
            });

            lastSavedTitle = currentTitle;
            lastSavedContent = currentContent;
            lastSavedFormat = currentFormatValue;
            updateSaveStatus("saved", "Saved");
        } catch (error) {
            updateSaveStatus("error", "Failed to save");
            showError(error.message || "Failed to save", errorContainer);
        } finally {
            isSaving = false;
        }
    }

    const scheduleAutosave = debounce(() => {
        if (!unifiedEditor) return;
        const currentTitle = titleInput.value.trim();
        const currentContent = unifiedEditor.getValue();
        const currentFormatValue = unifiedEditor.getFormat();

        if (
            currentTitle !== lastSavedTitle ||
            currentContent !== lastSavedContent ||
            currentFormatValue !== lastSavedFormat
        ) {
            save(currentTitle, currentContent, currentFormatValue);
        }
    }, 2000);

    titleInput.addEventListener("input", scheduleAutosave);

    saveBtn.addEventListener("click", () => {
        if (unifiedEditor) {
            save(
                titleInput.value.trim(),
                unifiedEditor.getValue(),
                unifiedEditor.getFormat(),
            );
        }
    });

    if (publishBtn) {
        publishBtn.addEventListener("click", async () => {
            if (unifiedEditor) {
                await save(
                    titleInput.value.trim(),
                    unifiedEditor.getValue(),
                    unifiedEditor.getFormat(),
                );
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
            if (unifiedEditor) {
                await save(
                    titleInput.value.trim(),
                    unifiedEditor.getValue(),
                    unifiedEditor.getFormat(),
                );
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
        if (!unifiedEditor) return;
        const currentContent = unifiedEditor.getValue();
        const currentFormatValue = unifiedEditor.getFormat();
        const hasUnsavedChanges =
            titleInput.value.trim() !== lastSavedTitle ||
            currentContent !== lastSavedContent ||
            currentFormatValue !== lastSavedFormat;

        if (hasUnsavedChanges) {
            e.preventDefault();
            e.returnValue = "";
        }
    });
});
