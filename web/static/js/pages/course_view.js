import { $, on } from "../core/dom.js";
import api from "../core/api.js";
import { confirmAction } from "../core/utils.js";

document.addEventListener("DOMContentLoaded", () => {
    const accordionHeaders = document.querySelectorAll(
        ".module-accordion-header",
    );

    accordionHeaders.forEach((header) => {
        on(header, "click", () => {
            const accordion = header.closest(".module-accordion");
            if (accordion) {
                accordion.classList.toggle("expanded");
            }
        });
    });

    const enrollBtn = $("#enroll-btn");
    const unenrollBtn = $("#unenroll-btn");

    if (enrollBtn) {
        on(enrollBtn, "click", async () => {
            const courseId = enrollBtn.dataset.courseId;
            if (!courseId) return;

            const isAuthenticated = enrollBtn.dataset.authenticated === "true";
            if (!isAuthenticated) {
                const returnUrl = encodeURIComponent(window.location.pathname);
                window.location.href = `/login?next=${returnUrl}`;
                return;
            }

            const confirmed = await confirmAction(
                "You will be enrolled in this course and receive a confirmation email.",
                {
                    title: "Enroll in Course?",
                    confirmText: "Enroll",
                    confirmButtonClass: "btn-primary",
                    variant: "info",
                }
            );

            if (!confirmed) return;

            enrollBtn.disabled = true;
            enrollBtn.textContent = "Enrolling...";

            try {
                await api.post(`/api/courses/${courseId}/actions/enroll`);
                window.location.reload();
            } catch (error) {
                alert(error.message || "Failed to enroll");
                enrollBtn.disabled = false;
                enrollBtn.textContent = "Enroll";
            }
        });
    }

    if (unenrollBtn) {
        on(unenrollBtn, "click", async () => {
            const courseId = unenrollBtn.dataset.courseId;
            if (!courseId) return;

            const confirmed = await confirmAction(
                "You will lose access to your progress and any course materials. You can re-enroll at any time.",
                {
                    title: "Unenroll from Course?",
                    confirmText: "Unenroll",
                    confirmButtonClass: "btn-danger",
                    variant: "warning",
                }
            );

            if (!confirmed) return;

            unenrollBtn.disabled = true;
            unenrollBtn.textContent = "Unenrolling...";

            try {
                await api.delete(`/api/courses/${courseId}/actions/enroll`);
                window.location.reload();
            } catch (error) {
                alert(error.message || "Failed to unenroll");
                unenrollBtn.disabled = false;
                unenrollBtn.textContent = "Unenroll";
            }
        });
    }
});
