/**
 * Renders markdown content to HTML with sanitization.
 * Requires marked and DOMPurify to be loaded globally.
 *
 * @param {string} content - The markdown content to render
 * @param {Object} options - Optional configuration
 * @param {Object} options.sanitizerOptions - Options to pass to DOMPurify
 * @returns {string} The rendered and sanitized HTML
 */
export function renderMarkdown(content, options = {}) {
    if (!content) return "";

    if (typeof marked === "undefined") {
        console.warn("marked library not loaded, returning raw content");
        return content;
    }

    const raw = marked.parse(content);

    if (typeof DOMPurify !== "undefined") {
        return DOMPurify.sanitize(raw, options.sanitizerOptions);
    }

    return raw;
}

/**
 * Updates a preview element with rendered markdown content.
 * Handles showing/hiding placeholder elements.
 *
 * @param {string} content - The markdown content to render
 * @param {HTMLElement} previewEl - The container element for the preview
 * @param {Object} options - Optional configuration
 * @param {HTMLElement} options.placeholderEl - Element to show when content is empty
 * @param {HTMLElement} options.valueEl - Element to render content into (defaults to previewEl)
 * @param {string} options.wrapperClass - CSS class to wrap the rendered HTML
 */
export function updateMarkdownPreview(content, previewEl, options = {}) {
    if (!previewEl) return;

    const { placeholderEl, valueEl, wrapperClass } = options;
    const targetEl = valueEl || previewEl;
    const trimmedContent = (content || "").trim();

    if (!trimmedContent) {
        if (placeholderEl) placeholderEl.classList.remove("hidden");
        if (valueEl) {
            valueEl.innerHTML = "";
            valueEl.classList.add("hidden");
        } else {
            previewEl.innerHTML = "";
        }
        return;
    }

    if (placeholderEl) placeholderEl.classList.add("hidden");
    if (valueEl) valueEl.classList.remove("hidden");

    const html = renderMarkdown(trimmedContent);

    if (wrapperClass) {
        targetEl.innerHTML = `<div class="${wrapperClass}">${html}</div>`;
    } else {
        targetEl.innerHTML = html;
    }
}
