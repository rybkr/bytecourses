import api from "../core/api.js";
import { $ } from "../core/dom.js";
import { showError, hideError } from "../core/utils.js";
import { createUnifiedEditor } from "../core/unified-editor.js";

const MAX_FILE_SIZE = 50 * 1024 * 1024; // 50 MB

document.addEventListener("DOMContentLoaded", () => {
    const { courseId, moduleId } = window.CONTENT_NEW_DATA || {};
    if (!courseId || !moduleId) return;

    const titleInput = $("#content-title");
    const typeSelect = $("#content-type");
    const saveBtn = $("#save-btn");
    const errorContainer = $("#content-new-error");
    const titleErrorEl = $("#title-error");
    const readingEditor = $("#reading-editor");
    const fileUploader = $("#file-uploader");
    const formatLabel = document.getElementById("editor-format-label");

    const fileInput = $("#file-input");
    const fileDropZone = $("#file-drop-zone");
    const filePreview = $("#file-preview");
    const filePreviewName = $("#file-preview-name");
    const filePreviewSize = $("#file-preview-size");
    const fileRemoveBtn = $("#file-remove-btn");

    let isSaving = false;
    let navigatingAfterCreate = false;
    let unifiedEditor = null;
    let selectedFile = null;
    const initialTitle = titleInput.value.trim();
    let initialBody = "";

    try {
        unifiedEditor = createUnifiedEditor(readingEditor, {
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

    function showContentType(type) {
        if (type === "file") {
            readingEditor.classList.add("hidden");
            fileUploader.classList.remove("hidden");
        } else {
            readingEditor.classList.remove("hidden");
            fileUploader.classList.add("hidden");
        }
    }

    function formatFileSize(bytes) {
        if (bytes < 1024) return bytes + " B";
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
        return (bytes / (1024 * 1024)).toFixed(1) + " MB";
    }

    function setSelectedFile(file) {
        if (!file) {
            selectedFile = null;
            filePreview.classList.add("hidden");
            fileDropZone.classList.remove("hidden");
            return;
        }

        if (file.size > MAX_FILE_SIZE) {
            showError("File is too large. Maximum size is 50 MB.", errorContainer);
            return;
        }

        selectedFile = file;
        filePreviewName.textContent = file.name;
        filePreviewSize.textContent = formatFileSize(file.size);
        filePreview.classList.remove("hidden");
        fileDropZone.classList.add("hidden");
        hideError(errorContainer);

        if (!titleInput.value.trim()) {
            titleInput.value = file.name;
        }
    }

    async function getNextOrder() {
        const response = await api.get(`/api/courses/${courseId}/modules/${moduleId}/content`);
        const readings = await response.json();
        if (!readings || readings.length === 0) return 0;
        return Math.max(...readings.map((r) => r.order || 0)) + 1;
    }

    async function createReading() {
        const title = titleInput.value.trim();
        const content = unifiedEditor ? unifiedEditor.getValue() : "";
        const format = unifiedEditor ? unifiedEditor.getFormat() : "markdown";
        const order = await getNextOrder();

        let sanitizedContent = content;
        if (format === "html" && typeof DOMPurify !== "undefined") {
            sanitizedContent = DOMPurify.sanitize(content);
        }

        await api.post(`/api/courses/${courseId}/modules/${moduleId}/content`, {
            type: "reading",
            title,
            order,
            format,
            content: sanitizedContent,
        });
    }

    async function createFile() {
        if (!selectedFile) {
            showError("Please select a file to upload.", errorContainer);
            return;
        }

        const title = titleInput.value.trim();
        const order = await getNextOrder();

        const formData = new FormData();
        formData.append("file", selectedFile);
        formData.append("title", title);
        formData.append("order", order.toString());

        const response = await fetch(`/api/courses/${courseId}/modules/${moduleId}/content/upload`, {
            method: "POST",
            body: formData,
            credentials: "include",
        });

        if (!response.ok) {
            const data = await response.json().catch(() => ({}));
            throw new Error(data.error || "Failed to upload file");
        }
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
            const contentType = typeSelect ? typeSelect.value.trim() : "reading";

            if (contentType === "file") {
                await createFile();
            } else {
                await createReading();
            }

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

    typeSelect.addEventListener("change", () => {
        showContentType(typeSelect.value);
    });

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

    fileInput.addEventListener("change", (e) => {
        const file = e.target.files?.[0];
        if (file) setSelectedFile(file);
    });

    fileDropZone.addEventListener("click", () => {
        fileInput.click();
    });

    fileDropZone.addEventListener("dragover", (e) => {
        e.preventDefault();
        fileDropZone.classList.add("drag-over");
    });

    fileDropZone.addEventListener("dragleave", () => {
        fileDropZone.classList.remove("drag-over");
    });

    fileDropZone.addEventListener("drop", (e) => {
        e.preventDefault();
        fileDropZone.classList.remove("drag-over");
        const file = e.dataTransfer?.files?.[0];
        if (file) setSelectedFile(file);
    });

    fileRemoveBtn.addEventListener("click", () => {
        setSelectedFile(null);
        fileInput.value = "";
    });

    window.addEventListener("beforeunload", (e) => {
        if (navigatingAfterCreate) return;
        const currentBody = unifiedEditor ? unifiedEditor.getValue() : "";
        const hasChanges = titleInput.value.trim() !== initialTitle ||
                          currentBody !== initialBody ||
                          selectedFile !== null;
        if (hasChanges) {
            e.preventDefault();
            e.returnValue = "";
        }
    });
});
