import Dropdown from "./components/Dropdown.js";
import Modal from "./components/Modal.js";
import MobileMenu from "./components/MobileMenu.js";
import api from "./core/api.js";
import { $, $$ } from "./core/dom.js";

async function handleLogout() {
    try {
        await api.post("/api/logout");
        window.location.href = "/";
    } catch (error) {
        window.location.href = "/";
    }
}

document.addEventListener("DOMContentLoaded", () => {
    if ($("#userDropdownMenu")) {
        new Dropdown(".user-menu-btn", "#userDropdownMenu");
    }
    if ($("#teachDropdownMenu")) {
        new Dropdown(".teach-menu-trigger", "#teachDropdownMenu");
    }
    if ($("#teach-modal")) {
        new Modal("#teach-modal");
    }
    if ($("#mobileMenu")) {
        new MobileMenu("#mobileMenu", "#mobileMenuOverlay", ".hamburger-btn");
    }

    $$("[data-logout]").forEach((btn) => {
        btn.addEventListener("click", handleLogout);
    });
});

export { handleLogout };
