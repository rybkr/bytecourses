import api from "../core/api.js";
import { $ } from "../core/dom.js";
import { showError, hideError } from "../core/utils.js";

export default class FormSubmitHandler {
    constructor(formSelector, options = {}) {
        this.form = $(formSelector);
        if (!this.form) return;

        this.options = {
            endpoint: options.endpoint,
            method: options.method || "POST",
            successRedirect: options.successRedirect,
            errorContainer: options.errorContainer || "#error-message",
            submitButton: options.submitButton,
            onSuccess: options.onSuccess,
            onError: options.onError,
            transformData: options.transformData || ((data) => data),
            ...options,
        };

        this.errorContainer = $(this.options.errorContainer);
        this.submitButton = this.options.submitButton
            ? $(this.options.submitButton)
            : null;

        this.init();
    }

    init() {
        if (!this.form) return;

        this.form.addEventListener("submit", async (e) => {
            e.preventDefault();
            await this.handleSubmit();
        });
    }

    getFormData() {
        const formData = new FormData(this.form);
        const data = {};
        for (const [key, value] of formData.entries()) {
            data[key] = value;
        }
        return this.options.transformData(data);
    }

    async handleSubmit() {
        if (!this.options.endpoint) {
            console.error("No endpoint specified");
            return;
        }

        hideError(this.errorContainer);

        if (this.submitButton) {
            this.submitButton.disabled = true;
        }

        try {
            const data = this.getFormData();
            const response = await api.post(this.options.endpoint, data);

            if (response && response.ok) {
                if (this.options.onSuccess) {
                    this.options.onSuccess(response);
                } else if (this.options.successRedirect) {
                    window.location.href = this.options.successRedirect;
                } else {
                    window.location.reload();
                }
            }
        } catch (error) {
            const errorMessage =
                error.message || "An error occurred. Please try again.";
            showError(errorMessage, this.errorContainer);

            if (this.options.onError) {
                this.options.onError(error);
            }
        } finally {
            if (this.submitButton) {
                this.submitButton.disabled = false;
            }
        }
    }

    destroy() {
        if (this.form) {
            this.form.removeEventListener("submit", this.handleSubmit);
        }
    }
}
