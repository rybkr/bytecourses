/**
 * Simple toast notification component.
 * Creates auto-dismissing toast messages that appear at the bottom of the screen.
 */

const TOAST_DURATION = 3000;
const TOAST_ANIMATION_DURATION = 300;

/**
 * Shows a toast notification message.
 *
 * @param {string} message - The message to display
 * @param {string} type - The toast type: "info", "success", "warning", or "error"
 * @param {Object} options - Optional configuration
 * @param {number} options.duration - How long to show the toast (ms), default 3000
 */
export function showToast(message, type = "info", options = {}) {
    const duration = options.duration ?? TOAST_DURATION;

    // Remove any existing toast
    const existing = document.querySelector(".toast");
    if (existing) existing.remove();

    const toast = document.createElement("div");
    toast.className = `toast toast-${type}`;
    toast.textContent = message;
    toast.setAttribute("role", "alert");
    toast.setAttribute("aria-live", "polite");
    document.body.appendChild(toast);

    // Trigger animation
    requestAnimationFrame(() => {
        toast.classList.add("show");
    });

    // Auto-dismiss
    setTimeout(() => {
        toast.classList.remove("show");
        setTimeout(() => toast.remove(), TOAST_ANIMATION_DURATION);
    }, duration);

    return toast;
}

/**
 * Shows a success toast notification.
 * @param {string} message - The message to display
 * @param {Object} options - Optional configuration
 */
export function showSuccessToast(message, options = {}) {
    return showToast(message, "success", options);
}

/**
 * Shows an error toast notification.
 * @param {string} message - The message to display
 * @param {Object} options - Optional configuration
 */
export function showErrorToast(message, options = {}) {
    return showToast(message, "error", options);
}

/**
 * Shows a warning toast notification.
 * @param {string} message - The message to display
 * @param {Object} options - Optional configuration
 */
export function showWarningToast(message, options = {}) {
    return showToast(message, "warning", options);
}
