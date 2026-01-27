import api from "../core/api.js";
import { $ } from "../core/dom.js";
import { showError, hideError } from "../core/utils.js";
import { createUnifiedEditor } from "../core/unified-editor.js";

document.addEventListener("DOMContentLoaded", () => {
    const { courseId, moduleId } = window.CONTENT_NEW_DATA || {};
    if (!courseId || !moduleId) return;

    const titleInput = $("#content-title");
    const typeSelect = $("#content-type");
    const saveBtn = $("#save-btn");
    const errorContainer = $("#content-new-error");
    const titleErrorEl = $("#title-error");
    const editorContainer = document.getElementById("lecture-editor");
    const formatLabel = document.getElementById("editor-format-label");

    let isSaving = false;
    let navigatingAfterCreate = false;
    let unifiedEditor = null;
    const initialTitle = titleInput.value.trim();
    let initialBody = "";

    try {
        unifiedEditor = createUnifiedEditor(editorContainer, {
            initialValue: "",
            initialFormat: "markdown",
            placeholder: "Write your content here...",
            onFormatChange: (newFormat) => {
                const names = { markdown: "Markdown", plain: "Plain Text", html: "Rich Text" };
                if (formatLabel) formatLabel.textContent = names[newFormat] || newFormat;
            },
            onUpdate: () => hideError(errorContainer),
        });
        initialBody = unifiedEditor.getValue();
    } catch (error) {
        console.error("Failed to initialize editor:", error);
        showError("Failed to load editor. Please refresh the page.", errorContainer);
    }

    async function getNextOrder() {
        const response = await api.get(`/api/courses/${courseId}/modules/${moduleId}/content`);
        const readings = await response.json();
        if (!readings || readings.length === 0) return 0;
        return Math.max(...readings.map((r) => r.order || 0)) + 1;
    }

    async function createContent() {
        if (isSaving) return;

        const title = titleInput.value.trim();
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
        saveBtn.textContent = "Creating...";

        try {
            const content = unifiedEditor ? unifiedEditor.getValue() : "";
            const format = unifiedEditor ? unifiedEditor.getFormat() : "markdown";
            const order = await getNextOrder();

            let sanitizedContent = content;
            if (format === "html" && typeof DOMPurify !== "undefined") {
                sanitizedContent = DOMPurify.sanitize(content);
            }

            await api.post(`/api/courses/${courseId}/modules/${moduleId}/content`, {
                type: typeSelect ? typeSelect.value.trim() : "reading",
                title,
                order,
                format,
                content: sanitizedContent,
            });

            navigatingAfterCreate = true;
            window.location.href = `/courses/${courseId}/modules/${moduleId}`;
        } catch (error) {
            showError(error.message || "Failed to create content", errorContainer);
        } finally {
            isSaving = false;
            saveBtn.disabled = false;
            saveBtn.textContent = "Create";
        }
    }

    titleInput.addEventListener("input", () => {
        hideError(errorContainer);
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
        if (titleInput.value.trim() !== initialTitle || currentBody !== initialBody) {
            e.preventDefault();
            e.returnValue = "";
        }
    });
});
