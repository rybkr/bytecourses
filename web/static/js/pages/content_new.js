import api from "../core/api.js";
import { $ } from "../core/dom.js";
import { showError, hideError } from "../core/utils.js";

document.addEventListener("DOMContentLoaded", () => {
    const { courseId, moduleId } = window.CONTENT_NEW_DATA || {};

    if (!courseId || !moduleId) return;

    const titleInput = $("#content-title");
    const contentTextarea = $("#content-body");
    const previewDiv = $("#content-preview");
    const saveBtn = $("#save-btn");
    const errorContainer = $("#content-new-error");
    const titleErrorEl = $("#title-error");

    let isSaving = false;
    let selectedContentType = "reading";
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
        const contentType = selectedContentType;

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
            const res = await fetch(
                `/api/courses/${courseId}/modules/${moduleId}/content`,
                {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        type: contentType,
                        title,
                        order,
                        format: "markdown",
                        content,
                    }),
                },
            );

            if (res.status === 401) {
                const next = encodeURIComponent(
                    window.location.pathname + window.location.search,
                );
                window.location.href = `/login?next=${next}`;
                return;
            }

            if (!res.ok) {
                const ct = res.headers.get("Content-Type") || "";
                let msg = "Failed to create content";
                if (ct.includes("application/json")) {
                    try {
                        const j = await res.json();
                        if (j.error) msg = j.error;
                        if (j.errors && Array.isArray(j.errors) && j.errors.length) {
                            const parts = j.errors.map((e) => {
                                const m = e.Message || e.message || String(e);
                                return e.Field ? `${e.Field}: ${m}` : m;
                            });
                            if (parts.length) msg = parts.join("; ");
                        }
                    } catch (_) {}
                } else {
                    const text = await res.text();
                    if (text) msg = text;
                }
                showError(msg, errorContainer);
                return;
            }

            const reading = await res.json();
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
        const dirty =
            titleInput.value.trim() !== initialTitle ||
            contentTextarea.value !== initialBody;
        if (dirty) {
            e.preventDefault();
            e.returnValue = "";
        }
    });
});
