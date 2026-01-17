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

let currentCloseHandler = null;
let currentTeachCloseHandler = null;

function toggleUserMenu(event) {
    event.stopPropagation();
    const menu = document.getElementById("userDropdownMenu");
    const button = event.currentTarget;
    if (!menu || !button) return;

    const isOpen = menu.classList.contains("show");

    if (currentCloseHandler) {
        document.removeEventListener("click", currentCloseHandler);
        currentCloseHandler = null;
    }
    if (currentTeachCloseHandler) {
        document.removeEventListener("click", currentTeachCloseHandler);
        currentTeachCloseHandler = null;
    }
    document.querySelectorAll(".user-dropdown-menu").forEach(m => {
        m.classList.remove("show");
    });
    document.querySelectorAll(".user-menu-btn").forEach(btn => {
        btn.setAttribute("aria-expanded", "false");
    });
    document.querySelectorAll(".teach-dropdown-menu").forEach(m => {
        m.classList.remove("show");
    });

    if (!isOpen) {
        menu.classList.add("show");
        button.setAttribute("aria-expanded", "true");
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

function toggleTeachMenu(event) {
    event.stopPropagation();
    const menu = document.getElementById("teachDropdownMenu");
    const trigger = event.currentTarget;
    if (!menu || !trigger) return;

    const isOpen = menu.classList.contains("show");

    if (currentTeachCloseHandler) {
        document.removeEventListener("click", currentTeachCloseHandler);
        currentTeachCloseHandler = null;
    }
    if (currentCloseHandler) {
        document.removeEventListener("click", currentCloseHandler);
        currentCloseHandler = null;
    }
    document.querySelectorAll(".teach-dropdown-menu").forEach(m => {
        m.classList.remove("show");
    });
    document.querySelectorAll(".user-dropdown-menu").forEach(m => {
        m.classList.remove("show");
    });
    document.querySelectorAll(".user-menu-btn").forEach(btn => {
        btn.setAttribute("aria-expanded", "false");
    });

    if (!isOpen) {
        menu.classList.add("show");
        currentTeachCloseHandler = function closeMenu(e) {
            if (!menu.contains(e.target) && !e.target.closest(".teach-menu-trigger")) {
                menu.classList.remove("show");
                document.removeEventListener("click", currentTeachCloseHandler);
                currentTeachCloseHandler = null;
            }
        };

        // Use setTimeout to avoid immediate closure
        setTimeout(() => {
            document.addEventListener("click", currentTeachCloseHandler);
        }, 10);
    }
}

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

    const teachMenuItems = document.querySelectorAll(".teach-dropdown-item");
    teachMenuItems.forEach(item => {
        item.addEventListener("click", function () {
            const menu = document.getElementById("teachDropdownMenu");
            if (menu) {
                setTimeout(() => {
                    menu.classList.remove("show");
                    if (currentTeachCloseHandler) {
                        document.removeEventListener("click", currentTeachCloseHandler);
                        currentTeachCloseHandler = null;
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
        closeMobileMenu();
    }
});

function toggleMobileMenu(event) {
    if (event) {
        event.stopPropagation();
    }
    
    const menu = document.getElementById("mobileMenu");
    const overlay = document.getElementById("mobileMenuOverlay");
    const button = document.querySelector(".hamburger-btn");
    
    if (!menu || !overlay || !button) return;
    
    const isOpen = menu.classList.contains("active");
    
    document.querySelectorAll(".user-dropdown-menu").forEach(m => {
        m.classList.remove("show");
    });
    document.querySelectorAll(".teach-dropdown-menu").forEach(m => {
        m.classList.remove("show");
    });
    document.querySelectorAll(".user-menu-btn").forEach(btn => {
        btn.setAttribute("aria-expanded", "false");
    });
    
    if (currentCloseHandler) {
        document.removeEventListener("click", currentCloseHandler);
        currentCloseHandler = null;
    }
    if (currentTeachCloseHandler) {
        document.removeEventListener("click", currentTeachCloseHandler);
        currentTeachCloseHandler = null;
    }
    
    if (!isOpen) {
        menu.classList.add("active");
        overlay.classList.add("active");
        button.classList.add("active");
        button.setAttribute("aria-expanded", "true");
        document.body.style.overflow = "hidden";
    } else {
        closeMobileMenu();
    }
}

function closeMobileMenu() {
    const menu = document.getElementById("mobileMenu");
    const overlay = document.getElementById("mobileMenuOverlay");
    const button = document.querySelector(".hamburger-btn");
    
    if (menu) {
        menu.classList.remove("active");
    }
    if (overlay) {
        overlay.classList.remove("active");
    }
    if (button) {
        button.classList.remove("active");
        button.setAttribute("aria-expanded", "false");
    }
    document.body.style.overflow = "";
}
