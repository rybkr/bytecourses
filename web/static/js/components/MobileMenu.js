export default class MobileMenu {
    constructor(menuSelector, overlaySelector, buttonSelector) {
        this.menu = document.querySelector(menuSelector);
        this.overlay = document.querySelector(overlaySelector);
        this.button = document.querySelector(buttonSelector);
        if (!this.menu || !this.overlay || !this.button) return;

        this.isOpen = false;
        this.handleKeydown = this.handleKeydown.bind(this);

        this.button.addEventListener("click", (e) => {
            e.stopPropagation();
            this.toggle();
        });

        this.overlay.addEventListener("click", () => this.close());

        this.menu.querySelectorAll("a, button").forEach((item) => {
            item.addEventListener("click", () => this.close());
        });
    }

    handleKeydown(e) {
        if (e.key === "Escape") {
            this.close();
        }
    }

    open() {
        if (this.isOpen) return;

        document.querySelectorAll(".user-dropdown-menu, .teach-dropdown-menu").forEach((m) => {
            m.classList.remove("show");
        });
        document.querySelectorAll("[aria-expanded]").forEach((el) => {
            el.setAttribute("aria-expanded", "false");
        });

        this.menu.classList.add("active");
        this.overlay.classList.add("active");
        this.button.classList.add("active");
        this.button.setAttribute("aria-expanded", "true");
        document.body.style.overflow = "hidden";
        this.isOpen = true;

        document.addEventListener("keydown", this.handleKeydown);
    }

    close() {
        if (!this.isOpen) return;

        this.menu.classList.remove("active");
        this.overlay.classList.remove("active");
        this.button.classList.remove("active");
        this.button.setAttribute("aria-expanded", "false");
        document.body.style.overflow = "";
        this.isOpen = false;

        document.removeEventListener("keydown", this.handleKeydown);
    }

    toggle() {
        if (this.isOpen) {
            this.close();
        } else {
            this.open();
        }
    }

    destroy() {
        this.close();
        document.removeEventListener("keydown", this.handleKeydown);
    }
}
