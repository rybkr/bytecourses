import Modal from "./Modal.js";
import { escapeHtml } from "../core/utils.js";

export default class ConfirmationModal {
    constructor(options = {}) {
        this.options = {
            title: "Confirm Action",
            message: "Are you sure you want to proceed?",
            confirmText: "Confirm",
            cancelText: "Cancel",
            confirmButtonClass: "btn-primary",
            variant: "info", // info, warning, danger
            ...options,
        };

        this.modal = null;
        this.resolve = null;
        this.createModal();
    }

    createModal() {
        const modalId = `confirmation-modal-${Date.now()}`;
        const modalHtml = `
            <div id="${modalId}" class="modal confirmation-modal" style="display: none;">
                <div class="modal-overlay"></div>
                <div class="modal-content confirmation-modal-content confirmation-modal-${this.options.variant}">
                    <div class="modal-header">
                        <h2>${escapeHtml(this.options.title)}</h2>
                        <button class="modal-close" type="button" data-modal-close aria-label="Close">&times;</button>
                    </div>
                    <div class="modal-body">
                        <p>${escapeHtml(this.options.message)}</p>
                        <div class="modal-actions">
                            <button class="btn ${this.options.confirmButtonClass}" data-confirm>${escapeHtml(this.options.confirmText)}</button>
                            <button type="button" class="btn btn-outline" data-cancel>${escapeHtml(this.options.cancelText)}</button>
                        </div>
                    </div>
                </div>
            </div>
        `;

        document.body.insertAdjacentHTML("beforeend", modalHtml);
        this.modal = document.getElementById(modalId);
        this.setupModal();
    }

    setupModal() {
        this.modalInstance = new Modal(`#${this.modal.id}`, {
            closeOnEscape: true,
            closeOnOverlayClick: false, // Don't close on overlay click for confirmations
        });

        const confirmBtn = this.modal.querySelector("[data-confirm]");
        const cancelBtn = this.modal.querySelector("[data-cancel]");
        const closeBtn = this.modal.querySelector("[data-modal-close]");

        const close = (result) => {
            this.modalInstance.close();
            if (this.resolve) {
                this.resolve(result);
                this.resolve = null;
            }
            // Clean up after animation
            setTimeout(() => {
                if (this.modal && this.modal.parentNode) {
                    this.modal.parentNode.removeChild(this.modal);
                }
            }, 300);
        };

        if (confirmBtn) {
            confirmBtn.addEventListener("click", () => close(true));
        }

        if (cancelBtn) {
            cancelBtn.addEventListener("click", () => close(false));
        }

        if (closeBtn) {
            closeBtn.addEventListener("click", () => close(false));
        }

        // Handle Escape key
        const handleEscape = (e) => {
            if (e.key === "Escape" && this.modalInstance.isOpen) {
                close(false);
            }
        };
        this.modal.addEventListener("keydown", handleEscape);
    }

    show() {
        return new Promise((resolve) => {
            this.resolve = resolve;
            this.modalInstance.open();
            // Focus confirm button for accessibility
            const confirmBtn = this.modal.querySelector("[data-confirm]");
            if (confirmBtn) {
                setTimeout(() => confirmBtn.focus(), 100);
            }
        });
    }

}
