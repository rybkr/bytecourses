import api from "../core/api.js";
import FormHandler from "../components/FormHandler.js";
import HelpTooltip from "../components/HelpTooltip.js";
import { $, on } from "../core/dom.js";
import { showError, hideError } from "../core/utils.js";

document.addEventListener("DOMContentLoaded", () => {
    const form = $("#course-form");
    if (!form) return;

    const courseId = Number(form.dataset.courseId);
    if (!Number.isFinite(courseId) || courseId <= 0) return;

    const saveDelay = Number(form.dataset.autosaveDelay) || 2000;
    const errorDiv = $("#error-message");
    const publishBtn = $("#publishBtn");

    const fieldIds = [
        "title",
        "summary",
        "target_audience",
        "learning_objectives",
        "assumed_prerequisites",
    ];

    form.addEventListener("submit", (e) => e.preventDefault());

    const handler = new FormHandler("#course-form", {
        apiPath: "/api/courses",
        entityId: courseId,
        autosaveDelay: saveDelay,
        fieldIds: fieldIds,
        errorContainer: "#error-message",
        statusContainer: "#save-status",
    });

    async function publish() {
        await handler.saveNow();

        hideError(errorDiv);
        publishBtn.disabled = true;

        try {
            await api.post(`/api/courses/${courseId}/actions/publish`);
            window.location.href = `/courses/${courseId}`;
        } catch (error) {
            showError(error.message || "Publish failed", errorDiv);
            publishBtn.disabled = false;
        }
    }

    if (publishBtn) {
        on(publishBtn, "click", (e) => {
            e.preventDefault();
            publish().catch(() => {
                showError("Publish failed", errorDiv);
                publishBtn.disabled = false;
            });
        });
    }

    document.addEventListener("visibilitychange", () => {
        if (document.visibilityState === "hidden") {
            handler.saveNow().catch(() => {});
        }
    });

    new HelpTooltip();
});
