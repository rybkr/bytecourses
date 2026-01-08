async function handleLogout() {
    try {
        const response = await fetch("/api/logout", {
            method: "POST",
        });

        if (response.ok || response.status === 204) {
            window.location.href = "/";
        }
    } catch (error) {
        console.error("Logout error:", error);
    }
}

function escapeHtml(text) {
    if (!text) return "";
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
}

function formatDate(dateString) {
    if (!dateString) return "";
    const date = new Date(dateString);
    return date.toLocaleDateString("en-US", {
        year: "numeric",
        month: "long",
        day: "numeric",
    });
}

// Global variable to track the current close handler
let currentCloseHandler = null;

function toggleUserMenu(event) {
    event.stopPropagation();
    const menu = document.getElementById("userDropdownMenu");
    const button = event.currentTarget;
    if (!menu || !button) return;

    const isOpen = menu.classList.contains("show");

    // Remove previous close handler if exists
    if (currentCloseHandler) {
        document.removeEventListener("click", currentCloseHandler);
        currentCloseHandler = null;
    }

    // Close all dropdowns
    document.querySelectorAll(".user-dropdown-menu").forEach(m => {
        m.classList.remove("show");
    });
    document.querySelectorAll(".user-menu-btn").forEach(btn => {
        btn.setAttribute("aria-expanded", "false");
    });

    // Toggle this dropdown
    if (!isOpen) {
        menu.classList.add("show");
        button.setAttribute("aria-expanded", "true");

        // Close on outside click
        currentCloseHandler = function closeMenu(e) {
            if (!menu.contains(e.target) && !e.target.closest(".user-menu-btn")) {
                menu.classList.remove("show");
                button.setAttribute("aria-expanded", "false");
                document.removeEventListener("click", currentCloseHandler);
                currentCloseHandler = null;
            }
        };

        // Use setTimeout to avoid immediate closure
        setTimeout(() => {
            document.addEventListener("click", currentCloseHandler);
        }, 10);
    } else {
        button.setAttribute("aria-expanded", "false");
    }
}

// Close dropdown when clicking on menu items (for better UX)
document.addEventListener("DOMContentLoaded", function () {
    const menuItems = document.querySelectorAll(".user-dropdown-item");
    menuItems.forEach(item => {
        item.addEventListener("click", function () {
            const menu = document.getElementById("userDropdownMenu");
            if (menu) {
                setTimeout(() => {
                    menu.classList.remove("show");
                    if (currentCloseHandler) {
                        document.removeEventListener("click", currentCloseHandler);
                        currentCloseHandler = null;
                    }
                }, 100);
            }
        });
    });
});

function openTeachModal(e) {
    if (e) {
        e.preventDefault();
    }
    const modal = document.getElementById("teach-modal");
    if (modal) {
        modal.style.display = "flex";
        document.body.style.overflow = "hidden";
    }
}

function closeTeachModal() {
    const modal = document.getElementById("teach-modal");
    if (modal) {
        modal.style.display = "none";
        document.body.style.overflow = "";
    }
}

document.addEventListener("keydown", (e) => {
    if (e.key === "Escape") {
        closeTeachModal();
    }
});
