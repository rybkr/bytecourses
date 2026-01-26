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
