import api from "../core/api.js";
import { $, on } from "../core/dom.js";

document.addEventListener("DOMContentLoaded", () => {
    const pathMatch = window.location.pathname.match(
        /\/courses\/(\d+)\/modules\/(\d+)/,
    );

    if (!pathMatch) return;

    const courseId = Number(pathMatch[1]);
    const moduleId = Number(pathMatch[2]);

    if (!courseId || !moduleId) return;

    function showToast(message, type = "info") {
        const existing = document.querySelector(".toast");
        if (existing) existing.remove();

        const toast = document.createElement("div");
        toast.className = `toast toast-${type}`;
        toast.textContent = message;
        document.body.appendChild(toast);

        requestAnimationFrame(() => {
            toast.classList.add("show");
        });

        setTimeout(() => {
            toast.classList.remove("show");
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }

    const publishModuleBtn = $(".publish-module-btn");
    if (publishModuleBtn) {
        publishModuleBtn.addEventListener("click", async () => {
            try {
                await api.post(
                    `/api/courses/${courseId}/modules/${moduleId}/actions/publish`,
                );
                window.location.reload();
            } catch (err) {
                showToast(err.message || "Failed to publish module", "error");
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
                showToast(err.message || "Failed to publish", "error");
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
                showToast(err.message || "Failed to unpublish", "error");
            }
        }
    });
});
