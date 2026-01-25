import api from "../core/api.js";
import { nowLabel, hideError, showError } from "../core/utils.js";

export default class FormHandler {
    constructor(formSelector, options) {
        this.form = document.querySelector(formSelector);
        if (!this.form) return;

        this.options = {
            autosaveDelay: 2000,
            ...options,
        };

        if (!this.options.apiPath || !this.options.fieldIds) {
            return;
        }

        this.entityId = this.options.entityId ?? null;
        this.errorContainer = this.options.errorContainer
            ? document.querySelector(this.options.errorContainer)
            : null;
        this.statusContainer = this.options.statusContainer
            ? document.querySelector(this.options.statusContainer)
            : null;

        this.saveTimer = null;
        this.dirty = false;
        this.saveInFlight = false;
        this.createInFlight = false;
        this.lastSavedJson = null;

        this.scheduleSave = this.scheduleSave.bind(this);
        this.saveNow = this.saveNow.bind(this);

        this.options.fieldIds.forEach((id) => {
            const el = document.getElementById(id);
            if (!el) return;
            el.addEventListener("input", this.scheduleSave);
            el.addEventListener("blur", () => this.saveNow().catch(() => {}));
        });

        this.lastSavedJson = JSON.stringify(this.readPayload());
    }

    readPayload() {
        const payload = {};
        this.options.fieldIds.forEach((id) => {
            const el = document.getElementById(id);
            payload[id] = el?.value ?? "";
        });
        return payload;
    }

    scheduleSave() {
        this.markDirty();
        clearTimeout(this.saveTimer);
        this.saveTimer = setTimeout(() => {
            this.saveNow().catch(() => {});
        }, this.options.autosaveDelay);
    }

    markDirty() {
        this.dirty = true;
    }

    markClean() {
        this.dirty = false;
    }

    async saveNow() {
        clearTimeout(this.saveTimer);
        if (!this.dirty || this.saveInFlight || this.createInFlight) {
            return;
        }

        const payload = this.readPayload();
        const json = JSON.stringify(payload);
        if (json === this.lastSavedJson) {
            this.dirty = false;
            this.updateStatus(`Saved at ${nowLabel()}`);
            return;
        }

        this.saveInFlight = true;
        this.clearError();

        try {
            if (this.entityId === null) {
                this.createInFlight = true;
                const response = await api.post(this.options.apiPath, payload);
                if (response) {
                    const data = await response.json();
                    this.entityId = data.id;
                    if (this.options.onEntityCreated) {
                        this.options.onEntityCreated(this.entityId);
                    }
                }
                this.createInFlight = false;
            } else {
                await api.patch(
                    `${this.options.apiPath}/${this.entityId}`,
                    payload,
                );
            }
            this.lastSavedJson = json;
            this.dirty = false;
            this.updateStatus(`Saved at ${nowLabel()}`);
            if (this.options.onSave) {
                this.options.onSave(payload);
            }
        } catch (e) {
            this.showError(e.message || "Autosave failed");
            this.createInFlight = false;
            if (this.options.onError) {
                this.options.onError(e);
            }
        } finally {
            this.saveInFlight = false;
        }
    }

    async waitForSave() {
        await this.saveNow();
        if (this.dirty || this.saveInFlight || this.createInFlight) {
            await new Promise((resolve) => {
                const checkInterval = setInterval(() => {
                    if (
                        !this.dirty &&
                        !this.saveInFlight &&
                        !this.createInFlight
                    ) {
                        clearInterval(checkInterval);
                        resolve();
                    }
                }, 100);
                setTimeout(() => {
                    clearInterval(checkInterval);
                    resolve();
                }, 1000);
            });
        }
    }

    async ensureCreated() {
        if (this.entityId !== null) return this.entityId;

        await this.waitForSave();

        if (this.entityId === null) {
            const payload = this.readPayload();
            try {
                this.createInFlight = true;
                const response = await api.post(this.options.apiPath, payload);
                if (response) {
                    const data = await response.json();
                    this.entityId = data.id;
                    if (this.options.onEntityCreated) {
                        this.options.onEntityCreated(this.entityId);
                    }
                }
            } finally {
                this.createInFlight = false;
            }
        }

        return this.entityId;
    }

    updateStatus(message) {
        if (this.statusContainer) {
            this.statusContainer.textContent = message;
        }
    }

    showError(message) {
        showError(message, this.errorContainer);
    }

    clearError() {
        hideError(this.errorContainer);
    }

    getEntityId() {
        return this.entityId;
    }

    destroy() {
        clearTimeout(this.saveTimer);
        this.options.fieldIds.forEach((id) => {
            const el = document.getElementById(id);
            if (!el) return;
            el.removeEventListener("input", this.scheduleSave);
        });
    }
}
