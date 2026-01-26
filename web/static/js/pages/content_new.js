import api from "../core/api.js";
import { $ } from "../core/dom.js";
import { showError, hideError } from "../core/utils.js";

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
    const initialBody = contentTextarea.value;

    function updatePreview() {
        const body = contentTextarea.value.trim();
        const placeholder = document.getElementById("content-preview-placeholder");
        const valueEl = previewDiv?.querySelector(".proposal-content-value");

        if (!body) {
            if (placeholder) placeholder.classList.remove("hidden");
            if (valueEl) {
                valueEl.innerHTML = "";
                valueEl.classList.add("hidden");
            }
            return;
        }

        if (placeholder) placeholder.classList.add("hidden");
        if (valueEl) {
            valueEl.classList.remove("hidden");
            if (typeof marked !== "undefined") {
                const raw = marked.parse(body);
                const html =
                    typeof DOMPurify !== "undefined"
                        ? DOMPurify.sanitize(raw)
                        : raw;
                valueEl.innerHTML = html;
            }
        }
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
        const content = contentTextarea.value;
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

    updatePreview();

    function clearError() {
        hideError(errorContainer);
    }

    contentTextarea.addEventListener("input", () => {
        updatePreview();
        clearError();
    });

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

    contentTextarea.addEventListener("keydown", (e) => {
        if ((e.ctrlKey || e.metaKey) && e.key === "s") {
            e.preventDefault();
            createContent();
        }
    });

    window.addEventListener("beforeunload", (e) => {
        if (navigatingAfterCreate) return;
        const dirty =
            titleInput.value.trim() !== initialTitle ||
            contentTextarea.value !== initialBody;
        if (dirty) {
            e.preventDefault();
            e.returnValue = "";
        }
    });
});
