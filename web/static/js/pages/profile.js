import api from "../core/api.js";
import { $ } from "../core/dom.js";
import { showError, hideError } from "../core/utils.js";

document.addEventListener("DOMContentLoaded", () => {
    const form = $("#profile-form");
    const saveBtn = $("#save-btn");
    const nameInput = $("#name");
    const nameError = $("#name-error");
    const statusDiv = $("#profile-status");

    if (!form || !saveBtn || !nameInput) return;

    form.addEventListener("submit", async (e) => {
        e.preventDefault();

        hideError(nameError);
        if (statusDiv) {
            statusDiv.classList.add("hidden");
            statusDiv.className = statusDiv.className
                .replace(/\b(success-message|error-message)\b/g, "")
                .trim();
            statusDiv.textContent = "";
        }

        const name = nameInput.value.trim();

        if (!name) {
            showError("Name is required", nameError);
            return;
        }

        saveBtn.disabled = true;
        saveBtn.textContent = "Saving...";

        try {
            await api.patch("/api/me", { name });

            if (statusDiv) {
                statusDiv.textContent = "Profile updated successfully";
                statusDiv.className = "success-message";
                statusDiv.classList.remove("hidden");
            }

            setTimeout(() => {
                window.location.reload();
            }, 1000);
        } catch (error) {
            if (statusDiv) {
                showError(
                    error.message ||
                        "Failed to update profile. Please try again.",
                    statusDiv,
                );
            } else {
                showError(
                    error.message ||
                        "Failed to update profile. Please try again.",
                    nameError,
                );
            }
        } finally {
            saveBtn.disabled = false;
            saveBtn.textContent = "Save Changes";
        }
    });
});
