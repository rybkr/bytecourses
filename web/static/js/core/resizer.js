export function createResizer(resizerElement, leftPane, rightPane, options = {}) {
    if (!resizerElement || !leftPane || !rightPane) {
        console.warn("Resizer: Missing required elements");
        return;
    }

    const {
        storageKey = "markdown-editor-split",
        minLeftWidth = 200,
        minRightWidth = 200,
        defaultRatio = 0.5,
    } = options;

    let isResizing = false;
    let startX = 0;
    let startLeftWidth = 0;

    function parseMinWidth(value) {
        if (typeof value === "number") return value;
        if (typeof value === "string") {
            if (value.endsWith("px")) {
                return parseInt(value, 10);
            }
            if (value.endsWith("%")) {
                return (parseFloat(value) / 100);
            }
        }
        return value;
    }

    const minLeft = parseMinWidth(minLeftWidth);
    const minRight = parseMinWidth(minRightWidth);

    const container = leftPane.parentElement;
    if (!container) {
        console.warn("Resizer: Could not find container");
        return;
    }

    function loadSavedRatio() {
        if (!storageKey) return null;
        try {
            const saved = localStorage.getItem(storageKey);
            if (saved !== null) {
                const ratio = parseFloat(saved);
                if (!isNaN(ratio) && ratio > 0 && ratio < 1) {
                    return ratio;
                }
            }
        } catch (e) {
            console.warn("Resizer: Failed to load saved ratio", e);
        }
        return null;
    }

    function saveRatio(ratio) {
        if (!storageKey) return;
        try {
            localStorage.setItem(storageKey, ratio.toString());
        } catch (e) {
            console.warn("Resizer: Failed to save ratio", e);
        }
    }

    function getRatio() {
        const containerWidth = container.clientWidth;
        const leftWidth = leftPane.offsetWidth;
        return leftWidth / containerWidth;
    }

    function setRatio(ratio) {
        const containerWidth = container.clientWidth;
        const resizerWidth = resizerElement.offsetWidth;
        const availableWidth = containerWidth - resizerWidth;

        let leftWidth = availableWidth * ratio;
        const rightWidth = availableWidth - leftWidth;

        const minLeftPx = typeof minLeft === "number" ? minLeft : (availableWidth * minLeft);
        const minRightPx = typeof minRight === "number" ? minRight : (availableWidth * minRight);

        if (leftWidth < minLeftPx) {
            leftWidth = minLeftPx;
        } else if (rightWidth < minRightPx) {
            leftWidth = availableWidth - minRightPx;
        }

        leftPane.style.flex = `0 0 ${leftWidth}px`;
        rightPane.style.flex = "1";
    }

    function initialize() {
        const savedRatio = loadSavedRatio();
        const ratio = savedRatio !== null ? savedRatio : defaultRatio;
        setRatio(ratio);
    }

    function startResize(e) {
        e.preventDefault();
        e.stopPropagation();
        isResizing = true;
        startX = e.clientX;
        startLeftWidth = leftPane.offsetWidth;

        resizerElement.classList.add("is-dragging");
        document.body.style.cursor = "col-resize";
        document.body.style.userSelect = "none";

        document.addEventListener("mousemove", handleResize);
        document.addEventListener("mouseup", stopResize);
    }

    function handleResize(e) {
        if (!isResizing) return;

        e.preventDefault();

        const containerWidth = container.clientWidth;
        const resizerWidth = resizerElement.offsetWidth;
        const deltaX = e.clientX - startX;
        const newLeftWidth = startLeftWidth + deltaX;

        const availableWidth = containerWidth - resizerWidth;
        let ratio = newLeftWidth / availableWidth;

        const minLeftPx = typeof minLeft === "number" ? minLeft : (availableWidth * minLeft);
        const minRightPx = typeof minRight === "number" ? minRight : (availableWidth * minRight);

        if (newLeftWidth < minLeftPx) {
            ratio = minLeftPx / availableWidth;
        } else if (availableWidth - newLeftWidth < minRightPx) {
            ratio = (availableWidth - minRightPx) / availableWidth;
        }

        setRatio(ratio);
    }

    function stopResize(e) {
        if (!isResizing) return;

        isResizing = false;

        resizerElement.classList.remove("is-dragging");
        document.body.style.cursor = "";
        document.body.style.userSelect = "";

        const ratio = getRatio();
        saveRatio(ratio);

        document.removeEventListener("mousemove", handleResize);
        document.removeEventListener("mouseup", stopResize);
    }

    function handleWindowResize() {
        if (!isResizing) {
            const ratio = getRatio();
            setRatio(ratio);
        }
    }

    initialize();

    resizerElement.addEventListener("mousedown", startResize);
    window.addEventListener("resize", handleWindowResize);

    return {
        destroy: () => {
            resizerElement.removeEventListener("mousedown", startResize);
            window.removeEventListener("resize", handleWindowResize);
            document.removeEventListener("mousemove", handleResize);
            document.removeEventListener("mouseup", stopResize);
        },
        setRatio: (ratio) => {
            setRatio(ratio);
            saveRatio(ratio);
        },
        getRatio: () => getRatio(),
    };
}
