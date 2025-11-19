function escapeHtml(text) {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
}

const navbarModule = {
    currentUser: null,
    onNavigate: null,
    dropdownOpen: false,
    clickOutsideHandler: null,

    init() {
        this.initElements();
        this.initEventListeners();
        this.initProfileDropdown();
        this.checkAuthState();
    },

    initElements() {
        this.mainNav = document.getElementById("mainNav");
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

        this.profileDropdown = document.getElementById("profileDropdown");
        this.profileTrigger = document.getElementById("profileTrigger");
        this.profileTriggerText = document.getElementById("profileTriggerText");
        this.profileMenu = document.getElementById("profileMenu");
        this.profileViewBtn = document.getElementById("profileViewBtn");
        this.profileLogoutBtn = document.getElementById("profileLogoutBtn");
    },

    initEventListeners() {
        if (this.homeBtn) {
            this.homeBtn.addEventListener("click", () => {
                if (this.onNavigate) {
                    this.onNavigate("home");
                } else {
                    window.location.href = "/";
                }
            });
        }

        if (this.viewCoursesBtn) {
            this.viewCoursesBtn.addEventListener("click", () => {
                if (this.onNavigate) {
                    this.onNavigate("courses");
                } else {
                    window.location.href = "/";
                }
            });
        }

        if (this.profileBtn) {
            this.profileBtn.addEventListener("click", () => {
                window.location.href = "/profile/";
            });
        }

        if (this.logoutBtn) {
            this.logoutBtn.addEventListener("click", () => this.handleLogout());
        }

        if (this.myCoursesBtn) {
            this.myCoursesBtn.addEventListener("click", () => {
                if (this.onNavigate) {
                    this.onNavigate("myCourses");
                } else {
                    window.location.href = "/";
                }
            });
        }

        if (this.adminBtn) {
            this.adminBtn.addEventListener("click", () => {
                if (this.onNavigate) {
                    this.onNavigate("admin");
                } else {
                    window.location.href = "/";
                }
            });
        }

        if (this.mobileHomeBtn) {
            this.mobileHomeBtn.addEventListener("click", () => {
                if (this.onNavigate) {
                    this.onNavigate("home");
                } else {
                    window.location.href = "/";
                }
            });
        }

        if (this.mobileViewCoursesBtn) {
            this.mobileViewCoursesBtn.addEventListener("click", () => {
                if (this.onNavigate) {
                    this.onNavigate("courses");
                } else {
                    window.location.href = "/";
                }
            });
        }

        if (this.mobileProfileBtn) {
            this.mobileProfileBtn.addEventListener("click", () => {
                window.location.href = "/profile/";
            });
        }

        if (this.mobileLogoutBtn) {
            this.mobileLogoutBtn.addEventListener("click", () => this.handleLogout());
        }

        if (this.mobileMyCoursesBtn) {
            this.mobileMyCoursesBtn.addEventListener("click", () => {
                if (this.onNavigate) {
                    this.onNavigate("myCourses");
                } else {
                    window.location.href = "/";
                }
            });
        }

        if (this.mobileAdminBtn) {
            this.mobileAdminBtn.addEventListener("click", () => {
                if (this.onNavigate) {
                    this.onNavigate("admin");
                } else {
                    window.location.href = "/";
                }
            });
        }

        if (this.hamburgerBtn) {
            this.hamburgerBtn.addEventListener("click", () => this.toggleMobileMenu());
        }

        if (this.mobileOverlay) {
            this.mobileOverlay.addEventListener("click", () => this.closeMobileMenu());
        }
    },

    initProfileDropdown() {
        if (this.profileTrigger) {
            this.profileTrigger.addEventListener("click", (e) => this.handleProfileClick(e));
        }

        if (this.profileViewBtn) {
            this.profileViewBtn.addEventListener("click", () => this.handleViewProfile());
        }

        if (this.profileLogoutBtn) {
            this.profileLogoutBtn.addEventListener("click", () => this.handleLogout());
        }
    },

    handleProfileClick(e) {
        e.preventDefault();
        e.stopPropagation();
        this.toggleDropdown();
    },

    toggleDropdown() {
        if (this.dropdownOpen) {
            this.closeDropdown();
        } else {
            this.openDropdown();
        }
    },

    openDropdown() {
        if (!this.profileDropdown) return;
        this.dropdownOpen = true;
        this.profileDropdown.classList.add("open");
        this.setupClickOutsideListener();
    },

    closeDropdown() {
        if (!this.profileDropdown) return;
        this.dropdownOpen = false;
        this.profileDropdown.classList.remove("open");
        this.removeClickOutsideListener();
    },

    setupClickOutsideListener() {
        if (this.clickOutsideHandler) return;
        this.clickOutsideHandler = (e) => {
            if (this.profileDropdown && !this.profileDropdown.contains(e.target)) {
                this.closeDropdown();
            }
        };
        document.addEventListener("click", this.clickOutsideHandler);
    },

    removeClickOutsideListener() {
        if (this.clickOutsideHandler) {
            document.removeEventListener("click", this.clickOutsideHandler);
            this.clickOutsideHandler = null;
        }
    },

    handleViewProfile() {
        this.closeDropdown();
        window.location.href = "/profile/";
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

        this.mainNav.style.display = "flex";

        if (isAuthenticated && user) {
            if (this.profileDropdown) {
                this.profileDropdown.style.display = "block";
                if (this.profileTriggerText) {
                    this.profileTriggerText.textContent = user.email;
                }
            }
            if (this.mobileUserInfo) {
                this.mobileUserInfo.textContent = `Logged in as ${user.email} (${user.role})`;
            }

            if (this.homeBtn) {
                this.homeBtn.style.display = "none";
            }
            if (this.viewCoursesBtn) {
                this.viewCoursesBtn.style.display = "inline-block";
            }
            if (this.profileBtn) {
                this.profileBtn.style.display = "none";
            }
            if (this.logoutBtn) {
                this.logoutBtn.style.display = "none";
            }

            if (this.mobileHomeBtn) {
                this.mobileHomeBtn.style.display = "none";
            }
            if (this.mobileViewCoursesBtn) {
                this.mobileViewCoursesBtn.style.display = "block";
            }
            if (this.mobileProfileBtn) {
                this.mobileProfileBtn.style.display = "block";
            }
            if (this.mobileLogoutBtn) {
                this.mobileLogoutBtn.style.display = "none";
            }

            if (user.role === "admin" && this.adminBtn) {
                this.adminBtn.style.display = "inline-block";
            }
            if (user.role === "admin" && this.mobileAdminBtn) {
                this.mobileAdminBtn.style.display = "block";
            } else {
                if (this.adminBtn) {
                    this.adminBtn.style.display = "none";
                }
                if (this.mobileAdminBtn) {
                    this.mobileAdminBtn.style.display = "none";
                }
            }

            if ((user.role === "instructor" || user.role === "admin") && this.myCoursesBtn) {
                this.myCoursesBtn.style.display = "inline-block";
            }
            if ((user.role === "instructor" || user.role === "admin") && this.mobileMyCoursesBtn) {
                this.mobileMyCoursesBtn.style.display = "block";
            } else {
                if (this.myCoursesBtn) {
                    this.myCoursesBtn.style.display = "none";
                }
                if (this.mobileMyCoursesBtn) {
                    this.mobileMyCoursesBtn.style.display = "none";
                }
            }
        } else {
            if (this.profileDropdown) {
                this.profileDropdown.style.display = "none";
            }
            if (this.mobileUserInfo) {
                this.mobileUserInfo.textContent = "";
            }

            if (this.homeBtn) {
                this.homeBtn.style.display = "inline-block";
            }
            if (this.viewCoursesBtn) {
                this.viewCoursesBtn.style.display = "none";
            }
            if (this.profileBtn) {
                this.profileBtn.style.display = "none";
            }
            if (this.logoutBtn) {
                this.logoutBtn.style.display = "none";
            }
            if (this.adminBtn) {
                this.adminBtn.style.display = "none";
            }
            if (this.myCoursesBtn) {
                this.myCoursesBtn.style.display = "none";
            }

            if (this.mobileHomeBtn) {
                this.mobileHomeBtn.style.display = "block";
            }
            if (this.mobileViewCoursesBtn) {
                this.mobileViewCoursesBtn.style.display = "none";
            }
            if (this.mobileProfileBtn) {
                this.mobileProfileBtn.style.display = "none";
            }
            if (this.mobileLogoutBtn) {
                this.mobileLogoutBtn.style.display = "none";
            }
            if (this.mobileAdminBtn) {
                this.mobileAdminBtn.style.display = "none";
            }
            if (this.mobileMyCoursesBtn) {
                this.mobileMyCoursesBtn.style.display = "none";
            }

            this.closeDropdown();
        }
    },

    handleLogout() {
        this.closeDropdown();
        sessionStorage.removeItem("authToken");
        this.currentUser = null;
        if (this.onNavigate) {
            this.onNavigate("home");
        } else {
            window.location.href = "/";
        }
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

document.addEventListener("DOMContentLoaded", () => {
    if (typeof app === "undefined" || !document.getElementById("homeView")) {
        navbarModule.init();
    }
});

