import api from "../core/api.js";
import FormHandler from "../components/FormHandler.js";
import HelpTooltip from "../components/HelpTooltip.js";
import { $ } from "../core/dom.js";
import { showError, hideError, confirmAction } from "../core/utils.js";

document.addEventListener("DOMContentLoaded", () => {
    const form = $("#proposal-form");
    if (!form) return;

    const proposalId = Number(form.dataset.proposalId);
    const isNewProposal = !Number.isFinite(proposalId) || proposalId <= 0;
    const saveDelay = Number(form.dataset.autosaveDelay) || 2000;

    const errorDiv = $("#error-message");
    const submitBtn = $("#submitBtn");
    const saveDraftBtn = $("#saveDraftBtn");

    const fieldIds = [
        "title",
        "summary",
        "qualifications",
        "target_audience",
        "learning_objectives",
        "outline",
        "assumed_prerequisites",
    ];

    form.addEventListener("submit", (e) => e.preventDefault());

    const handler = new FormHandler("#proposal-form", {
        apiPath: "/api/proposals",
        entityId: isNewProposal ? null : proposalId,
        autosaveDelay: saveDelay,
        fieldIds: fieldIds,
        errorContainer: "#error-message",
        statusContainer: "#save-status",
        onEntityCreated: (id) => {
            form.dataset.proposalId = id.toString();
        },
    });

    async function submit() {
        const id = await handler.ensureCreated();
        if (!id) {
            showError("Failed to create proposal", errorDiv);
            return;
        }

        hideError(errorDiv);

        try {
            await api.post(`/api/proposals/${id}/actions/submit`);
            window.location.href = `/proposals/${id}`;
        } catch (e) {
            showError(e.message || "Submit failed", errorDiv);
        }
    }

    async function saveDraftAndExit() {
        const id = await handler.ensureCreated();
        if (!id) {
            showError("Failed to create proposal", errorDiv);
            return;
        }

        window.location.href = `/proposals/${id}`;
    }

    submitBtn.addEventListener("click", async (e) => {
        e.preventDefault();
        const confirmed = await confirmAction(
            "Once submitted, your proposal will be sent to administrators for review. You won't be able to edit it until they respond.",
            {
                title: "Submit for Review?",
                confirmText: "Submit",
                confirmButtonClass: "btn-primary",
                variant: "info",
            }
        );

        if (!confirmed) {
            return;
        }

        submit().catch(() => {
            showError("Submit failed", errorDiv);
        });
    });

    saveDraftBtn.addEventListener("click", (e) => {
        e.preventDefault();
        saveDraftAndExit().catch(() => {
            showError("Save failed", errorDiv);
        });
    });

    document.addEventListener("visibilitychange", () => {
        if (document.visibilityState !== "hidden") return;

        const id = handler.getEntityId();
        if (id === null) return;

        const payload = handler.readPayload();
        try {
            navigator.sendBeacon(
                `/api/proposals/${id}`,
                JSON.stringify(payload),
            );
        } catch (_) {}
    });

    new HelpTooltip();
});
