/**
 * Render code blocks with syntax highlighting using highlight.js
 */
function highlightCodeBlocks(html) {
    if (typeof hljs === "undefined") {
        return html; // highlight.js not loaded
    }

    const tempDiv = document.createElement("div");
    tempDiv.innerHTML = html;

    // Find all code blocks
    const codeBlocks = tempDiv.querySelectorAll("pre code");
    codeBlocks.forEach((block) => {
        // Skip if already highlighted
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
            // If highlighting fails, just add hljs class
            block.className = `hljs ${language}`;
        }
    });

    return tempDiv.innerHTML;
}

/**
 * Render math expressions using KaTeX
 * Note: We use a marker class to avoid double-rendering if auto-render is active
 */
function renderMath(html) {
    if (typeof katex === "undefined") {
        return html; // KaTeX not loaded
    }

    // Use a temporary container to avoid modifying the original
    const tempDiv = document.createElement("div");
    tempDiv.innerHTML = html;

    // Render block math ($$...$$) - but skip if already rendered
    const blockMathRegex = /\$\$([\s\S]*?)\$\$/g;
    let text = tempDiv.innerHTML;
    
    // Only process if not already rendered (check for katex-display class)
    if (!text.includes('katex-display')) {
        text = text.replace(blockMathRegex, (match, formula) => {
            // Skip if this looks like it's already been processed
            if (match.includes('katex')) return match;
            try {
                return katex.renderToString(formula.trim(), {
                    displayMode: true,
                    throwOnError: false,
                });
            } catch (e) {
                return match; // Return original if rendering fails
            }
        });
    }

    // Render inline math ($...$) - but skip if already rendered
    if (!text.includes('katex')) {
        const inlineMathRegex = /(?<!\$)\$([^\$\n]+?)\$(?!\$)/g;
        text = text.replace(inlineMathRegex, (match, formula) => {
            try {
                return katex.renderToString(formula.trim(), {
                    displayMode: false,
                    throwOnError: false,
                });
            } catch (e) {
                return match; // Return original if rendering fails
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

    // Apply code highlighting
    html = highlightCodeBlocks(html);

    // Apply math rendering
    html = renderMath(html);

    // Sanitize HTML
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

    // Re-highlight code blocks after DOM update (in case highlight.js wasn't ready)
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
