export function createMarkdownEditor(textarea, options = {}) {
    if (typeof EasyMDE === "undefined") {
        throw new Error("EasyMDE is not loaded. Make sure easymde.min.js is included before this script.");
    }

    const {
        initialValue = "",
        placeholder = "Write your content here using Markdown...",
        lineNumbers = true,
        onUpdate = null,
    } = options;

    if (initialValue) {
        textarea.value = initialValue;
    }

    if (placeholder && !textarea.placeholder) {
        textarea.placeholder = placeholder;
    }

    const easyMDE = new EasyMDE({
        element: textarea,
        spellChecker: false,
        lineNumbers: lineNumbers,
        indentWithTabs: false,
        tabSize: 4,
        autofocus: false,
        placeholder: placeholder,
        sideBySideFullscreen: false,
        renderingConfig: {
            codeSyntaxHighlighting: true,
        },
        toolbar: [
            "bold",
            "italic",
            "strikethrough",
            "|",
            "heading-1",
            "heading-2",
            "heading-3",
            "|",
            "link",
            "image",
            "|",
            "code",
            "quote",
            "unordered-list",
            "ordered-list",
            "|",
            "horizontal-rule",
            "|",
            "preview",
            "side-by-side",
            "fullscreen",
            "|",
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

    const customPreviewElement = options.customPreviewElement || null;
    const editorContainer = options.editorContainer || null;

    if (editorContainer) {
        const setupToolbarButtons = () => {
            const easyMDEContainer = textarea.closest('.EasyMDEContainer');
            const toolbar = easyMDEContainer ? easyMDEContainer.querySelector('.editor-toolbar') : null;
            if (!toolbar) {
                setTimeout(setupToolbarButtons, 50);
                return;
            }

            const buttons = toolbar.querySelectorAll("button");
            buttons.forEach((button) => {
                const icon = button.querySelector("i");
                if (!icon) return;

                const iconClass = icon.className || "";

                if (iconClass.includes("fa-eye") || iconClass.includes("fa-eye-slash")) {
                    button.addEventListener("click", (e) => {
                        e.preventDefault();
                        e.stopImmediatePropagation();
                        
                        const isPreviewMode = editorContainer.classList.contains("mode-preview");
                        if (isPreviewMode) {
                            editorContainer.classList.remove("mode-preview");
                            editorContainer.classList.add("mode-markdown");
                            button.classList.remove("active");
                        } else {
                            editorContainer.classList.remove("mode-markdown", "mode-split");
                            editorContainer.classList.add("mode-preview");
                            button.classList.add("active");
                        }
                        return false;
                    }, true);
                }

                if (iconClass.includes("fa-columns")) {
                    button.addEventListener("click", (e) => {
                        e.preventDefault();
                        e.stopImmediatePropagation();
                        
                        const isSplitMode = editorContainer.classList.contains("mode-split");
                        if (isSplitMode) {
                            editorContainer.classList.remove("mode-split");
                            editorContainer.classList.add("mode-markdown");
                            button.classList.remove("active");
                        } else {
                            editorContainer.classList.remove("mode-markdown", "mode-preview");
                            editorContainer.classList.add("mode-split");
                            button.classList.add("active");
                        }
                        return false;
                    }, true);
                }

                if (iconClass.includes("fa-arrows-alt") || iconClass.includes("fa-compress")) {
                    button.addEventListener("click", (e) => {
                        setTimeout(() => {
                            const isFullscreen = document.querySelector(".CodeMirror-fullscreen") || 
                                               document.querySelector(".editor-preview-side") ||
                                               document.body.classList.contains("EasyMDEContainer-fullscreen");
                            if (isFullscreen && editorContainer) {
                                if (!editorContainer.classList.contains("mode-preview")) {
                                    editorContainer.classList.remove("mode-markdown");
                                    editorContainer.classList.add("mode-split");
                                }
                            }
                        }, 100);
                    });
                }
            });
        };

        setupToolbarButtons();
    }

    if (onUpdate) {
        easyMDE.codemirror.on("change", () => {
            onUpdate(easyMDE.value());
        });
    }

    return {
        editor: easyMDE,
        getValue: () => easyMDE.value(),
        setValue: (value) => easyMDE.value(value),
        focus: () => easyMDE.codemirror.focus(),
        getEditor: () => easyMDE,
    };
}

export function setupScrollSync(easyMDE, previewElement) {
    if (!easyMDE || !easyMDE.codemirror || !previewElement) {
        return;
    }

    let isScrolling = false;

    function syncEditorToPreview() {
        if (isScrolling) return;
        isScrolling = true;

        const editorScrollTop = easyMDE.codemirror.getScrollInfo().top;
        const editorScrollHeight =
            easyMDE.codemirror.getScrollInfo().height -
            easyMDE.codemirror.getScrollInfo().clientHeight;
        const scrollRatio =
            editorScrollHeight > 0 ? editorScrollTop / editorScrollHeight : 0;

        const previewScrollHeight =
            previewElement.scrollHeight - previewElement.clientHeight;
        const previewScrollTop = previewScrollHeight * scrollRatio;
        previewElement.scrollTop = previewScrollTop;

        requestAnimationFrame(() => {
            isScrolling = false;
        });
    }

    function syncPreviewToEditor() {
        if (isScrolling) return;
        isScrolling = true;

        const previewScrollTop = previewElement.scrollTop;
        const previewScrollHeight =
            previewElement.scrollHeight - previewElement.clientHeight;
        const scrollRatio =
            previewScrollHeight > 0 ? previewScrollTop / previewScrollHeight : 0;

        const editorScrollInfo = easyMDE.codemirror.getScrollInfo();
        const editorScrollTop =
            (editorScrollInfo.height - editorScrollInfo.clientHeight) * scrollRatio;
        easyMDE.codemirror.scrollTo(null, editorScrollTop);

        requestAnimationFrame(() => {
            isScrolling = false;
        });
    }

    easyMDE.codemirror.on("scroll", syncEditorToPreview);
    previewElement.addEventListener("scroll", syncPreviewToEditor);
}

export function addCustomShortcut(easyMDE, key, handler) {
    if (!easyMDE || !easyMDE.codemirror) {
        return;
    }

    const existingKeys = easyMDE.codemirror.getOption("extraKeys") || {};
    
    const normalizedKey = key.replace("Mod-", "Ctrl-");
    const newKeys = { ...existingKeys };
    
    newKeys[normalizedKey] = (cm) => {
        handler();
        return true;
    };
    
    if (key.includes("Mod-")) {
        const cmdKey = key.replace("Mod-", "Cmd-");
        newKeys[cmdKey] = (cm) => {
            handler();
            return true;
        };
    }
    
    easyMDE.codemirror.setOption("extraKeys", newKeys);
}
