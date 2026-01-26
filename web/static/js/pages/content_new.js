import api from "../core/api.js";
import { $ } from "../core/dom.js";
import { showError, hideError, debounce } from "../core/utils.js";
import { updateMarkdownPreview } from "../core/markdown.js";
import {
    createMarkdownEditor,
    setupScrollSync,
    addCustomShortcut,
} from "../core/markdown-editor.js";
import { createResizer } from "../core/resizer.js";

document.addEventListener("DOMContentLoaded", () => {
    const { courseId, moduleId } = window.CONTENT_NEW_DATA || {};

    if (!courseId || !moduleId) return;

    const titleInput = $("#content-title");
    const contentTextarea = $("#content-body");
    const typeSelect = $("#content-type");
    const previewDiv = $("#content-preview");
    const saveBtn = $("#save-btn");
    const errorContainer = $("#content-new-error");
    const titleErrorEl = $("#title-error");

    let isSaving = false;
    let navigatingAfterCreate = false;
    const initialTitle = titleInput.value.trim();
    let markdownEditor = null;
    let initialBody = "";

    // Debounced preview update
    const debouncedUpdatePreview = debounce((content) => {
        const placeholder = document.getElementById("content-preview-placeholder");
        const valueEl = previewDiv?.querySelector(".proposal-content-value");

        updateMarkdownPreview(content, previewDiv, {
            placeholderEl: placeholder,
            valueEl: valueEl,
        });
    }, 300);

    function updatePreview(content) {
        debouncedUpdatePreview(content);
    }

    // Initialize EasyMDE editor
    try {
        markdownEditor = createMarkdownEditor(contentTextarea, {
            initialValue: "",
            placeholder: "Write your content here using Markdown...",
            lineNumbers: true,
            onUpdate: (content) => {
                updatePreview(content);
                clearError();
            },
        });

        // Add custom keyboard shortcut for Ctrl/Cmd+Enter to save
        addCustomShortcut(markdownEditor.editor, "Mod-Enter", () => {
            createContent();
        });

        // Setup scroll sync
        setupScrollSync(markdownEditor.editor, previewDiv);

        // Initialize resizer
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

        // Initialize editor mode toggle
        const editorContainer = document.getElementById("lecture-editor");
        const togglePreviewBtn = document.getElementById("toggle-preview-btn");
        const toggleMarkdownBtn = document.getElementById("toggle-markdown-btn");
        const storageKey = "markdown-editor-mode";

        function setEditorMode(mode) {
            if (!editorContainer) return;

            // Remove existing mode classes
            editorContainer.classList.remove("mode-markdown", "mode-preview", "mode-split");

            // Check if mobile
            const isMobile = window.matchMedia("(max-width: 900px)").matches;

            if (isMobile) {
                // On mobile, use markdown or preview
                if (mode === "preview") {
                    editorContainer.classList.add("mode-preview");
                } else {
                    editorContainer.classList.add("mode-markdown");
                }
            } else {
                // On desktop, always use split mode
                editorContainer.classList.add("mode-split");
            }

            // Save to localStorage
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

        // Load saved mode or default
        const savedMode = getEditorMode();
        setEditorMode(savedMode);

        // Handle toggle buttons
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

        // Handle window resize to switch modes
        let resizeTimeout;
        window.addEventListener("resize", () => {
            clearTimeout(resizeTimeout);
            resizeTimeout = setTimeout(() => {
                const isMobile = window.matchMedia("(max-width: 900px)").matches;
                const currentMode = getEditorMode();
                setEditorMode(isMobile ? currentMode : "split");
            }, 100);
        });

        initialBody = markdownEditor.getValue();
    } catch (error) {
        console.error("Failed to initialize markdown editor:", error);
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
        const content = markdownEditor ? markdownEditor.getValue() : "";
        const contentType = typeSelect ? typeSelect.value.trim() : "reading";

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
            const res = await api.post(
                `/api/courses/${courseId}/modules/${moduleId}/content`,
                {
                    type: contentType,
                    title,
                    order,
                    format: "markdown",
                    content,
                },
            );

            if (!res) {
                return;
            }

            const reading = await res.json();
            navigatingAfterCreate = true;
            window.location.href = `/courses/${courseId}/modules/${moduleId}`;
        } finally {
            isSaving = false;
            saveBtn.disabled = false;
            saveBtn.removeAttribute("aria-busy");
            saveBtn.textContent = "Create";
        }
    }

    // Initial preview update
    updatePreview("");

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
        const currentBody = markdownEditor ? markdownEditor.getValue() : "";
        const dirty =
            titleInput.value.trim() !== initialTitle || currentBody !== initialBody;
        if (dirty) {
            e.preventDefault();
            e.returnValue = "";
        }
    });
});
