import api from "../core/api.js";
import FormHandler from "../components/FormHandler.js";
import { $, $$ } from "../core/dom.js";
import { showError, hideError } from "../core/utils.js";

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

    submitBtn.addEventListener("click", (e) => {
        e.preventDefault();
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
            navigator.sendBeacon(`/api/proposals/${id}`, JSON.stringify(payload));
        } catch (_) {}
    });

    let currentTooltip = null;
    let tooltipCloseHandler = null;
    let tooltipRepositionHandlers = [];

    function closeTooltip() {
        if (currentTooltip) {
            const helpIcon = currentTooltip.icon;
            const tooltip = currentTooltip.element;
            if (tooltip && tooltip.parentNode) {
                tooltip.parentNode.removeChild(tooltip);
            }
            if (helpIcon) {
                helpIcon.classList.remove("active");
                helpIcon.setAttribute("aria-expanded", "false");
            }
            currentTooltip = null;
        }
        if (tooltipCloseHandler) {
            document.removeEventListener("click", tooltipCloseHandler);
            document.removeEventListener("focusin", tooltipCloseHandler);
            tooltipCloseHandler = null;
        }
        tooltipRepositionHandlers.forEach((h) => {
            window.removeEventListener("scroll", h, true);
            window.removeEventListener("resize", h);
        });
        tooltipRepositionHandlers = [];
    }

    function positionTooltip(tooltip, icon) {
        const iconRect = icon.getBoundingClientRect();
        const tooltipRect = tooltip.getBoundingClientRect();
        const viewportWidth = window.innerWidth;
        const viewportHeight = window.innerHeight;
        const padding = 16;

        let top = iconRect.bottom + 8;

        if (top + tooltipRect.height > viewportHeight - padding) {
            const topPosition = iconRect.top - tooltipRect.height - 8;
            if (topPosition >= padding) {
                top = topPosition;
            } else {
                top = Math.max(padding, Math.min(viewportHeight - tooltipRect.height - padding, top));
            }
        }

        let left = iconRect.left;
        if (left + tooltipRect.width > viewportWidth - padding) {
            left = viewportWidth - tooltipRect.width - padding;
        }
        if (left < padding) {
            left = padding;
        }

        tooltip.style.top = `${top}px`;
        tooltip.style.left = `${left}px`;
    }

    function showTooltip(fieldName) {
        const helpIcon = $(`.help-icon[data-help="${fieldName}"]`);
        const helpPanel = $(`#help-${fieldName}`);

        if (!helpIcon || !helpPanel) return;

        closeTooltip();

        const tooltip = helpPanel.cloneNode(true);
        tooltip.id = `tooltip-${fieldName}`;
        tooltip.removeAttribute("style");
        tooltip.style.display = "block";
        tooltip.style.position = "fixed";
        tooltip.style.background = "#ffffff";
        tooltip.style.zIndex = "10000";
        document.body.appendChild(tooltip);

        positionTooltip(tooltip, helpIcon);

        helpIcon.classList.add("active");
        helpIcon.setAttribute("aria-expanded", "true");

        currentTooltip = {
            element: tooltip,
            icon: helpIcon,
            fieldName: fieldName,
        };

        const reposition = () => {
            if (currentTooltip && currentTooltip.element) {
                positionTooltip(currentTooltip.element, helpIcon);
            }
        };
        window.addEventListener("scroll", reposition, true);
        window.addEventListener("resize", reposition);
        tooltipRepositionHandlers.push(reposition);

        tooltipCloseHandler = function (e) {
            const clickedHelpIcon = e.target.closest(".help-icon");
            const clickedTooltip = e.target.closest(".help-panel");

            if (!clickedHelpIcon && !clickedTooltip) {
                closeTooltip();
            }
        };

        setTimeout(() => {
            document.addEventListener("click", tooltipCloseHandler);
            document.addEventListener("focusin", tooltipCloseHandler);
        }, 10);
    }

    $$(".help-icon").forEach((icon) => {
        icon.setAttribute("aria-expanded", "false");
        icon.addEventListener("click", (e) => {
            e.preventDefault();
            e.stopPropagation();
            const fieldName = icon.getAttribute("data-help");
            if (fieldName) {
                const currentFieldName = currentTooltip?.fieldName;
                if (currentFieldName === fieldName) {
                    closeTooltip();
                } else {
                    showTooltip(fieldName);
                }
            }
        });
        icon.addEventListener("keydown", (e) => {
            if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                e.stopPropagation();
                const fieldName = icon.getAttribute("data-help");
                if (fieldName) {
                    const currentFieldName = currentTooltip?.fieldName;
                    if (currentFieldName === fieldName) {
                        closeTooltip();
                    } else {
                        showTooltip(fieldName);
                    }
                }
            }
        });
    });
});
