export function createMarkdownEditor(textarea, options = {}) {
    if (typeof EasyMDE === "undefined") {
        throw new Error("EasyMDE is not loaded");
    }

    const {
        initialValue = "",
        placeholder = "Write your content here using Markdown...",
        lineNumbers = true,
        onUpdate = null,
    } = options;

    if (initialValue) textarea.value = initialValue;
    if (placeholder && !textarea.placeholder) textarea.placeholder = placeholder;

    const easyMDE = new EasyMDE({
        element: textarea,
        spellChecker: false,
        lineNumbers,
        indentWithTabs: false,
        tabSize: 4,
        autofocus: false,
        placeholder,
        toolbar: [
            "bold", "italic", "strikethrough", "|",
            "heading-1", "heading-2", "heading-3", "|",
            "link", "image", "|",
            "code", "quote", "unordered-list", "ordered-list", "|",
            "horizontal-rule", "|",
            "preview", "side-by-side", "fullscreen", "|",
            "guide",
        ],
        shortcuts: {
            toggleBold: "Ctrl-B",
            toggleItalic: "Ctrl-I",
            drawLink: "Ctrl-K",
            togglePreview: "Ctrl-P",
            toggleSideBySide: "F9",
            toggleFullScreen: "F11",
        },
    });

    if (onUpdate) {
        easyMDE.codemirror.on("change", () => onUpdate(easyMDE.value()));
    }

    return {
        editor: easyMDE,
        getValue: () => easyMDE.value(),
        setValue: (value) => easyMDE.value(value),
        focus: () => easyMDE.codemirror.focus(),
    };
}
