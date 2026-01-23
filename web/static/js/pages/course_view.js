import { $, on } from "../core/dom.js";

document.addEventListener("DOMContentLoaded", () => {
    const accordionHeaders = document.querySelectorAll(".module-accordion-header");
    
    accordionHeaders.forEach((header) => {
        on(header, "click", () => {
            const accordion = header.closest(".module-accordion");
            if (accordion) {
                accordion.classList.toggle("expanded");
            }
        });
    });
});
