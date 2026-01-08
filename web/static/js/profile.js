document.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById("profile-form");
    const saveBtn = document.getElementById("save-btn");
    const nameInput = document.getElementById("name");
    const nameError = document.getElementById("name-error");
    const statusDiv = document.getElementById("profile-status");

    if (!form || !saveBtn || !nameInput) {
        return;
    }

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
            const response = await fetch("/api/profile", {
                method: "PATCH",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ name }),
            });

            if (!response.ok) {
                if (response.status === 401) {
                    window.location.href = "/login";
                    return;
                }
                const errorText = await response.text();
                throw new Error(errorText || "Failed to update profile");
            }

            const user = await response.json();

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

