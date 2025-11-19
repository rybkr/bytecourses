function escapeHtml(text) {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
}

const navbarModule = {
    currentUser: null,

    init() {
        this.initElements();
        this.initEventListeners();
        this.checkAuthState();
    },

    initElements() {
        this.mainNav = document.getElementById("mainNav");
        this.userInfo = document.getElementById("userInfo");
        this.mobileUserInfo = document.getElementById("mobileUserInfo");
        this.hamburgerBtn = document.getElementById("hamburgerBtn");
        this.mobileNav = document.getElementById("mobileNav");
        this.mobileOverlay = document.getElementById("mobileOverlay");

        this.homeBtn = document.getElementById("homeBtn");
        this.viewCoursesBtn = document.getElementById("viewCoursesBtn");
        this.myCoursesBtn = document.getElementById("myCoursesBtn");
        this.profileBtn = document.getElementById("profileBtn");
        this.adminBtn = document.getElementById("adminBtn");
        this.logoutBtn = document.getElementById("logoutBtn");

        this.mobileHomeBtn = document.getElementById("mobileHomeBtn");
        this.mobileViewCoursesBtn = document.getElementById("mobileViewCoursesBtn");
        this.mobileMyCoursesBtn = document.getElementById("mobileMyCoursesBtn");
        this.mobileProfileBtn = document.getElementById("mobileProfileBtn");
        this.mobileAdminBtn = document.getElementById("mobileAdminBtn");
        this.mobileLogoutBtn = document.getElementById("mobileLogoutBtn");
    },

    initEventListeners() {
        if (this.homeBtn) {
            this.homeBtn.addEventListener("click", () => {
                window.location.href = "/";
            });
        }

        if (this.viewCoursesBtn) {
            this.viewCoursesBtn.addEventListener("click", () => {
                window.location.href = "/";
            });
        }

        if (this.profileBtn) {
            this.profileBtn.addEventListener("click", () => {
                window.location.href = "/";
            });
        }

        if (this.logoutBtn) {
            this.logoutBtn.addEventListener("click", () => this.handleLogout());
        }

        if (this.mobileHomeBtn) {
            this.mobileHomeBtn.addEventListener("click", () => {
                window.location.href = "/";
            });
        }

        if (this.mobileViewCoursesBtn) {
            this.mobileViewCoursesBtn.addEventListener("click", () => {
                window.location.href = "/";
            });
        }

        if (this.mobileProfileBtn) {
            this.mobileProfileBtn.addEventListener("click", () => {
                window.location.href = "/";
            });
        }

        if (this.mobileLogoutBtn) {
            this.mobileLogoutBtn.addEventListener("click", () => this.handleLogout());
        }

        if (this.hamburgerBtn) {
            this.hamburgerBtn.addEventListener("click", () => this.toggleMobileMenu());
        }

        if (this.mobileOverlay) {
            this.mobileOverlay.addEventListener("click", () => this.closeMobileMenu());
        }
    },

    parseJWT(token) {
        try {
            const base64Url = token.split(".")[1];
            const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
            const jsonPayload = decodeURIComponent(
                atob(base64)
                    .split("")
                    .map(function (c) {
                        return "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2);
                    })
                    .join(""),
            );
            return JSON.parse(jsonPayload);
        } catch (e) {
            return null;
        }
    },

    checkAuthState() {
        const token = sessionStorage.getItem("authToken");
        if (!token) {
            this.updateNavbarVisibility(false, null);
            return;
        }

        const payload = this.parseJWT(token);
        if (!payload) {
            sessionStorage.removeItem("authToken");
            this.updateNavbarVisibility(false, null);
            return;
        }

        if (payload.exp && payload.exp * 1000 < Date.now()) {
            sessionStorage.removeItem("authToken");
            this.updateNavbarVisibility(false, null);
            return;
        }

        this.currentUser = {
            id: payload.user_id,
            email: payload.email,
            role: payload.role,
        };

        this.updateNavbarVisibility(true, this.currentUser);
    },

    updateNavbarVisibility(isAuthenticated, user) {
        if (!this.mainNav) return;

        if (isAuthenticated && user) {
            this.mainNav.style.display = "flex";
            if (this.userInfo) {
                this.userInfo.style.display = "block";
                this.userInfo.textContent = `Logged in as ${user.email} (${user.role})`;
            }
            if (this.mobileUserInfo) {
                this.mobileUserInfo.textContent = `Logged in as ${user.email} (${user.role})`;
            }

            if (user.role === "admin" && this.adminBtn) {
                this.adminBtn.style.display = "inline-block";
            }
            if (user.role === "admin" && this.mobileAdminBtn) {
                this.mobileAdminBtn.style.display = "block";
            }

            if ((user.role === "instructor" || user.role === "admin") && this.myCoursesBtn) {
                this.myCoursesBtn.style.display = "inline-block";
            }
            if ((user.role === "instructor" || user.role === "admin") && this.mobileMyCoursesBtn) {
                this.mobileMyCoursesBtn.style.display = "block";
            }
        } else {
            this.mainNav.style.display = "flex";
            if (this.userInfo) {
                this.userInfo.style.display = "none";
            }
            if (this.adminBtn) {
                this.adminBtn.style.display = "none";
            }
            if (this.myCoursesBtn) {
                this.myCoursesBtn.style.display = "none";
            }
            if (this.mobileAdminBtn) {
                this.mobileAdminBtn.style.display = "none";
            }
            if (this.mobileMyCoursesBtn) {
                this.mobileMyCoursesBtn.style.display = "none";
            }
        }
    },

    handleLogout() {
        sessionStorage.removeItem("authToken");
        window.location.href = "/";
    },

    toggleMobileMenu() {
        if (!this.mobileNav) return;
        const isOpen = this.mobileNav.classList.contains("open");
        if (isOpen) {
            this.closeMobileMenu();
        } else {
            this.openMobileMenu();
        }
    },

    openMobileMenu() {
        if (this.mobileNav) this.mobileNav.classList.add("open");
        if (this.mobileOverlay) this.mobileOverlay.classList.add("open");
        if (this.hamburgerBtn) this.hamburgerBtn.classList.add("open");
        document.body.style.overflow = "hidden";
    },

    closeMobileMenu() {
        if (this.mobileNav) this.mobileNav.classList.remove("open");
        if (this.mobileOverlay) this.mobileOverlay.classList.remove("open");
        if (this.hamburgerBtn) this.hamburgerBtn.classList.remove("open");
        document.body.style.overflow = "";
    },
};

document.addEventListener("DOMContentLoaded", () => navbarModule.init());

