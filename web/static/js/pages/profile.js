import api from "../core/api.js";
import { $ } from "../core/dom.js";

document.addEventListener("DOMContentLoaded", () => {
    const form = $("#profile-form");
    const saveBtn = $("#save-btn");
    const nameInput = $("#name");
    const nameError = $("#name-error");
    const statusDiv = $("#profile-status");

    if (!form || !saveBtn || !nameInput) return;

    form.addEventListener("submit", async (e) => {
        e.preventDefault();

        nameError.style.display = "none";
        statusDiv.style.display = "none";
        statusDiv.className = "";
        statusDiv.textContent = "";

        const name = nameInput.value.trim();

        if (!name) {
            nameError.textContent = "Name is required";
            nameError.style.display = "block";
            return;
        }

        saveBtn.disabled = true;
        saveBtn.textContent = "Saving...";

        try {
            await api.patch("/api/me", { name });

            statusDiv.textContent = "Profile updated successfully";
            statusDiv.className = "success-message";
            statusDiv.style.display = "block";
            statusDiv.style.background = "var(--success-color, #4a9a9a)";
            statusDiv.style.color = "white";

            setTimeout(() => {
                window.location.reload();
            }, 1000);
        } catch (error) {
            statusDiv.textContent = error.message || "Failed to update profile. Please try again.";
            statusDiv.className = "error-message";
            statusDiv.style.display = "block";
            statusDiv.style.background = "var(--danger-color, #c85a5a)";
            statusDiv.style.color = "white";
        } finally {
            saveBtn.disabled = false;
            saveBtn.textContent = "Save Changes";
        }
    });
});
