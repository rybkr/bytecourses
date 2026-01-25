import api from "../core/api.js";
import { escapeHtml, confirmAction } from "../core/utils.js";
import { $, delegate } from "../core/dom.js";

document.addEventListener("DOMContentLoaded", () => {
    const container = $("#proposals-list");
    if (!container) return;

    const isAdmin = container.getAttribute("data-is-admin") === "true";

    async function loadProposals() {
        try {
            const response = await api.get("/api/proposals");
            if (!response) return;

            const proposals = await response.json();

            if (proposals.length === 0) {
                if (isAdmin) {
                    container.innerHTML =
                        '<div class="empty-state"><p>No proposals have been submitted for review.</p></div>';
                } else {
                    container.innerHTML =
                        '<div class="empty-state"><p>No proposals yet. <a href="/proposals/new">Create your first proposal</a></p></div>';
                }
                return;
            }

            container.innerHTML = proposals
                .slice()
                .sort((a, b) => new Date(b.updated_at) - new Date(a.updated_at))
                .map((p) => {
                    let actionsHtml = "";
                    if (!isAdmin) {
                        const actions = [];
                        if (
                            p.status === "draft" ||
                            p.status === "changes_requested"
                        ) {
                            actions.push(
                                `<a href="/proposals/${p.id}/edit" class="btn btn-secondary btn-sm">Edit</a>`,
                            );
                        }
                        if (p.status === "submitted") {
                            actions.push(
                                `<button class="btn btn-secondary btn-sm" data-withdraw-id="${p.id}">Withdraw</button>`,
                            );
                        }
                        if (
                            p.status === "draft" ||
                            p.status === "withdrawn" ||
                            p.status === "rejected"
                        ) {
                            actions.push(
                                `<button class="btn btn-danger btn-sm" data-delete-id="${p.id}">Delete</button>`,
                            );
                        }
                        if (actions.length > 0) {
                            actionsHtml = `<div class="proposal-actions">${actions.join("")}</div>`;
                        }
                    }

                    const authorHtml = isAdmin
                        ? `<div class="proposal-author">Author ID: ${p.author_id}</div>`
                        : "";

                    return `
                        <div class="proposal-card" data-proposal-id="${p.id}">
                            <div class="proposal-header">
                                <h3><a href="/proposals/${p.id}">${escapeHtml(p.title || "Untitled Proposal")}</a></h3>
                                <span class="status-badge status-${p.status}">${p.status}</span>
                            </div>
                            <p class="proposal-summary">${escapeHtml(p.summary)}</p>
                            ${authorHtml}
                            <div class="proposal-meta">
                                <span>Created: ${new Date(p.created_at).toLocaleDateString()}</span>
                                <span>Updated: ${new Date(p.updated_at).toLocaleDateString()}</span>
                            </div>
                            ${actionsHtml}
                        </div>
                    `;
                })
                .join("");
        } catch (error) {
            container.innerHTML =
                '<div class="error-message">Failed to load proposals. Please refresh the page.</div>';
        }
    }

    delegate(container, "[data-delete-id]", "click", async (e, target) => {
        e.preventDefault();
        const proposalId = target.getAttribute("data-delete-id");
        const card = target.closest(".proposal-card");

        if (
            !confirmAction(
                "Are you sure you want to delete this proposal? This action cannot be undone.",
            )
        ) {
            return;
        }

        try {
            await api.delete(`/api/proposals/${proposalId}`);
            if (card) {
                card.remove();
            }
        } catch (error) {
            if (error.message === "Failed to fetch") {
                alert("Network error. Please try again.");
            } else {
                alert(
                    error.message ||
                        "Failed to delete proposal. Please try again.",
                );
            }
        }
    });

    delegate(container, "[data-withdraw-id]", "click", async (e, target) => {
        e.preventDefault();
        const proposalId = target.getAttribute("data-withdraw-id");

        if (
            !confirmAction(
                "Are you sure you want to withdraw this proposal? It will be removed from review.",
            )
        ) {
            return;
        }

        try {
            await api.post(`/api/proposals/${proposalId}/actions/withdraw`);
            loadProposals();
        } catch (error) {
            if (error.message === "Failed to fetch") {
                alert("Network error. Please try again.");
            } else {
                alert(
                    error.message ||
                        "Failed to withdraw proposal. Please try again.",
                );
            }
        }
    });

    loadProposals();
});
