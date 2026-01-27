import { createMarkdownEditor, setupScrollSync, addCustomShortcut } from "./markdown-editor.js";
import { convertContent } from "./format-converter.js";
import { $ } from "./dom.js";
import { updateMarkdownPreview } from "./markdown.js";

export function createUnifiedEditor(container, options = {}) {
    const {
        initialValue = "",
        initialFormat = "markdown",
        placeholder = "Write your content here...",
        onFormatChange = null,
        onUpdate = null,
        previewElement = null,
        editorContainer = null,
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
            placeholder: placeholder,
            lineNumbers: true,
            customPreviewElement: previewElement,
            editorContainer: editorContainer,
            onUpdate: (content) => {
                currentContent = content;
                if (onUpdate) onUpdate(content);
                if (previewElement) {
                    updateMarkdownPreview(content, previewElement, {
                        wrapperClass: "proposal-content-value",
                    });
                }
            },
        });

        if (previewElement && editorContainer) {
            setupScrollSync(editor.editor, previewElement);
        }

        return {
            getValue: () => editor.getValue(),
            setValue: (value) => editor.setValue(value),
            focus: () => editor.focus(),
            destroy: () => {
                editorArea.innerHTML = "";
            },
        };
    }

    function createPlainTextEditorInstance() {
        const textarea = document.createElement("textarea");
        textarea.className = "plain-text-editor-textarea";
        textarea.placeholder = placeholder;
        textarea.style.width = "100%";
        textarea.style.minHeight = "400px";
        textarea.style.fontFamily = "monospace";
        textarea.style.fontSize = "14px";
        textarea.style.padding = "12px";
        textarea.style.border = "1px solid #ddd";
        textarea.style.borderRadius = "4px";
        textarea.value = currentContent;
        editorArea.innerHTML = "";
        editorArea.appendChild(textarea);

        if (typeof CodeMirror !== "undefined") {
            const cm = CodeMirror.fromTextArea(textarea, {
                lineNumbers: true,
                mode: "text/plain",
                theme: "default",
                indentUnit: 4,
                lineWrapping: true,
            });
            cm.setValue(currentContent);
            cm.on("change", () => {
                currentContent = cm.getValue();
                if (onUpdate) onUpdate(currentContent);
            });

            return {
                getValue: () => cm.getValue(),
                setValue: (value) => cm.setValue(value),
                focus: () => cm.focus(),
                destroy: () => {
                    cm.toTextArea();
                    editorArea.innerHTML = "";
                },
            };
        }

        textarea.addEventListener("input", () => {
            currentContent = textarea.value;
            if (onUpdate) onUpdate(currentContent);
        });

        return {
            getValue: () => textarea.value,
            setValue: (value) => {
                textarea.value = value;
                currentContent = value;
            },
            focus: () => textarea.focus(),
            destroy: () => {
                editorArea.innerHTML = "";
            },
        };
    }

    function createRichTextEditorInstance() {
        if (typeof Quill === "undefined") {
            throw new Error("Quill is not loaded. Make sure quill.min.js is included.");
        }

        editorArea.innerHTML = "";
        const editorDiv = document.createElement("div");
        editorDiv.style.minHeight = "400px";
        editorArea.appendChild(editorDiv);

        const quill = new Quill(editorDiv, {
            theme: "snow",
            placeholder: placeholder,
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

        let sanitizedContent = currentContent || "";
        if (typeof DOMPurify !== "undefined" && sanitizedContent) {
            sanitizedContent = DOMPurify.sanitize(sanitizedContent);
        }
        quill.root.innerHTML = sanitizedContent;
        quill.on("text-change", () => {
            let html = quill.root.innerHTML;
            if (typeof DOMPurify !== "undefined") {
                html = DOMPurify.sanitize(html);
            }
            currentContent = html;
            if (onUpdate) onUpdate(html);
        });

        return {
            getValue: () => {
                let html = quill.root.innerHTML;
                if (typeof DOMPurify !== "undefined") {
                    html = DOMPurify.sanitize(html);
                }
                return html;
            },
            setValue: (value) => {
                let sanitizedValue = value || "";
                if (typeof DOMPurify !== "undefined" && sanitizedValue) {
                    sanitizedValue = DOMPurify.sanitize(sanitizedValue);
                }
                quill.root.innerHTML = sanitizedValue;
                currentContent = sanitizedValue;
            },
            focus: () => quill.focus(),
            destroy: () => {
                editorArea.innerHTML = "";
            },
        };
    }

    function createEditor(format) {
        switch (format) {
            case "markdown":
                return createMarkdownEditorInstance();
            case "plain":
                return createPlainTextEditorInstance();
            case "html":
                return createRichTextEditorInstance();
            default:
                throw new Error(`Unknown format: ${format}`);
        }
    }

    async function switchFormat(newFormat) {
        if (newFormat === currentFormat) return;

        const oldContent = currentEditor ? currentEditor.getValue() : currentContent;
        const convertedContent = await convertContent(oldContent, currentFormat, newFormat);

        if (currentEditor) {
            currentEditor.destroy();
        }

        currentFormat = newFormat;
        currentContent = convertedContent;

        if (formatSelect) {
            formatSelect.value = newFormat;
        }

        currentEditor = createEditor(newFormat);
        currentEditor.setValue(convertedContent);

        if (onFormatChange) {
            onFormatChange(newFormat, convertedContent);
        }
    }

    if (formatSelect) {
        formatSelect.value = currentFormat;
        formatSelect.addEventListener("change", async (e) => {
            const newFormat = e.target.value;
            await switchFormat(newFormat);
        });
    }

    currentEditor = createEditor(currentFormat);

    return {
        getValue: () => (currentEditor ? currentEditor.getValue() : currentContent),
        setValue: (value) => {
            currentContent = value;
            if (currentEditor) {
                currentEditor.setValue(value);
            }
        },
        getFormat: () => currentFormat,
        setFormat: async (format) => {
            await switchFormat(format);
        },
        focus: () => {
            if (currentEditor) {
                currentEditor.focus();
            }
        },
        destroy: () => {
            if (currentEditor) {
                currentEditor.destroy();
            }
        },
    };
}
