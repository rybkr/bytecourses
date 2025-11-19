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
		navbarModule.onNavigate = (view) => this.showView(view);
		navbarModule.init();
		authModule.validateToken();
	},

	initElements() {
		this.homeView = document.getElementById("homeView");
		this.aboutView = document.getElementById("aboutView");
		this.authView = document.getElementById("authView");
		this.coursesView = document.getElementById("coursesView");
		this.submitView = document.getElementById("submitView");
		this.myCoursesView = document.getElementById("myCoursesView");
		this.profileView = document.getElementById("profileView");
		this.adminView = document.getElementById("adminView");

		this.getStartedBtn = document.getElementById("getStartedBtn");
		this.aboutGetStartedBtn = document.getElementById("aboutGetStartedBtn");
	},

	initEventListeners() {
		this.getStartedBtn.addEventListener("click", () => this.showView("auth"));
		this.aboutGetStartedBtn.addEventListener("click", () => this.showView("auth"));
	},

	initModules() {
		authModule.init();
		coursesModule.init();
		profileModule.init();
		instructorModule.init();
		adminModule.init();
	},

	showView(view) {
		if (navbarModule.closeMobileMenu) {
			navbarModule.closeMobileMenu();
		}

		this.homeView.classList.remove("active");
		this.aboutView.classList.remove("active");
		this.authView.classList.remove("active");
		this.coursesView.classList.remove("active");
		this.myCoursesView.classList.remove("active");
		this.profileView.classList.remove("active");
		this.adminView.classList.remove("active");

		if (view === "home") {
			this.homeView.classList.add("active");
		} else if (view === "auth") {
			this.authView.classList.add("active");
		} else if (view === "courses") {
			this.coursesView.classList.add("active");
			coursesModule.load();
		} else if (view === "myCourses") {
			this.myCoursesView.classList.add("active");
			instructorModule.load();
		} else if (view === "profile") {
			this.profileView.classList.add("active");
			profileModule.load();
		} else if (view === "admin") {
			this.adminView.classList.add("active");
			adminModule.load();
		}
	},

	showAuthenticatedUI() {
		this.homeView.classList.remove("active");
		this.aboutView.classList.remove("active");
		this.authView.classList.remove("active");
		this.coursesView.classList.add("active");
		navbarModule.checkAuthState();
	},

	showUnauthenticatedUI() {
		this.homeView.classList.add("active");
		this.aboutView.classList.remove("active");
		this.authView.classList.remove("active");
		this.coursesView.classList.remove("active");
		this.submitView.classList.remove("active");
		this.myCoursesView.classList.remove("active");
		this.profileView.classList.remove("active");
		this.adminView.classList.remove("active");
		navbarModule.checkAuthState();
		const authForm = document.getElementById("authForm");
		if (authForm) {
			authForm.reset();
		}
	},

};

document.addEventListener("DOMContentLoaded", () => app.init());
