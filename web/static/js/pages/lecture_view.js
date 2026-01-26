import api from "../core/api.js";
import { $ } from "../core/dom.js";

document.addEventListener("DOMContentLoaded", () => {
    const { courseId, moduleId, readingId } = window.LECTURE_VIEW_DATA || {};

    if (!courseId || !moduleId || !readingId) return;

    const apiUrl = `/api/courses/${courseId}/modules/${moduleId}/content/${readingId}`;
    const publishBtn = $("#publish-btn");
    const unpublishBtn = $("#unpublish-btn");

    if (publishBtn) {
        publishBtn.addEventListener("click", async () => {
            try {
                await api.post(`${apiUrl}/actions/publish`);
                window.location.reload();
            } catch (error) {
                alert(error.message || "Failed to publish");
            }
        });
    }

    if (unpublishBtn) {
        unpublishBtn.addEventListener("click", async () => {
            try {
                await api.post(`${apiUrl}/actions/unpublish`);
                window.location.reload();
            } catch (error) {
                alert(error.message || "Failed to unpublish");
            }
        });
    }
});
