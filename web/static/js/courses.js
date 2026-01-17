document.addEventListener("DOMContentLoaded", () => {
    const searchInput = document.getElementById("course-search");
    const courseGrid = document.getElementById("course-grid");
    const noResults = document.getElementById("no-results");

    if (!searchInput || !courseGrid) {
        return;
    }

    let searchTimer = null;

    function performSearch() {
        const query = searchInput.value.toLowerCase().trim();
        const cards = courseGrid.querySelectorAll(".course-card");
        let visibleCount = 0;

        cards.forEach((card) => {
            const title = (card.dataset.title || "").toLowerCase();
            const summary = (card.dataset.summary || "").toLowerCase();
            const matches = query === "" || title.includes(query) || summary.includes(query);

            if (matches) {
                card.style.display = "flex";
                visibleCount++;
            } else {
                card.style.display = "none";
            }
        });

        if (visibleCount === 0 && query !== "") {
            noResults.style.display = "block";
            courseGrid.style.display = "none";
        } else {
            noResults.style.display = "none";
            courseGrid.style.display = "grid";
        }
    }

    searchInput.addEventListener("input", () => {
        clearTimeout(searchTimer);
        searchTimer = setTimeout(performSearch, 300);
    });

    // Initial search to handle any pre-filled values
    performSearch();
});
