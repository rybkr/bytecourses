import api from "../core/api.js";
import { $ } from "../core/dom.js";
import { showError, hideError, debounce } from "../core/utils.js";
import { updateMarkdownPreview } from "../core/markdown.js";
import {
    createMarkdownEditor,
    setupScrollSync,
    addCustomShortcut,
} from "../core/markdown-editor.js";

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
