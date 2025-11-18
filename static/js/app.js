function escapeHtml(text) {
	const div = document.createElement("div");
	div.textContent = text;
	return div.innerHTML;
}

const app = {
	currentUser: null,

	init() {
		this.initElements();
		this.initEventListeners();
		this.initModules();
		authModule.validateToken();
	},

	initElements() {
		this.authView = document.getElementById("authView");
		this.coursesView = document.getElementById("coursesView");
		this.submitView = document.getElementById("submitView");
		this.myCoursesView = document.getElementById("myCoursesView");
		this.profileView = document.getElementById("profileView");
		this.adminView = document.getElementById("adminView");
		this.mainNav = document.getElementById("mainNav");
		this.userInfo = document.getElementById("userInfo");
		this.mobileUserInfo = document.getElementById("mobileUserInfo");
		this.hamburgerBtn = document.getElementById("hamburgerBtn");
		this.mobileNav = document.getElementById("mobileNav");
		this.mobileOverlay = document.getElementById("mobileOverlay");

		this.viewCoursesBtn = document.getElementById("viewCoursesBtn");
		this.submitCourseBtn = document.getElementById("submitCourseBtn");
		this.myCoursesBtn = document.getElementById("myCoursesBtn");
		this.profileBtn = document.getElementById("profileBtn");
		this.adminBtn = document.getElementById("adminBtn");
		this.logoutBtn = document.getElementById("logoutBtn");

		this.mobileViewCoursesBtn = document.getElementById("mobileViewCoursesBtn");
		this.mobileSubmitCourseBtn = document.getElementById("mobileSubmitCourseBtn");
		this.mobileMyCoursesBtn = document.getElementById("mobileMyCoursesBtn");
		this.mobileProfileBtn = document.getElementById("mobileProfileBtn");
		this.mobileAdminBtn = document.getElementById("mobileAdminBtn");
		this.mobileLogoutBtn = document.getElementById("mobileLogoutBtn");
	},

	initEventListeners() {
		this.viewCoursesBtn.addEventListener("click", () =>
			this.showView("courses"),
		);
		this.submitCourseBtn.addEventListener("click", () =>
			this.showView("submit"),
		);
		this.myCoursesBtn.addEventListener("click", () =>
			this.showView("myCourses"),
		);
		this.profileBtn.addEventListener("click", () => this.showView("profile"));
		this.adminBtn.addEventListener("click", () => this.showView("admin"));
		this.logoutBtn.addEventListener("click", () => authModule.logout());

		this.mobileViewCoursesBtn.addEventListener("click", () =>
			this.showView("courses"),
		);
		this.mobileSubmitCourseBtn.addEventListener("click", () =>
			this.showView("submit"),
		);
		this.mobileMyCoursesBtn.addEventListener("click", () =>
			this.showView("myCourses"),
		);
		this.mobileProfileBtn.addEventListener("click", () =>
			this.showView("profile"),
		);
		this.mobileAdminBtn.addEventListener("click", () => this.showView("admin"));
		this.mobileLogoutBtn.addEventListener("click", () => authModule.logout());

		this.hamburgerBtn.addEventListener("click", () => this.toggleMobileMenu());
		this.mobileOverlay.addEventListener("click", () => this.closeMobileMenu());
	},

	initModules() {
		authModule.init();
		coursesModule.init();
		profileModule.init();
		instructorModule.init();
		adminModule.init();
	},

	showView(view) {
		this.closeMobileMenu();

		this.coursesView.classList.remove("active");
		this.submitView.classList.remove("active");
		this.myCoursesView.classList.remove("active");
		this.profileView.classList.remove("active");
		this.adminView.classList.remove("active");
		this.viewCoursesBtn.classList.remove("active");
		this.submitCourseBtn.classList.remove("active");
		this.myCoursesBtn.classList.remove("active");
		this.profileBtn.classList.remove("active");
		this.adminBtn.classList.remove("active");
		this.mobileViewCoursesBtn.classList.remove("active");
		this.mobileSubmitCourseBtn.classList.remove("active");
		this.mobileMyCoursesBtn.classList.remove("active");
		this.mobileProfileBtn.classList.remove("active");
		this.mobileAdminBtn.classList.remove("active");

		if (view === "courses") {
			this.coursesView.classList.add("active");
			this.viewCoursesBtn.classList.add("active");
			this.mobileViewCoursesBtn.classList.add("active");
			coursesModule.load();
		} else if (view === "submit") {
			this.submitView.classList.add("active");
			this.submitCourseBtn.classList.add("active");
			this.mobileSubmitCourseBtn.classList.add("active");
		} else if (view === "myCourses") {
			this.myCoursesView.classList.add("active");
			this.myCoursesBtn.classList.add("active");
			this.mobileMyCoursesBtn.classList.add("active");
			instructorModule.load();
		} else if (view === "profile") {
			this.profileView.classList.add("active");
			this.profileBtn.classList.add("active");
			this.mobileProfileBtn.classList.add("active");
			profileModule.load();
		} else if (view === "admin") {
			this.adminView.classList.add("active");
			this.adminBtn.classList.add("active");
			this.mobileAdminBtn.classList.add("active");
			adminModule.load();
		}
	},

	showAuthenticatedUI() {
		this.authView.classList.remove("active");
		this.coursesView.classList.add("active");
		this.mainNav.style.display = "flex";
		this.userInfo.style.display = "block";

		if (this.currentUser) {
			const userText = `Logged in as ${this.currentUser.email} (${this.currentUser.role})`;
			this.userInfo.textContent = userText;
			this.mobileUserInfo.textContent = userText;

			if (this.currentUser.role === "admin") {
				this.adminBtn.style.display = "inline-block";
				this.mobileAdminBtn.style.display = "block";
			}

			if (
				this.currentUser.role === "instructor" ||
				this.currentUser.role === "admin"
			) {
				this.myCoursesBtn.style.display = "inline-block";
				this.mobileMyCoursesBtn.style.display = "block";
			}
		}
	},

	showUnauthenticatedUI() {
		this.authView.classList.add("active");
		this.coursesView.classList.remove("active");
		this.submitView.classList.remove("active");
		this.myCoursesView.classList.remove("active");
		this.profileView.classList.remove("active");
		this.adminView.classList.remove("active");
		this.mainNav.style.display = "none";
		this.userInfo.style.display = "none";
		this.adminBtn.style.display = "none";
		this.myCoursesBtn.style.display = "none";
		this.mobileAdminBtn.style.display = "none";
		this.mobileMyCoursesBtn.style.display = "none";
		this.closeMobileMenu();
		document.getElementById("authForm").reset();
	},

	toggleMobileMenu() {
		const isOpen = this.mobileNav.classList.contains("open");
		if (isOpen) {
			this.closeMobileMenu();
		} else {
			this.openMobileMenu();
		}
	},

	openMobileMenu() {
		this.mobileNav.classList.add("open");
		this.mobileOverlay.classList.add("open");
		this.hamburgerBtn.classList.add("open");
		document.body.style.overflow = "hidden";
	},

	closeMobileMenu() {
		this.mobileNav.classList.remove("open");
		this.mobileOverlay.classList.remove("open");
		this.hamburgerBtn.classList.remove("open");
		document.body.style.overflow = "";
	},
};

document.addEventListener("DOMContentLoaded", () => app.init());
