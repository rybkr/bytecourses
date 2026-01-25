export function escapeHtml(text) {
    if (!text) return "";
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
}

export function formatDate(dateString) {
    if (!dateString) return "";
    const date = new Date(dateString);
    return date.toLocaleDateString("en-US", {
        year: "numeric",
        month: "long",
        day: "numeric",
    });
}

export function debounce(fn, delay) {
    let timer = null;
    return function (...args) {
        clearTimeout(timer);
        timer = setTimeout(() => fn.apply(this, args), delay);
    };
}

export function showError(message, container) {
    if (!container) return;
    container.textContent = message;
    container.classList.remove("hidden");
}

export function hideError(container) {
    if (!container) return;
    container.textContent = "";
    container.classList.add("hidden");
}

export function confirmAction(message) {
    return confirm(message);
}

export function nowLabel() {
    const d = new Date();
    return d.toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
    });
}
