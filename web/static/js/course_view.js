document.addEventListener("DOMContentLoaded", () => {
    const accordions = document.querySelectorAll(".module-accordion");

    accordions.forEach((accordion) => {
        const header = accordion.querySelector(".module-accordion-header");
        if (!header) return;

        header.addEventListener("click", () => {
            const isExpanded = accordion.classList.contains("expanded");
            
            if (isExpanded) {
                accordion.classList.remove("expanded");
                accordion.setAttribute("aria-expanded", "false");
            } else {
                accordion.classList.add("expanded");
                accordion.setAttribute("aria-expanded", "true");
            }
        });

        header.addEventListener("keydown", (e) => {
            if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                header.click();
            }
        });

        // Set initial aria-expanded state
        const isInitiallyExpanded = accordion.classList.contains("expanded");
        accordion.setAttribute("aria-expanded", isInitiallyExpanded ? "true" : "false");
    });
});
