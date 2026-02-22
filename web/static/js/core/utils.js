import ConfirmationModal from "../components/ConfirmationModal.js";
import api from "./api.js";

export { extractErrorMessage } from "./errors.js";

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

export async function confirmAction(message, options = {}) {
    const modal = new ConfirmationModal({
        title: options.title || "Confirm Action",
        message: message,
        confirmText: options.confirmText || "Confirm",
        cancelText: options.cancelText || "Cancel",
        confirmButtonClass: options.confirmButtonClass || "btn-primary",
        variant: options.variant || "info",
    });

    return await modal.show();
}

export function nowLabel() {
    const d = new Date();
    return d.toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
    });
}

export async function deleteProposal(proposalId, status, options = {}) {
    const isSubmitted = status === "submitted";

    let confirmMessage;
    if (isSubmitted) {
        confirmMessage =
            "This proposal will be withdrawn from review and then permanently deleted. This action cannot be undone.";
    } else {
        confirmMessage =
            "This action cannot be undone. The proposal and all its data will be permanently deleted.";
    }

    const confirmed = await confirmAction(confirmMessage, {
        title: "Delete Proposal?",
        confirmText: "Delete",
        confirmButtonClass: "btn-danger",
        variant: "danger",
    });

    if (!confirmed) {
        return false;
    }

    try {
        if (isSubmitted) {
            await api.post(`/api/proposals/${proposalId}/actions/withdraw`);
        }
        await api.delete(`/api/proposals/${proposalId}`);
        return true;
    } catch (error) {
        if (options.onError) {
            options.onError(error);
        } else {
            throw error;
        }
        return false;
    }
}

export const SANITIZE_CONFIG = {
    ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 'u', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'ul', 'ol', 'li', 'blockquote', 'code', 'pre', 'a', 'img', 'iframe'],
    ALLOWED_ATTR: ['href', 'src', 'alt', 'title', 'width', 'height', 'frameborder', 'allow', 'allowfullscreen', 'loading', 'class'],
    ALLOWED_URI_REGEXP: /^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i,
};

export function extractVideoEmbedCode(url) {
    if (!url || typeof url !== "string") {
        return null;
    }

    const trimmedUrl = url.trim();
    let videoId = null;
    let embedUrl = null;

    if (trimmedUrl.includes("youtube.com/watch") || trimmedUrl.includes("youtu.be/")) {
        const youtubeRegex = /(?:youtube\.com\/watch\?(?:.*&)?v=|youtu\.be\/)([a-zA-Z0-9_-]{11})/;
        const match = trimmedUrl.match(youtubeRegex);
        if (match && match[1]) {
            videoId = match[1];
            embedUrl = `https://www.youtube.com/embed/${videoId}`;
        }
    } else if (trimmedUrl.includes("vimeo.com/")) {
        const vimeoRegex = /vimeo\.com\/(\d+)/;
        const match = trimmedUrl.match(vimeoRegex);
        if (match && match[1]) {
            videoId = match[1];
            embedUrl = `https://player.vimeo.com/video/${videoId}`;
        }
    }

    if (!embedUrl) {
        return null;
    }

    return `<iframe src="${embedUrl}" width="560" height="315" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>`;
}
