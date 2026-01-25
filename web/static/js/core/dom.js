export function $(selector) {
    return document.querySelector(selector);
}

export function $$(selector) {
    return document.querySelectorAll(selector);
}

export function on(element, event, handler) {
    if (!element) return;
    element.addEventListener(event, handler);
}

export function delegate(container, selector, event, handler) {
    if (!container) return;
    container.addEventListener(event, (e) => {
        const target = e.target.closest(selector);
        if (target && container.contains(target)) {
            handler.call(target, e, target);
        }
    });
}
