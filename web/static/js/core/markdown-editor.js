export function createMarkdownEditor(textarea, options = {}) {
    if (typeof EasyMDE === "undefined") {
        throw new Error("EasyMDE is not loaded");
    }

    const {
        initialValue = "",
        placeholder = "Write your content here using Markdown...",
        lineNumbers = true,
        onUpdate = null,
        extractVideoEmbedCode = null,
    } = options;

    if (initialValue) textarea.value = initialValue;
    if (placeholder && !textarea.placeholder) textarea.placeholder = placeholder;

    let videoButtonHandler = null;
    if (typeof extractVideoEmbedCode === "function") {
        videoButtonHandler = {
            name: "video",
            action: function (editor) {
                const url = prompt("Enter video URL (YouTube or Vimeo):");
                if (!url) return;

                const embedCode = extractVideoEmbedCode(url);
                if (!embedCode) {
                    alert("Invalid video URL. Please enter a valid YouTube or Vimeo URL.");
                    return;
                }

                const cm = editor.codemirror;
                const embedMarkdown = "\n\n" + embedCode + "\n\n";
                cm.replaceSelection(embedMarkdown);
                cm.focus();
            },
            className: "fa fa-video",
            title: "Insert Video",
        };
    }

    const toolbar = [
        "bold", "italic", "strikethrough", "|",
        "heading-1", "heading-2", "heading-3", "|",
        "link", "image",
    ];

    if (videoButtonHandler) {
        toolbar.push("|", videoButtonHandler);
    }

    toolbar.push(
        "|",
        "code", "quote", "unordered-list", "ordered-list", "|",
        "horizontal-rule", "|",
        "preview", "side-by-side", "fullscreen", "|",
        "guide"
    );

    const easyMDE = new EasyMDE({
        element: textarea,
        spellChecker: false,
        lineNumbers,
        indentWithTabs: false,
        tabSize: 4,
        autofocus: false,
        placeholder,
        toolbar: toolbar,
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
