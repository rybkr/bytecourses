import { createMarkdownEditor } from "./markdown-editor.js";
import { convertContent } from "./format-converter.js";

export function createUnifiedEditor(container, options = {}) {
    const {
        initialValue = "",
        initialFormat = "markdown",
        placeholder = "Write your content here...",
        onFormatChange = null,
        onUpdate = null,
    } = options;

    let currentFormat = initialFormat;
    let currentEditor = null;
    let currentContent = initialValue;

    const formatSelect = container.querySelector(".format-selector");
    const editorArea = container.querySelector(".unified-editor-area");

    if (!editorArea) {
        throw new Error("Editor area not found");
    }

    function createMarkdownEditorInstance() {
        const textarea = document.createElement("textarea");
        textarea.className = "markdown-editor-textarea";
        textarea.placeholder = placeholder;
        editorArea.innerHTML = "";
        editorArea.appendChild(textarea);

        const editor = createMarkdownEditor(textarea, {
            initialValue: currentContent,
            placeholder,
            lineNumbers: true,
            onUpdate: (content) => {
                currentContent = content;
                if (onUpdate) onUpdate(content);
            },
        });

        return {
            getValue: () => editor.getValue(),
            setValue: (value) => editor.setValue(value),
            focus: () => editor.focus(),
            destroy: () => { editorArea.innerHTML = ""; },
        };
    }

    function createPlainTextEditorInstance() {
        editorArea.innerHTML = "";

        if (typeof CodeMirror !== "undefined") {
            const wrapper = document.createElement("div");
            wrapper.className = "plain-text-editor-wrapper";
            editorArea.appendChild(wrapper);

            const cm = CodeMirror(wrapper, {
                value: currentContent,
                lineNumbers: true,
                mode: "text/plain",
                theme: "default",
                indentUnit: 4,
                lineWrapping: true,
            });

            cm.on("change", () => {
                currentContent = cm.getValue();
                if (onUpdate) onUpdate(currentContent);
            });

            setTimeout(() => cm.refresh(), 10);

            return {
                getValue: () => cm.getValue(),
                setValue: (value) => { cm.setValue(value); currentContent = value; },
                focus: () => cm.focus(),
                destroy: () => { editorArea.innerHTML = ""; },
            };
        }

        const textarea = document.createElement("textarea");
        textarea.className = "plain-text-editor-textarea";
        textarea.placeholder = placeholder;
        textarea.value = currentContent;
        editorArea.appendChild(textarea);

        textarea.addEventListener("input", () => {
            currentContent = textarea.value;
            if (onUpdate) onUpdate(currentContent);
        });

        return {
            getValue: () => textarea.value,
            setValue: (value) => { textarea.value = value; currentContent = value; },
            focus: () => textarea.focus(),
            destroy: () => { editorArea.innerHTML = ""; },
        };
    }

    function createRichTextEditorInstance() {
        if (typeof Quill === "undefined") {
            throw new Error("Quill is not loaded");
        }

        editorArea.innerHTML = "";
        const editorDiv = document.createElement("div");
        editorDiv.className = "rich-text-editor";
        editorArea.appendChild(editorDiv);

        const quill = new Quill(editorDiv, {
            theme: "snow",
            placeholder,
            modules: {
                toolbar: [
                    [{ header: [1, 2, 3, false] }],
                    ["bold", "italic", "underline", "strike"],
                    [{ list: "ordered" }, { list: "bullet" }],
                    ["link", "blockquote", "code-block"],
                    ["clean"],
                ],
            },
        });

        let sanitized = currentContent || "";
        if (typeof DOMPurify !== "undefined" && sanitized) {
            sanitized = DOMPurify.sanitize(sanitized);
        }
        quill.root.innerHTML = sanitized;

        quill.on("text-change", () => {
            let html = quill.root.innerHTML;
            if (typeof DOMPurify !== "undefined") html = DOMPurify.sanitize(html);
            currentContent = html;
            if (onUpdate) onUpdate(html);
        });

        return {
            getValue: () => {
                let html = quill.root.innerHTML;
                if (typeof DOMPurify !== "undefined") html = DOMPurify.sanitize(html);
                return html;
            },
            setValue: (value) => {
                let s = value || "";
                if (typeof DOMPurify !== "undefined" && s) s = DOMPurify.sanitize(s);
                quill.root.innerHTML = s;
                currentContent = s;
            },
            focus: () => quill.focus(),
            destroy: () => { editorArea.innerHTML = ""; },
        };
    }

    function createEditor(format) {
        switch (format) {
            case "markdown": return createMarkdownEditorInstance();
            case "plain": return createPlainTextEditorInstance();
            case "html": return createRichTextEditorInstance();
            default: throw new Error(`Unknown format: ${format}`);
        }
    }

    async function switchFormat(newFormat) {
        if (newFormat === currentFormat) return;

        const oldContent = currentEditor ? currentEditor.getValue() : currentContent;
        const convertedContent = await convertContent(oldContent, currentFormat, newFormat);

        if (currentEditor) currentEditor.destroy();

        currentFormat = newFormat;
        currentContent = convertedContent;

        if (formatSelect) formatSelect.value = newFormat;

        currentEditor = createEditor(newFormat);
        currentEditor.setValue(convertedContent);

        if (onFormatChange) onFormatChange(newFormat, convertedContent);
    }

    if (formatSelect) {
        formatSelect.value = currentFormat;
        formatSelect.addEventListener("change", (e) => switchFormat(e.target.value));
    }

    currentEditor = createEditor(currentFormat);

    return {
        getValue: () => currentEditor ? currentEditor.getValue() : currentContent,
        setValue: (value) => {
            currentContent = value;
            if (currentEditor) currentEditor.setValue(value);
        },
        getFormat: () => currentFormat,
        setFormat: (format) => switchFormat(format),
        focus: () => currentEditor?.focus(),
        destroy: () => currentEditor?.destroy(),
    };
}
