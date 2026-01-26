const TOAST_DURATION = 3000;
const TOAST_ANIMATION_DURATION = 300;

export function showToast(message, type = "info", options = {}) {
    const duration = options.duration ?? TOAST_DURATION;

    const existing = document.querySelector(".toast");
    if (existing) existing.remove();

    const toast = document.createElement("div");
    toast.className = `toast toast-${type}`;
    toast.textContent = message;
    toast.setAttribute("role", "alert");
    toast.setAttribute("aria-live", "polite");
    document.body.appendChild(toast);

    requestAnimationFrame(() => {
        toast.classList.add("show");
    });

    setTimeout(() => {
        toast.classList.remove("show");
        setTimeout(() => toast.remove(), TOAST_ANIMATION_DURATION);
    }, duration);

    return toast;
}

export function showSuccessToast(message, options = {}) {
    return showToast(message, "success", options);
}

export function showErrorToast(message, options = {}) {
    return showToast(message, "error", options);
}

export function showWarningToast(message, options = {}) {
    return showToast(message, "warning", options);
}
