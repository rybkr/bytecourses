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
                <div class="proposal-card">
                    <div class="proposal-header" >
                        <h3><a href="/proposals/${p.id}">${escapeHtml(p.title || "Untitled Proposal")}</a></h3>
                        <span class="status-badge status-${p.status}">${p.status}</span>
                    </div>
                    <p class="proposal-summary">${escapeHtml(p.summary)}</p>
                    <div class="proposal-meta">
                        <span>Created: ${new Date(p.created_at).toLocaleDateString()}</span>
                        <span>Updated: ${new Date(p.updated_at).toLocaleDateString()}</span>
                    </div>
                </div>
            `,
                )
                .join("");
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

    loadProposals();
});
