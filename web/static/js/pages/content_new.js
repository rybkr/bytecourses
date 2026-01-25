import api from "../core/api.js";
import { $ } from "../core/dom.js";

document.addEventListener("DOMContentLoaded", () => {
    const { courseId, moduleId } = window.CONTENT_NEW_DATA || {};

    if (!courseId || !moduleId) return;

    const contentTypeCard = $("#content-type-reading");
    const titleInput = $("#content-title");
    const contentTextarea = $("#content-body");
    const previewDiv = $("#content-preview");
    const saveBtn = $("#save-btn");

    let isSaving = false;
    let selectedContentType = "reading";

    function updatePreview() {
        if (typeof marked !== "undefined") {
            const html = marked.parse(contentTextarea.value || "");
            previewDiv.querySelector(".proposal-content-value").innerHTML =
                html;
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
            titleInput.focus();
            return;
        }

        isSaving = true;
        saveBtn.disabled = true;
        saveBtn.textContent = "Creating...";

        try {
            const order = await getNextReadingOrder();
            const response = await api.post(
                `/api/courses/${courseId}/modules/${moduleId}/content`,
                {
                    type: contentType,
                    title: title,
                    order: order,
                    format: "markdown",
                    content: content,
                },
            );
            const reading = await response.json();
            window.location.href = `/courses/${courseId}/content?readingId=${reading.id}`;
        } catch (error) {
            alert(error.message || "Failed to create content");
            saveBtn.disabled = false;
            saveBtn.textContent = "Create";
            isSaving = false;
        }
    }

    if (contentTypeCard) {
        contentTypeCard.addEventListener("click", () => {
            document.querySelectorAll(".content-type-card").forEach((card) => {
                card.classList.remove("active");
            });
            contentTypeCard.classList.add("active");
            selectedContentType = contentTypeCard.dataset.type;
        });
    }

    updatePreview();

    contentTextarea.addEventListener("input", () => {
        updatePreview();
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
});
