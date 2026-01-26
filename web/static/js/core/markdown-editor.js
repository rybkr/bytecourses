/**
 * Markdown Editor Module using EasyMDE
 * Provides a reusable markdown editor with syntax highlighting, toolbar, and keyboard shortcuts
 */

/**
 * Create a markdown editor instance using EasyMDE
 * @param {HTMLTextAreaElement} textarea - Textarea element for the editor
 * @param {Object} options - Configuration options
 * @returns {{ editor: EasyMDE, getValue: Function, setValue: Function, focus: Function, getEditor: Function }}
 */
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

    // Set initial value if provided
    if (initialValue) {
        textarea.value = initialValue;
    }

    // Set placeholder
    if (placeholder && !textarea.placeholder) {
        textarea.placeholder = placeholder;
    }

    // Configure EasyMDE
    const easyMDE = new EasyMDE({
        element: textarea,
        spellChecker: false,
        lineNumbers: lineNumbers,
        indentWithTabs: false,
        tabSize: 4,
        autofocus: false,
        placeholder: placeholder,
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

    // Setup update listener if callback provided
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

/**
 * Setup scroll synchronization between editor and preview
 * @param {EasyMDE} easyMDE - EasyMDE editor instance
 * @param {HTMLElement} previewElement - Preview container element
 */
export function setupScrollSync(easyMDE, previewElement) {
    if (!easyMDE || !easyMDE.codemirror || !previewElement) {
        return;
    }

    let isScrolling = false;

    // Sync editor scroll to preview
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

    // Sync preview scroll to editor
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

    // Attach scroll listeners
    easyMDE.codemirror.on("scroll", syncEditorToPreview);
    previewElement.addEventListener("scroll", syncPreviewToEditor);
}

/**
 * Add custom keyboard shortcut to EasyMDE editor
 * @param {EasyMDE} easyMDE - EasyMDE editor instance
 * @param {string} key - Key combination (e.g., "Ctrl-Enter", "Mod-s")
 * @param {Function} handler - Handler function
 */
export function addCustomShortcut(easyMDE, key, handler) {
    if (!easyMDE || !easyMDE.codemirror) {
        return;
    }

    // Get existing extraKeys or empty object
    const existingKeys = easyMDE.codemirror.getOption("extraKeys") || {};
    
    // Map Mod to Ctrl/Cmd - CodeMirror 5 uses "Ctrl" for both Ctrl and Cmd
    // We need to handle both Ctrl and Cmd separately for cross-platform support
    const normalizedKey = key.replace("Mod-", "Ctrl-");
    
    // Create new extraKeys object
    const newKeys = { ...existingKeys };
    
    // Add the shortcut
    newKeys[normalizedKey] = (cm) => {
        handler();
        return true; // Prevent default behavior
    };
    
    // Also add Cmd version for Mac
    if (key.includes("Mod-")) {
        const cmdKey = key.replace("Mod-", "Cmd-");
        newKeys[cmdKey] = (cm) => {
            handler();
            return true;
        };
    }
    
    easyMDE.codemirror.setOption("extraKeys", newKeys);
}
