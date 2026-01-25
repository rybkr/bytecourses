export default class Dropdown {
    constructor(triggerSelector, menuSelector, options = {}) {
        this.trigger = document.querySelector(triggerSelector);
        this.menu = document.querySelector(menuSelector);
        if (!this.trigger || !this.menu) return;

        this.options = {
            closeOnClickOutside: true,
            closeOnItemClick: true,
            ...options,
        };

        this.isOpen = false;
        this.handleOutsideClick = this.handleOutsideClick.bind(this);
        this.handleTriggerClick = this.handleTriggerClick.bind(this);
        this.handleItemClick = this.handleItemClick.bind(this);

        this.trigger.addEventListener("click", this.handleTriggerClick);

        if (this.options.closeOnItemClick) {
            this.menu.querySelectorAll("a, button").forEach((item) => {
                item.addEventListener("click", this.handleItemClick);
            });
        }
    }

    handleTriggerClick(e) {
        e.stopPropagation();
        e.preventDefault();
        this.toggle();
    }

    handleOutsideClick(e) {
        if (!this.menu.contains(e.target) && !this.trigger.contains(e.target)) {
            this.close();
        }
    }

    handleItemClick() {
        setTimeout(() => this.close(), 100);
    }

    open() {
        if (this.isOpen) return;

        document
            .querySelectorAll(".user-dropdown-menu, .teach-dropdown-menu")
            .forEach((m) => {
                m.classList.remove("show");
            });
        document.querySelectorAll("[aria-expanded]").forEach((el) => {
            el.setAttribute("aria-expanded", "false");
        });

        this.menu.classList.add("show");
        this.trigger.setAttribute("aria-expanded", "true");
        this.isOpen = true;

        if (this.options.closeOnClickOutside) {
            setTimeout(() => {
                document.addEventListener("click", this.handleOutsideClick);
            }, 10);
        }
    }

    close() {
        if (!this.isOpen) return;

        this.menu.classList.remove("show");
        this.trigger.setAttribute("aria-expanded", "false");
        this.isOpen = false;

        document.removeEventListener("click", this.handleOutsideClick);
    }

    toggle() {
        if (this.isOpen) {
            this.close();
        } else {
            this.open();
        }
    }

    destroy() {
        this.trigger.removeEventListener("click", this.handleTriggerClick);
        document.removeEventListener("click", this.handleOutsideClick);

        if (this.options.closeOnItemClick) {
            this.menu.querySelectorAll("a, button").forEach((item) => {
                item.removeEventListener("click", this.handleItemClick);
            });
        }
    }
}
