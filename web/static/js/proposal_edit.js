document.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById("proposal-form");
    if (!form) {
        return;
    }

    let proposalId = Number(form.dataset.proposalId);
    const isNewProposal = !Number.isFinite(proposalId) || proposalId <= 0;

    const saveDelay = Number(form.dataset.autosaveDelay);

    const errorDiv = document.getElementById("error-message");
    const statusDiv = document.getElementById("save-status");
    const submitBtn = document.getElementById("submitBtn");
    const saveDraftBtn = document.getElementById("saveDraftBtn");

    const fieldIds = [
        "title",
        "summary",
        "qualifications",
        "target_audience",
        "learning_objectives",
        "outline",
        "assumed_prerequisites",
    ];

    let saveTimer = null;
    let dirty = false;
    let saveInFlight = false;
    let lastSavedJson = null;
    let createInFlight = false;

    function nowLabel() {
        const d = new Date();
        return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" });
    }

    function readPayload() {
        return {
            title: document.getElementById("title")?.value ?? "",
            summary: document.getElementById("summary")?.value ?? "",
            qualifications: document.getElementById("qualifications")?.value ?? "",
            target_audience:
                document.getElementById("target_audience")?.value ?? "",
            learning_objectives:
                document.getElementById("learning_objectives")?.value ?? "",
            outline: document.getElementById("outline")?.value ?? "",
            assumed_prerequisites:
                document.getElementById("assumed_prerequisites")?.value ?? "",
        };
    }

    async function createProposal(payload) {
        const res = await fetch(`/api/proposals`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload),
        });

        if (!res.ok) {
            const txt = await res.text();
            throw new Error(txt || "Create failed");
        }

        const proposal = await res.json();
        return proposal.id;
    }

    async function patchProposal(payload) {
        const res = await fetch(`/api/proposals/${proposalId}`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload),
        });

        if (!res.ok) {
            const txt = await res.text();
            throw new Error(txt || "Save failed");
        }
    }

    function scheduleSave() {
        dirty = true;
        clearTimeout(saveTimer);
        saveTimer = setTimeout(() => {
            saveNow().catch(() => { });
        }, saveDelay);
    }

    async function saveNow() {
        clearTimeout(saveTimer);
        if (!dirty || saveInFlight || createInFlight) {
            return;
        }

        const payload = readPayload();
        const json = JSON.stringify(payload);
        if (json === lastSavedJson) {
            dirty = false;
            statusDiv.textContent = `Saved at ${nowLabel()}`;
            return;
        }

        saveInFlight = true;
        errorDiv.textContent = "";
        errorDiv.style.display = "none";

        try {
            // If this is a new proposal, create it first
            if (isNewProposal || proposalId <= 0) {
                createInFlight = true;
                const newId = await createProposal(payload);
                proposalId = newId;
                form.dataset.proposalId = proposalId.toString();
                createInFlight = false;
            } else {
                await patchProposal(payload);
            }
            lastSavedJson = json;
            dirty = false;
            statusDiv.textContent = `Saved at ${nowLabel()}`;
        } catch (e) {
            errorDiv.textContent = e.message || "Autosave failed";
            errorDiv.style.display = "block";
            createInFlight = false;
        } finally {
            saveInFlight = false;
        }
    }

    async function submit() {
        // Ensure proposal is created/saved first
        await saveNow();
        if (dirty || saveInFlight || createInFlight) {
            await new Promise((resolve) => {
                const checkInterval = setInterval(() => {
                    if (!dirty && !saveInFlight && !createInFlight) {
                        clearInterval(checkInterval);
                        resolve();
                    }
                }, 100);
                setTimeout(() => {
                    clearInterval(checkInterval);
                    resolve();
                }, 1000);
            });
        }

        // If still no proposal ID, create it now
        if (isNewProposal || proposalId <= 0) {
            const payload = readPayload();
            try {
                const newId = await createProposal(payload);
                proposalId = newId;
                form.dataset.proposalId = proposalId.toString();
            } catch (e) {
                errorDiv.textContent = e.message || "Failed to create proposal";
                errorDiv.style.display = "block";
                return;
            }
        }

        errorDiv.textContent = "";
        errorDiv.style.display = "none";

        const res = await fetch(`/api/proposals/${proposalId}/actions/submit`, {
            method: "POST",
        });

        if (!res.ok) {
            const txt = await res.text();
            errorDiv.textContent = txt || "Submit failed";
            errorDiv.style.display = "block";
            return;
        }

        window.location.href = `/proposals/${proposalId}`;
    }

    async function saveDraftAndExit() {
        // Ensure proposal is created/saved first
        await saveNow();
        if (dirty || saveInFlight || createInFlight) {
            await new Promise((resolve) => {
                const checkInterval = setInterval(() => {
                    if (!dirty && !saveInFlight && !createInFlight) {
                        clearInterval(checkInterval);
                        resolve();
                    }
                }, 100);
                setTimeout(() => {
                    clearInterval(checkInterval);
                    resolve();
                }, 1000);
            });
        }

        // If still no proposal ID, create it now
        if (isNewProposal || proposalId <= 0) {
            const payload = readPayload();
            try {
                const newId = await createProposal(payload);
                proposalId = newId;
                form.dataset.proposalId = proposalId.toString();
            } catch (e) {
                errorDiv.textContent = e.message || "Failed to create proposal";
                errorDiv.style.display = "block";
                return;
            }
        }

        window.location.href = `/proposals/${proposalId}`;
    }

    for (const id of fieldIds) {
        const el = document.getElementById(id);
        if (!el) {
            continue;
        }
        el.addEventListener("input", scheduleSave);
        el.addEventListener("blur", () => saveNow().catch(() => { }));
    }

    submitBtn.addEventListener("click", (e) => {
        e.preventDefault();
        submit().catch(() => {
            errorDiv.textContent = "Submit failed";
            errorDiv.style.display = "block";
        });
    });

    saveDraftBtn.addEventListener("click", (e) => {
        e.preventDefault();
        saveDraftAndExit().catch(() => {
            errorDiv.textContent = "Save failed";
            errorDiv.style.display = "block";
        });
    });

    document.addEventListener("visibilitychange", () => {
        if (document.visibilityState !== "hidden") {
            return;
        }
        if (!dirty) {
            return;
        }
        // Don't try to save new proposals via sendBeacon (can't create via beacon)
        if (isNewProposal || proposalId <= 0) {
            return;
        }

        try {
            const payload = readPayload();
            navigator.sendBeacon(
                `/api/proposals/${proposalId}`,
                JSON.stringify(payload),
            );
        } catch (_) { }
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
        tooltipRepositionHandlers.forEach(handler => {
            window.removeEventListener("scroll", handler, true);
            window.removeEventListener("resize", handler);
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
        let preferAbove = false;

        if (top + tooltipRect.height > viewportHeight - padding) {
            const topPosition = iconRect.top - tooltipRect.height - 8;
            if (topPosition >= padding) {
                top = topPosition;
                preferAbove = true;
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
        const helpIcon = document.querySelector(`.help-icon[data-help="${fieldName}"]`);
        const helpPanel = document.getElementById(`help-${fieldName}`);

        if (!helpIcon || !helpPanel) {
            return;
        }

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
            fieldName: fieldName
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

    const helpIcons = document.querySelectorAll(".help-icon");
    helpIcons.forEach((icon) => {
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

    lastSavedJson = JSON.stringify(readPayload());
});
