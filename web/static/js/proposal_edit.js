document.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById("proposal-form");
    if (!form) {
        return;
    }

    const proposalId = Number(form.dataset.proposalId);
    if (!Number.isFinite(proposalId) || proposalId <= 0) {
        console.warn("Missing or invalid proposal id");
        return;
    }

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
            saveNow().catch(() => {});
        }, saveDelay);
    }

    async function saveNow() {
        clearTimeout(saveTimer);
        if (!dirty || saveInFlight) {
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
            await patchProposal(payload);
            lastSavedJson = json;
            dirty = false;
            statusDiv.textContent = `Saved at ${nowLabel()}`;
        } catch (e) {
            errorDiv.textContent = e.message || "Autosave failed";
            errorDiv.style.display = "block";
        } finally {
            saveInFlight = false;
        }
    }

    async function submit() {
        await saveNow();
        if (dirty) {
            return;
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
        await saveNow();
        if (!dirty) {
            window.location.href = `/proposals/${proposalId}`;
        }
    }

    for (const id of fieldIds) {
        const el = document.getElementById(id);
        if (!el) {
            continue;
        }
        el.addEventListener("input", scheduleSave);
        el.addEventListener("blur", () => saveNow().catch(() => {}));
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

        try {
            const payload = readPayload();
            navigator.sendBeacon(
                `/api/proposals/${proposalId}`,
                JSON.stringify(payload),
            );
        } catch (_) {}
    });

    lastSavedJson = JSON.stringify(readPayload());
});
