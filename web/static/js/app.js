async function handleLogout() {
    try {
        const response = await fetch("/api/logout", {
            method: "POST",
        });

        if (response.ok || response.status === 204) {
            window.location.href = "/";
        }
    } catch (error) {
        console.error("Logout error:", error);
    }
}

function escapeHtml(text) {
    if (!text) return "";
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
}

function formatDate(dateString) {
    if (!dateString) return "";
    const date = new Date(dateString);
    return date.toLocaleDateString("en-US", {
        year: "numeric",
        month: "long",
        day: "numeric",
    });
}
