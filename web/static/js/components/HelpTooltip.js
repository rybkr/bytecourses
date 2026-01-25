import { $, $$ } from "../core/dom.js";

export default class HelpTooltip {
    constructor(containerSelector = "body") {
        this.container = $(containerSelector) || document.body;
        this.currentTooltip = null;
        this.tooltipCloseHandler = null;
        this.tooltipRepositionHandlers = [];

        this.init();
    }

    init() {
        $$(".help-icon").forEach((icon) => {
            icon.setAttribute("aria-expanded", "false");
            icon.addEventListener("click", (e) => {
                e.preventDefault();
                e.stopPropagation();
                const fieldName = icon.getAttribute("data-help");
                if (fieldName) {
                    const currentFieldName = this.currentTooltip?.fieldName;
                    if (currentFieldName === fieldName) {
                        this.close();
                    } else {
                        this.show(fieldName);
                    }
                }
            });
            icon.addEventListener("keydown", (e) => {
                if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    e.stopPropagation();
                    const fieldName = icon.getAttribute("data-help");
                    if (fieldName) {
                        const currentFieldName = this.currentTooltip?.fieldName;
                        if (currentFieldName === fieldName) {
                            this.close();
                        } else {
                            this.show(fieldName);
                        }
                    }
                }
            });
        });
    }

    close() {
        if (this.currentTooltip) {
            const helpIcon = this.currentTooltip.icon;
            const tooltip = this.currentTooltip.element;
            if (tooltip && tooltip.parentNode) {
                tooltip.parentNode.removeChild(tooltip);
            }
            if (helpIcon) {
                helpIcon.classList.remove("active");
                helpIcon.setAttribute("aria-expanded", "false");
            }
            this.currentTooltip = null;
        }
        if (this.tooltipCloseHandler) {
            document.removeEventListener("click", this.tooltipCloseHandler);
            document.removeEventListener("focusin", this.tooltipCloseHandler);
            this.tooltipCloseHandler = null;
        }
        this.tooltipRepositionHandlers.forEach((h) => {
            window.removeEventListener("scroll", h, true);
            window.removeEventListener("resize", h);
        });
        this.tooltipRepositionHandlers = [];
    }

    positionTooltip(tooltip, icon) {
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
                top = Math.max(
                    padding,
                    Math.min(
                        viewportHeight - tooltipRect.height - padding,
                        top,
                    ),
                );
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

    show(fieldName) {
        const helpIcon = $(`.help-icon[data-help="${fieldName}"]`);
        const helpPanel = $(`#help-${fieldName}`);

        if (!helpIcon || !helpPanel) return;

        this.close();

        const tooltip = helpPanel.cloneNode(true);
        tooltip.id = `tooltip-${fieldName}`;
        tooltip.removeAttribute("style");
        tooltip.style.display = "block";
        tooltip.style.position = "fixed";
        tooltip.style.background = "#ffffff";
        tooltip.style.zIndex = "10000";
        document.body.appendChild(tooltip);

        this.positionTooltip(tooltip, helpIcon);

        helpIcon.classList.add("active");
        helpIcon.setAttribute("aria-expanded", "true");

        this.currentTooltip = {
            element: tooltip,
            icon: helpIcon,
            fieldName: fieldName,
        };

        const reposition = () => {
            if (this.currentTooltip && this.currentTooltip.element) {
                this.positionTooltip(this.currentTooltip.element, helpIcon);
            }
        };
        window.addEventListener("scroll", reposition, true);
        window.addEventListener("resize", reposition);
        this.tooltipRepositionHandlers.push(reposition);

        this.tooltipCloseHandler = (e) => {
            const clickedHelpIcon = e.target.closest(".help-icon");
            const clickedTooltip = e.target.closest(".help-panel");

            if (!clickedHelpIcon && !clickedTooltip) {
                this.close();
            }
        };

        setTimeout(() => {
            document.addEventListener("click", this.tooltipCloseHandler);
            document.addEventListener("focusin", this.tooltipCloseHandler);
        }, 10);
    }

    destroy() {
        this.close();
        $$(".help-icon").forEach((icon) => {
            const newIcon = icon.cloneNode(true);
            icon.parentNode.replaceChild(newIcon, icon);
        });
    }
}
