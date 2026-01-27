import api from "../core/api.js";
import { $ } from "../core/dom.js";
import { showError, hideError, debounce } from "../core/utils.js";
import { updateMarkdownPreview } from "../core/markdown.js";
import { createUnifiedEditor } from "../core/unified-editor.js";
import { createResizer } from "../core/resizer.js";

document.addEventListener("DOMContentLoaded", () => {
    const { courseId, moduleId } = window.CONTENT_NEW_DATA || {};

    if (!courseId || !moduleId) return;

    const titleInput = $("#content-title");
    const formatSelect = $("#content-format");
    const typeSelect = $("#content-type");
    const previewDiv = $("#content-preview");
    const saveBtn = $("#save-btn");
    const errorContainer = $("#content-new-error");
    const titleErrorEl = $("#title-error");
    const editorContainer = document.getElementById("lecture-editor");
    const formatLabel = document.getElementById("editor-format-label");

    let isSaving = false;
    let navigatingAfterCreate = false;
    const initialTitle = titleInput.value.trim();
    let unifiedEditor = null;
    let initialBody = "";
    let currentFormat = formatSelect ? formatSelect.value : "markdown";

    function updatePreviewForFormat(content, format) {
        const placeholder = document.getElementById("content-preview-placeholder");
        const valueEl = previewDiv?.querySelector(".proposal-content-value");

        if (!content || content.trim() === "") {
            if (placeholder) placeholder.style.display = "block";
            if (valueEl) valueEl.innerHTML = "";
            return;
        }

        if (placeholder) placeholder.style.display = "none";

        if (format === "markdown") {
            updateMarkdownPreview(content, previewDiv, {
                placeholderEl: placeholder,
                valueEl: valueEl,
            });
        } else if (format === "html") {
            if (valueEl) {
                const sanitized = typeof DOMPurify !== "undefined" ? DOMPurify.sanitize(content) : content;
                valueEl.innerHTML = sanitized;
            }
        } else {
            if (valueEl) {
                const pre = document.createElement("pre");
                pre.style.whiteSpace = "pre-wrap";
                pre.textContent = content;
                valueEl.innerHTML = "";
                valueEl.appendChild(pre);
            }
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
            initialValue: "",
            initialFormat: currentFormat,
            placeholder: "Write your content here...",
            previewElement: previewDiv,
            editorContainer: editorContainer,
            onFormatChange: (newFormat, content) => {
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
                clearError();
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

        if (formatSelect) {
            formatSelect.addEventListener("change", (e) => {
                const newFormat = e.target.value;
                unifiedEditor.setFormat(newFormat);
            });
        }

        initialBody = unifiedEditor.getValue();
        updatePreview(initialBody, currentFormat);
    } catch (error) {
        console.error("Failed to initialize editor:", error);
        showError("Failed to load editor. Please refresh the page.", errorContainer);
    }

    async function getNextReadingOrder() {
        const response = await api.get(
            `/api/courses/${courseId}/modules/${moduleId}/content`,
        );
        const readings = await response.json();
        if (!readings || readings.length === 0) return 0;
        return Math.max(...readings.map((r) => r.order || 0)) + 1;
    }

    async function createContent() {
        if (isSaving) return;

        const title = titleInput.value.trim();
        const content = unifiedEditor ? unifiedEditor.getValue() : "";
        const contentType = typeSelect ? typeSelect.value.trim() : "reading";
        const format = unifiedEditor ? unifiedEditor.getFormat() : "markdown";

        if (!title) {
            if (titleErrorEl) {
                titleErrorEl.textContent = "Title is required";
                titleErrorEl.classList.remove("hidden");
            }
            titleInput.focus();
            return;
        }

        if (titleErrorEl) {
            titleErrorEl.textContent = "";
            titleErrorEl.classList.add("hidden");
        }
        hideError(errorContainer);
        isSaving = true;
        saveBtn.disabled = true;
        saveBtn.setAttribute("aria-busy", "true");
        saveBtn.textContent = "Creating...";

        try {
            const order = await getNextReadingOrder();
            let sanitizedContent = content;
            if (format === "html" && typeof DOMPurify !== "undefined") {
                sanitizedContent = DOMPurify.sanitize(content);
            }

            const res = await api.post(
                `/api/courses/${courseId}/modules/${moduleId}/content`,
                {
                    type: contentType,
                    title,
                    order,
                    format: format,
                    content: sanitizedContent,
                },
            );

            if (!res) {
                return;
            }

            const reading = await res.json();
            navigatingAfterCreate = true;
            window.location.href = `/courses/${courseId}/modules/${moduleId}`;
        } catch (error) {
            showError(error.message || "Failed to create content", errorContainer);
        } finally {
            isSaving = false;
            saveBtn.disabled = false;
            saveBtn.removeAttribute("aria-busy");
            saveBtn.textContent = "Create";
        }
    }

    function clearError() {
        hideError(errorContainer);
    }

    titleInput.addEventListener("input", () => {
        clearError();
        if (titleErrorEl) {
            titleErrorEl.textContent = "";
            titleErrorEl.classList.add("hidden");
        }
    });

    saveBtn.addEventListener("click", createContent);

    titleInput.addEventListener("keydown", (e) => {
        if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
            e.preventDefault();
            createContent();
        }
    });

    window.addEventListener("beforeunload", (e) => {
        if (navigatingAfterCreate) return;
        const currentBody = unifiedEditor ? unifiedEditor.getValue() : "";
        const dirty =
            titleInput.value.trim() !== initialTitle || currentBody !== initialBody;
        if (dirty) {
            e.preventDefault();
            e.returnValue = "";
        }
    });
});
