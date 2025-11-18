function escapeHtml(text) {
    const div = document.createElement('div');
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
        this.authView = document.getElementById('authView');
        this.coursesView = document.getElementById('coursesView');
        this.submitView = document.getElementById('submitView');
        this.myCoursesView = document.getElementById('myCoursesView');
        this.profileView = document.getElementById('profileView');
        this.adminView = document.getElementById('adminView');
        this.mainNav = document.getElementById('mainNav');
        this.userInfo = document.getElementById('userInfo');
        
        this.viewCoursesBtn = document.getElementById('viewCoursesBtn');
        this.submitCourseBtn = document.getElementById('submitCourseBtn');
        this.myCoursesBtn = document.getElementById('myCoursesBtn');
        this.profileBtn = document.getElementById('profileBtn');
        this.adminBtn = document.getElementById('adminBtn');
        this.logoutBtn = document.getElementById('logoutBtn');
    },
    
    initEventListeners() {
        this.viewCoursesBtn.addEventListener('click', () => this.showView('courses'));
        this.submitCourseBtn.addEventListener('click', () => this.showView('submit'));
        this.myCoursesBtn.addEventListener('click', () => this.showView('myCourses'));
        this.profileBtn.addEventListener('click', () => this.showView('profile'));
        this.adminBtn.addEventListener('click', () => this.showView('admin'));
        this.logoutBtn.addEventListener('click', () => authModule.logout());
    },
    
    initModules() {
        authModule.init();
        coursesModule.init();
        profileModule.init();
        instructorModule.init();
    },
    
    showView(view) {
        this.coursesView.classList.remove('active');
        this.submitView.classList.remove('active');
        this.myCoursesView.classList.remove('active');
        this.profileView.classList.remove('active');
        this.adminView.classList.remove('active');
        this.viewCoursesBtn.classList.remove('active');
        this.submitCourseBtn.classList.remove('active');
        this.myCoursesBtn.classList.remove('active');
        this.profileBtn.classList.remove('active');
        this.adminBtn.classList.remove('active');
        
        if (view === 'courses') {
            this.coursesView.classList.add('active');
            this.viewCoursesBtn.classList.add('active');
            coursesModule.load();
        } else if (view === 'submit') {
            this.submitView.classList.add('active');
            this.submitCourseBtn.classList.add('active');
        } else if (view === 'myCourses') {
            this.myCoursesView.classList.add('active');
            this.myCoursesBtn.classList.add('active');
            instructorModule.load();
        } else if (view === 'profile') {
            this.profileView.classList.add('active');
            this.profileBtn.classList.add('active');
            profileModule.load();
        } else if (view === 'admin') {
            this.adminView.classList.add('active');
            this.adminBtn.classList.add('active');
            adminModule.load();
        }
    },
    
    showAuthenticatedUI() {
        this.authView.classList.remove('active');
        this.coursesView.classList.add('active');
        this.mainNav.style.display = 'block';
        this.userInfo.style.display = 'block';
        
        if (this.currentUser) {
            this.userInfo.textContent = `Logged in as ${this.currentUser.email} (${this.currentUser.role})`;
            
            if (this.currentUser.role === 'admin') {
                this.adminBtn.style.display = 'inline-block';
            }
            
            if (this.currentUser.role === 'instructor' || this.currentUser.role === 'admin') {
                this.myCoursesBtn.style.display = 'inline-block';
            }
        }
    },
    
    showUnauthenticatedUI() {
        this.authView.classList.add('active');
        this.coursesView.classList.remove('active');
        this.submitView.classList.remove('active');
        this.myCoursesView.classList.remove('active');
        this.profileView.classList.remove('active');
        this.adminView.classList.remove('active');
        this.mainNav.style.display = 'none';
        this.userInfo.style.display = 'none';
        this.adminBtn.style.display = 'none';
        this.myCoursesBtn.style.display = 'none';
        document.getElementById('authForm').reset();
    }
};

document.addEventListener('DOMContentLoaded', () => app.init());
