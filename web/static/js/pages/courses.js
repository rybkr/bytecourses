import SearchFilter from "../components/SearchFilter.js";

document.addEventListener("DOMContentLoaded", () => {
    new SearchFilter("#course-search", "#course-grid", {
        searchFields: ["title", "summary"],
        itemSelector: ".course-card",
        noResultsSelector: "#no-results",
    });
});
