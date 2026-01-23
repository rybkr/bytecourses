import api from "../core/api.js";
import FormHandler from "../components/FormHandler.js";
import { escapeHtml, confirmAction, showError, hideError } from "../core/utils.js";
import { $, on } from "../core/dom.js";

document.addEventListener("DOMContentLoaded", () => {
    const pathMatch = window.location.pathname.match(/^\/courses\/(\d+)/);
    const courseId = pathMatch ? Number(pathMatch[1]) : null;
    if (!courseId || !Number.isFinite(courseId)) return;

    const editToggleBtn = $("#edit-toggle-btn");
    const courseForm = $("#course-form");
    const viewModeElements = document.querySelectorAll(".view-mode");
    const editModeElements = document.querySelectorAll(".edit-mode");
    const courseViewMain = $(".course-view-main");
    const errorDiv = $("#error-message");
    const saveBtn = $("#saveBtn");
    const publishBtn = $("#publishBtn");

    let isEditMode = false;
    let formHandler = null;

    if (!editToggleBtn) return;

    const fieldIds = ["title", "summary", "target_audience", "learning_objectives", "assumed_prerequisites"];

    function initializeFormHandler() {
        if (!courseForm) return null;

        courseForm.addEventListener("submit", (e) => e.preventDefault());

        return new FormHandler("#course-form", {
            apiPath: "/api/courses",
            entityId: courseId,
            autosaveDelay: 2000,
            fieldIds: fieldIds,
            errorContainer: "#error-message",
            statusContainer: "#save-status",
        });
    }

    function enterEditMode() {
        isEditMode = true;
        editToggleBtn.textContent = "Cancel";

        const viewTitle = $("#view-title");
        const heroTitle = $("#hero-title");
        const viewSummary = $("#view-summary");
        const heroSummary = $("#hero-summary");

        if (viewTitle && heroTitle) {
            viewTitle.style.display = "none";
            heroTitle.style.display = "block";
            heroTitle.value = viewTitle.textContent.trim();
        }

        if (viewSummary && heroSummary) {
            viewSummary.style.display = "none";
            heroSummary.style.display = "block";
            heroSummary.value = viewSummary.textContent.trim();
        }

        viewModeElements.forEach((el) => {
            if (el && el.id !== "view-title" && el.id !== "view-summary") {
                el.style.display = "none";
            }
        });
        editModeElements.forEach((el) => {
            if (el && el.id !== "hero-title" && el.id !== "hero-summary") {
                el.style.display = "";
            }
        });

        if (courseForm) {
            courseForm.style.display = "block";
        }
        if (courseViewMain) {
            courseViewMain.style.display = "none";
        }

        if (!formHandler) {
            formHandler = initializeFormHandler();
        }

        const formTitle = $("#title");
        const formSummary = $("#summary");
        if (formTitle && heroTitle) {
            formTitle.value = heroTitle.value;
        }
        if (formSummary && heroSummary) {
            formSummary.value = heroSummary.value;
        }

        if (formHandler && heroTitle && formTitle) {
            heroTitle.addEventListener("input", syncHeroToForm);
        }
        if (formHandler && heroSummary && formSummary) {
            heroSummary.addEventListener("input", syncHeroToForm);
        }
    }

    function syncHeroToForm() {
        const heroTitle = $("#hero-title");
        const heroSummary = $("#hero-summary");
        const formTitle = $("#title");
        const formSummary = $("#summary");

        if (heroTitle && formTitle) {
            formTitle.value = heroTitle.value;
            formTitle.dispatchEvent(new Event("input", { bubbles: true }));
        }
        if (heroSummary && formSummary) {
            formSummary.value = heroSummary.value;
            formSummary.dispatchEvent(new Event("input", { bubbles: true }));
        }
    }

    function exitEditMode() {
        isEditMode = false;
        editToggleBtn.textContent = "Edit";

        const heroTitle = $("#hero-title");
        const heroSummary = $("#hero-summary");
        if (heroTitle) {
            heroTitle.removeEventListener("input", syncHeroToForm);
        }
        if (heroSummary) {
            heroSummary.removeEventListener("input", syncHeroToForm);
        }

        const viewTitle = $("#view-title");
        const viewSummary = $("#view-summary");

        if (viewTitle && heroTitle) {
            viewTitle.style.display = "";
            heroTitle.style.display = "none";
        }

        if (viewSummary && heroSummary) {
            viewSummary.style.display = "";
            heroSummary.style.display = "none";
        }

        viewModeElements.forEach((el) => {
            if (el && el.id !== "view-title" && el.id !== "view-summary") {
                el.style.display = "";
            }
        });
        editModeElements.forEach((el) => {
            if (el && el.id !== "hero-title" && el.id !== "hero-summary") {
                el.style.display = "none";
            }
        });

        if (courseForm) {
            courseForm.style.display = "none";
        }
        if (courseViewMain) {
            courseViewMain.style.display = "";
        }

        if (formHandler) {
            formHandler.saveNow().then(() => {
                window.location.reload();
            }).catch(() => {
                if (confirm("You have unsaved changes. Reload anyway?")) {
                    window.location.reload();
                }
            });
        } else {
            window.location.reload();
        }
    }

    if (editToggleBtn) {
        editToggleBtn.addEventListener("click", () => {
            if (isEditMode) {
                exitEditMode();
            } else {
                enterEditMode();
            }
        });
    }

    if (saveBtn) {
        on(saveBtn, "click", async (e) => {
            e.preventDefault();
            if (!formHandler) {
                formHandler = initializeFormHandler();
            }
            if (formHandler) {
                try {
                    await formHandler.saveNow();
                    exitEditMode();
                } catch (err) {
                    showError("Save failed", errorDiv);
                }
            }
        });
    }

    async function publish() {
        if (formHandler) {
            await formHandler.saveNow();
        }

        hideError(errorDiv);
        if (publishBtn) publishBtn.disabled = true;

        try {
            await api.post(`/api/courses/${courseId}/publish`);
            window.location.reload();
        } catch (error) {
            showError(error.message || "Publish failed", errorDiv);
            if (publishBtn) publishBtn.disabled = false;
        }
    }

    if (publishBtn) {
        on(publishBtn, "click", (e) => {
            e.preventDefault();
            publish().catch(() => {
                showError("Publish failed", errorDiv);
                if (publishBtn) publishBtn.disabled = false;
            });
        });
    }

    document.addEventListener("visibilitychange", () => {
        if (document.visibilityState === "hidden" && isEditMode && formHandler) {
            formHandler.saveNow().catch(() => {});
        }
    });
});
