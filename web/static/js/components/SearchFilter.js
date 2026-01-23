import { debounce } from "../core/utils.js";
import { $ } from "../core/dom.js";

export default class SearchFilter {
    constructor(inputSelector, containerSelector, options = {}) {
        this.input = $(inputSelector);
        this.container = $(containerSelector);
        if (!this.input || !this.container) return;

        this.options = {
            searchFields: ["title"],
            debounceMs: 300,
            itemSelector: ".course-card",
            noResultsSelector: "#no-results",
            ...options,
        };

        this.noResults = $(this.options.noResultsSelector);
        this.performSearch = debounce(this.performSearch.bind(this), this.options.debounceMs);

        this.input.addEventListener("input", this.performSearch);
        this.performSearch();
    }

    performSearch() {
        const query = this.input.value.toLowerCase().trim();
        const items = this.container.querySelectorAll(this.options.itemSelector);
        let visibleCount = 0;

        items.forEach((item) => {
            const matches =
                query === "" ||
                this.options.searchFields.some((field) => {
                    const value = (item.dataset[field] || "").toLowerCase();
                    return value.includes(query);
                });

            if (matches) {
                item.style.display = "flex";
                visibleCount++;
            } else {
                item.style.display = "none";
            }
        });

        if (this.noResults) {
            if (visibleCount === 0 && query !== "") {
                this.noResults.style.display = "block";
                this.container.style.display = "none";
            } else {
                this.noResults.style.display = "none";
                this.container.style.display = "grid";
            }
        }
    }

    destroy() {
        this.input.removeEventListener("input", this.performSearch);
    }
}
