import api from "../core/api.js";
import { $, on } from "../core/dom.js";

document.addEventListener("DOMContentLoaded", () => {
    const pathMatch = window.location.pathname.match(
        /\/courses\/(\d+)\/modules\/(\d+)/,
    );

    if (!pathMatch) return;

    const courseId = Number(pathMatch[1]);
    const moduleId = Number(pathMatch[2]);

    if (!courseId || !moduleId) return;

    function showToast(message, type = "info") {
        const existing = document.querySelector(".toast");
        if (existing) existing.remove();

        const toast = document.createElement("div");
        toast.className = `toast toast-${type}`;
        toast.textContent = message;
        document.body.appendChild(toast);

        requestAnimationFrame(() => {
            toast.classList.add("show");
        });

        setTimeout(() => {
            toast.classList.remove("show");
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }
});
