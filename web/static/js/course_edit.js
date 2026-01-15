document.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById("course-form");
    if (!form) {
        return;
    }

    const courseId = Number(form.dataset.courseId);
    if (!Number.isFinite(courseId) || courseId <= 0) {
        console.warn("Missing or invalid course id");
        return;
    }

    const saveDelay = Number(form.dataset.autosaveDelay);

    const errorDiv = document.getElementById("error-message");
    const statusDiv = document.getElementById("save-status");
    const publishBtn = document.getElementById("publishBtn");
    const saveBtn = document.getElementById("saveBtn");

    const fieldIds = ["title", "summary"];

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
        };
    }

    async function patchCourse(payload) {
        const res = await fetch(`/api/courses/${courseId}`, {
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
            await patchCourse(payload);
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

    async function publish() {
        await saveNow();
        if (dirty) {
            return;
        }

        errorDiv.textContent = "";
        errorDiv.style.display = "none";
        publishBtn.disabled = true;

        try {
            const res = await fetch(`/api/courses/${courseId}/actions/publish`, {
                method: "POST",
            });

            if (!res.ok) {
                const txt = await res.text();
                errorDiv.textContent = txt || "Publish failed";
                errorDiv.style.display = "block";
                publishBtn.disabled = false;
                return;
            }

            window.location.href = `/courses/${courseId}`;
        } catch (error) {
            errorDiv.textContent = "Network error. Please try again.";
            errorDiv.style.display = "block";
            publishBtn.disabled = false;
        }
    }

    for (const id of fieldIds) {
        const el = document.getElementById(id);
        if (!el) {
            continue;
        }
        el.addEventListener("input", scheduleSave);
        el.addEventListener("blur", () => saveNow().catch(() => { }));
    }

    if (publishBtn) {
        publishBtn.addEventListener("click", (e) => {
            e.preventDefault();
            publish().catch(() => {
                errorDiv.textContent = "Publish failed";
                errorDiv.style.display = "block";
            });
        });
    }

    if (saveBtn) {
        saveBtn.addEventListener("click", (e) => {
            e.preventDefault();
            saveNow().catch(() => {
                errorDiv.textContent = "Save failed";
                errorDiv.style.display = "block";
            });
        });
    }

    document.addEventListener("visibilitychange", () => {
        if (document.visibilityState !== "hidden") {
            return;
        }
        if (dirty) {
            saveNow().catch(() => { });
        }
    });
});
