export default class Modal {
    constructor(modalSelector, options = {}) {
        this.modal = document.querySelector(modalSelector);
        if (!this.modal) return;

        this.options = {
            closeOnEscape: true,
            closeOnOverlayClick: true,
            ...options,
        };

        this.isOpen = false;
        this.handleKeydown = this.handleKeydown.bind(this);
        this.handleOverlayClick = this.handleOverlayClick.bind(this);

        const closeBtn = this.modal.querySelector("[data-modal-close]");
        if (closeBtn) {
            closeBtn.addEventListener("click", () => this.close());
        }

        const triggers = document.querySelectorAll(
            `[data-modal-open="${modalSelector}"]`,
        );
        triggers.forEach((trigger) => {
            trigger.addEventListener("click", (e) => {
                e.preventDefault();
                this.open();
            });
        });
    }

    handleKeydown(e) {
        if (e.key === "Escape" && this.options.closeOnEscape) {
            this.close();
        }
    }

    handleOverlayClick(e) {
        if (e.target === this.modal && this.options.closeOnOverlayClick) {
            this.close();
        }
    }

    open() {
        if (this.isOpen) return;

        this.modal.style.display = "flex";
        document.body.style.overflow = "hidden";
        this.isOpen = true;

        document.addEventListener("keydown", this.handleKeydown);
        this.modal.addEventListener("click", this.handleOverlayClick);
    }

    close() {
        if (!this.isOpen) return;

        this.modal.style.display = "none";
        document.body.style.overflow = "";
        this.isOpen = false;

        document.removeEventListener("keydown", this.handleKeydown);
        this.modal.removeEventListener("click", this.handleOverlayClick);
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
        this.modal.removeEventListener("click", this.handleOverlayClick);
    }
}
