import api from "../core/api.js";
import { $, on } from "../core/dom.js";
import { showErrorToast } from "../components/Toast.js";

document.addEventListener("DOMContentLoaded", () => {
    const pathMatch = window.location.pathname.match(
        /\/courses\/(\d+)\/modules\/(\d+)/,
    );

    if (!pathMatch) return;

    const courseId = Number(pathMatch[1]);
    const moduleId = Number(pathMatch[2]);

    if (!courseId || !moduleId) return;

    const publishModuleBtn = $(".publish-module-btn");
    if (publishModuleBtn) {
        publishModuleBtn.addEventListener("click", async () => {
            try {
                await api.post(
                    `/api/courses/${courseId}/modules/${moduleId}/actions/publish`,
                );
                window.location.reload();
            } catch (err) {
                showErrorToast(err.message || "Failed to publish module");
            }
        });
    }

    document.addEventListener("click", async (e) => {
        const pub = e.target.closest(".module-card-publish-btn");
        const unpub = e.target.closest(".module-card-unpublish-btn");
        const card = e.target.closest(".module-reading-card");
        if (!card) return;
        const cid = card.dataset.courseId;
        const mid = card.dataset.moduleId;
        const rid = card.dataset.readingId;
        if (!cid || !mid || !rid) return;

        const base = `/api/courses/${cid}/modules/${mid}/content/${rid}`;
        if (pub) {
            e.preventDefault();
            e.stopPropagation();
            try {
                await api.post(`${base}/actions/publish`);
                window.location.reload();
            } catch (err) {
                showErrorToast(err.message || "Failed to publish");
            }
            return;
        }
        if (unpub) {
            e.preventDefault();
            e.stopPropagation();
            try {
                await api.post(`${base}/actions/unpublish`);
                window.location.reload();
            } catch (err) {
                showErrorToast(err.message || "Failed to unpublish");
            }
        }
    });
});
