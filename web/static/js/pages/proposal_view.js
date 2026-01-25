import api from "../core/api.js";
import Modal from "../components/Modal.js";
import { $ } from "../core/dom.js";
import { showError, hideError, confirmAction } from "../core/utils.js";

document.addEventListener("DOMContentLoaded", () => {
    const proposalIdElement = document.querySelector("[data-proposal-id]");
    if (!proposalIdElement) {
        console.error("Proposal ID not found");
        return;
    }

    const proposalId = Number(proposalIdElement.dataset.proposalId);
    if (!proposalId || proposalId <= 0) {
        console.error("Invalid proposal ID:", proposalId);
        return;
    }

    const submitBtn = $("#submitBtn");
    const withdrawBtn = $("#withdrawBtn");
    const deleteBtn = $("#deleteBtn");
    const createCourseBtn = $("#createCourseBtn");
    const errorDiv = $("#error-message");
    const createCourseModal = $("#create-course-modal")
        ? new Modal("#create-course-modal")
        : null;

    if (submitBtn) {
        async function submit() {
            hideError(errorDiv);
            submitBtn.disabled = true;

            try {
                await api.post(`/api/proposals/${proposalId}/actions/submit`);
                window.location.reload();
            } catch (error) {
                showError(error.message || "Submit failed", errorDiv);
                submitBtn.disabled = false;
            }
        }

        submitBtn.addEventListener("click", submit);
    }

    if (withdrawBtn) {
        async function withdraw() {
            if (
                !confirmAction(
                    "Are you sure you want to withdraw this proposal? It will be removed from review.",
                )
            ) {
                return;
            }

            hideError(errorDiv);
            withdrawBtn.disabled = true;

            try {
                await api.post(`/api/proposals/${proposalId}/actions/withdraw`);
                window.location.reload();
            } catch (error) {
                showError(error.message || "Withdraw failed", errorDiv);
                withdrawBtn.disabled = false;
            }
        }

        withdrawBtn.addEventListener("click", withdraw);
    }

    if (deleteBtn) {
        async function deleteProposal() {
            if (
                !confirmAction(
                    "Are you sure you want to delete this proposal? This action cannot be undone.",
                )
            ) {
                return;
            }

            hideError(errorDiv);
            deleteBtn.disabled = true;

            try {
                await api.delete(`/api/proposals/${proposalId}`);
                window.location.href = "/proposals";
            } catch (error) {
                showError(error.message || "Delete failed", errorDiv);
                deleteBtn.disabled = false;
            }
        }

        deleteBtn.addEventListener("click", deleteProposal);
    }

    if (createCourseBtn && createCourseModal) {
        createCourseBtn.addEventListener("click", () => {
            createCourseModal.open();
        });
    }

    const confirmCreateCourseBtn = $("#confirm-create-course");
    if (confirmCreateCourseBtn) {
        async function createCourse() {
            hideError(errorDiv);
            confirmCreateCourseBtn.disabled = true;

            try {
                const response = await api.post(
                    `/api/proposals/${proposalId}/actions/create-course`,
                );
                if (response) {
                    const course = await response.json();
                    if (createCourseModal) {
                        createCourseModal.close();
                    }
                    window.location.href = `/courses/${course.id}/edit`;
                }
            } catch (error) {
                if (error.message && error.message.includes("409")) {
                    try {
                        const response = await fetch(
                            `/api/proposals/${proposalId}/actions/create-course`,
                            {
                                method: "POST",
                            },
                        );
                        if (response.status === 409) {
                            const data = await response.json();
                            if (data.course_id) {
                                window.location.href = `/courses/${data.course_id}/edit`;
                                return;
                            }
                        }
                    } catch (e) {
                        // Fall through to show error
                    }
                }
                showError(error.message || "Create course failed", errorDiv);
                confirmCreateCourseBtn.disabled = false;
            }
        }

        confirmCreateCourseBtn.addEventListener("click", createCourse);
    }

    const approveBtn = $("#approveBtn");
    const requestChangesBtn = $("#requestChangesBtn");
    const rejectBtn = $("#rejectBtn");
    const reviewErrorDiv = $("#review-error");
    const reviewNotes = $("#review-notes");

    async function handleReviewAction(action) {
        if (!proposalId || proposalId <= 0) {
            showError("Invalid proposal ID", reviewErrorDiv);
            return;
        }

        hideError(reviewErrorDiv);

        const buttons = [approveBtn, requestChangesBtn, rejectBtn];
        buttons.forEach((btn) => {
            if (btn) btn.disabled = true;
        });

        try {
            await api.post(`/api/proposals/${proposalId}/actions/${action}`, {
                review_notes: reviewNotes ? reviewNotes.value : "",
            });
            window.location.reload();
        } catch (error) {
            let errorMsg = error.message || "Action failed";
            if (error.message && error.message.includes("404")) {
                errorMsg =
                    "Proposal not found. Please refresh the page and try again.";
            } else if (error.message && error.message.includes("403")) {
                errorMsg = "You don't have permission to perform this action.";
            } else if (error.message && error.message.includes("409")) {
                errorMsg =
                    "Proposal status has changed. Please refresh the page.";
            }
            showError(errorMsg, reviewErrorDiv);
            buttons.forEach((btn) => {
                if (btn) btn.disabled = false;
            });
        }
    }

    if (approveBtn) {
        approveBtn.addEventListener("click", () =>
            handleReviewAction("approve"),
        );
    }
    if (requestChangesBtn) {
        requestChangesBtn.addEventListener("click", () =>
            handleReviewAction("request-changes"),
        );
    }
    if (rejectBtn) {
        rejectBtn.addEventListener("click", () => handleReviewAction("reject"));
    }
});
