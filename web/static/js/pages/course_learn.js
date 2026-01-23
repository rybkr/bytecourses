import { $, on } from "../core/dom.js";

document.addEventListener("DOMContentLoaded", () => {
    const sidebar = $(".course-learn-sidebar");
    if (sidebar) {
        const activeReading = $(".course-learn-reading.active");
        if (activeReading) {
            activeReading.scrollIntoView({ behavior: "smooth", block: "center" });
        }
    }
});
