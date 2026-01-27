export function htmlToPlainText(html) {
    if (!html) return "";
    const tmp = document.createElement("div");
    tmp.innerHTML = html;
    return tmp.textContent || tmp.innerText || "";
}

export function plainTextToHtml(text) {
    if (!text) return "";
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML.replace(/\n/g, "<br>");
}

export function markdownToPlainText(markdown) {
    if (!markdown) return "";
    return markdown
        .replace(/^#{1,6}\s+/gm, "")
        .replace(/\*\*([^*]+)\*\*/g, "$1")
        .replace(/\*([^*]+)\*/g, "$1")
        .replace(/`([^`]+)`/g, "$1")
        .replace(/\[([^\]]+)\]\([^\)]+\)/g, "$1")
        .replace(/!\[([^\]]*)\]\([^\)]+\)/g, "$1")
        .replace(/^\s*[-*+]\s+/gm, "")
        .replace(/^\s*\d+\.\s+/gm, "")
        .trim();
}

export function plainTextToMarkdown(text) {
    if (!text) return "";
    return text;
}

export async function htmlToMarkdown(html) {
    if (!html) return "";
    if (typeof TurndownService === "undefined") {
        console.warn("TurndownService not loaded, falling back to plain text extraction");
        return htmlToPlainText(html);
    }
    const turndownService = new TurndownService();
    return turndownService.turndown(html);
}

export async function markdownToHtml(markdown) {
    if (!markdown) return "";
    if (typeof marked === "undefined") {
        console.warn("marked not loaded, returning raw markdown");
        return markdown;
    }
    return marked.parse(markdown);
}

export async function convertContent(content, fromFormat, toFormat) {
    if (!content || fromFormat === toFormat) {
        return content || "";
    }

    if (fromFormat === "markdown" && toFormat === "html") {
        return await markdownToHtml(content);
    }
    if (fromFormat === "markdown" && toFormat === "plain") {
        return markdownToPlainText(content);
    }
    if (fromFormat === "html" && toFormat === "markdown") {
        return await htmlToMarkdown(content);
    }
    if (fromFormat === "html" && toFormat === "plain") {
        return htmlToPlainText(content);
    }
    if (fromFormat === "plain" && toFormat === "markdown") {
        return plainTextToMarkdown(content);
    }
    if (fromFormat === "plain" && toFormat === "html") {
        return plainTextToHtml(content);
    }

    return content;
}
