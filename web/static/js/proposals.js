document.addEventListener("DOMContentLoaded", () => {
    async function loadProposals() {
        try {
            const response = await fetch("/api/proposals");
            if (!response.ok) {
                if (response.status === 401) {
                    window.location.href = "/login";
                    return;
                }
                throw new Error("Failed to load proposals");
            }

            const proposals = await response.json();
            const container = document.getElementById("proposals-list");

            if (proposals.length === 0) {
                container.innerHTML =
                    '<div class="empty-state"><p>No proposals yet. <a href="/proposals/new">Create your first proposal</a></p></div>';
                return;
            }

            container.innerHTML = proposals
                .slice()
                .sort((a, b) => new Date(b.updated_at) - new Date(a.updated_at))
                .map(
                    (p) => `
                <div class="proposal-card" data-proposal-id="${p.id}">
                    <div class="proposal-header" >
                        <h3><a href="/proposals/${p.id}">${escapeHtml(p.title || "Untitled Proposal")}</a></h3>
                        <span class="status-badge status-${p.status}">${p.status}</span>
                    </div>
                    <p class="proposal-summary">${escapeHtml(p.summary)}</p>
                    <div class="proposal-meta">
                        <span>Created: ${new Date(p.created_at).toLocaleDateString()}</span>
                        <span>Updated: ${new Date(p.updated_at).toLocaleDateString()}</span>
                    </div>
                    <div class="proposal-actions">
                        <a href="/proposals/${p.id}/edit" class="btn btn-secondary">Edit</a>
                        <button class="btn btn-danger" data-delete-id="${p.id}">Delete</button>
                    </div>
                </div>
            `,
                )
                .join("");

            attachDeleteHandlers();
        } catch (error) {
            document.getElementById("proposals-list").innerHTML =
                '<div class="error-message">Failed to load proposals. Please refresh the page.</div>';
        }
    }

    function escapeHtml(text) {
        const div = document.createElement("div");
        div.textContent = text;
        return div.innerHTML;
    }

    function attachDeleteHandlers() {
        const container = document.getElementById("proposals-list");
        container.addEventListener("click", async (e) => {
            if (e.target.matches("[data-delete-id]")) {
                e.preventDefault();
                const proposalId = e.target.getAttribute("data-delete-id");
                const card = e.target.closest(".proposal-card");

                if (!confirm("Are you sure you want to delete this proposal? This action cannot be undone.")) {
                    return;
                }

                try {
                    const response = await fetch(`/api/proposals/${proposalId}`, {
                        method: "DELETE",
                    });

                    if (!response.ok) {
                        if (response.status === 401) {
                            window.location.href = "/login";
                            return;
                        }
                        if (response.status === 404) {
                            alert("Proposal not found");
                            return;
                        }
                        if (response.status === 403) {
                            alert("You don't have permission to delete this proposal");
                            return;
                        }
                        throw new Error("Failed to delete proposal");
                    }

                    if (card) {
                        card.remove();
                    }
                } catch (error) {
                    if (error.message === "Failed to fetch") {
                        alert("Network error. Please try again.");
                    } else {
                        alert("Failed to delete proposal. Please try again.");
                    }
                }
            }
        });
    }

    loadProposals();
});
