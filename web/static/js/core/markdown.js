function highlightCodeBlocks(html) {
    if (typeof hljs === "undefined") {
        return html;
    }

    const tempDiv = document.createElement("div");
    tempDiv.innerHTML = html;

    const codeBlocks = tempDiv.querySelectorAll("pre code");
    codeBlocks.forEach((block) => {
        if (block.classList.length > 0) return;

        const language = block.className || "plaintext";
        const code = block.textContent;

        try {
            const highlighted = hljs.highlight(code, {
                language: language.replace("language-", "") || "plaintext",
            });
            block.innerHTML = highlighted.value;
            block.className = `hljs ${language}`;
        } catch (e) {
            block.className = `hljs ${language}`;
        }
    });

    return tempDiv.innerHTML;
}

function renderMath(html) {
    if (typeof katex === "undefined") {
        return html;
    }

    const tempDiv = document.createElement("div");
    tempDiv.innerHTML = html;

    const blockMathRegex = /\$\$([\s\S]*?)\$\$/g;
    let text = tempDiv.innerHTML;
    
    if (!text.includes('katex-display')) {
        text = text.replace(blockMathRegex, (match, formula) => {
            if (match.includes('katex')) return match;
            try {
                return katex.renderToString(formula.trim(), {
                    displayMode: true,
                    throwOnError: false,
                });
            } catch (e) {
                return match;
            }
        });
    }

    if (!text.includes('katex')) {
        const inlineMathRegex = /(?<!\$)\$([^\$\n]+?)\$(?!\$)/g;
        text = text.replace(inlineMathRegex, (match, formula) => {
            try {
                return katex.renderToString(formula.trim(), {
                    displayMode: false,
                    throwOnError: false,
                });
            } catch (e) {
                return match;
            }
        });
    }

    tempDiv.innerHTML = text;
    return tempDiv.innerHTML;
}

export function renderMarkdown(content, options = {}) {
    if (!content) return "";

    if (typeof marked === "undefined") {
        console.warn("marked library not loaded, returning raw content");
        return content;
    }

    let html = marked.parse(content);

    html = highlightCodeBlocks(html);
    html = renderMath(html);

    if (typeof DOMPurify !== "undefined") {
        html = DOMPurify.sanitize(html, options.sanitizerOptions);
    }

    return html;
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

    if (typeof hljs !== "undefined") {
        const codeBlocks = targetEl.querySelectorAll("pre code:not(.hljs)");
        codeBlocks.forEach((block) => {
            const language = block.className || "plaintext";
            const code = block.textContent;
            try {
                const highlighted = hljs.highlight(code, {
                    language: language.replace("language-", "") || "plaintext",
                });
                block.innerHTML = highlighted.value;
                block.className = `hljs ${language}`;
            } catch (e) {
                block.className = `hljs ${language}`;
            }
        });
    }
}
