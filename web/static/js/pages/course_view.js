import { $, on } from "../core/dom.js";
import api from "../core/api.js";

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

            if (!confirm("Are you sure you want to unenroll from this course?")) {
                return;
            }

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
