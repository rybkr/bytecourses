const API_BASE = '/api';

let authToken = localStorage.getItem('authToken');
let currentUser = null;

const authView = document.getElementById('authView');
const coursesView = document.getElementById('coursesView');
const submitView = document.getElementById('submitView');
const mainNav = document.getElementById('mainNav');
const userInfo = document.getElementById('userInfo');

const authForm = document.getElementById('authForm');
const authTitle = document.getElementById('authTitle');
const authSubmitBtn = document.getElementById('authSubmitBtn');
const authToggleBtn = document.getElementById('authToggleBtn');
const authToggleText = document.getElementById('authToggleText');
const roleGroup = document.getElementById('roleGroup');
const authMessage = document.getElementById('authMessage');

const viewCoursesBtn = document.getElementById('viewCoursesBtn');
const submitCourseBtn = document.getElementById('submitCourseBtn');
const logoutBtn = document.getElementById('logoutBtn');
const coursesList = document.getElementById('coursesList');
const courseForm = document.getElementById('courseForm');
const statusFilter = document.getElementById('statusFilter');
const formMessage = document.getElementById('formMessage');

let isSignupMode = false;

if (authToken) {
    fetchCurrentUser();
}

authToggleBtn.addEventListener('click', () => {
    isSignupMode = !isSignupMode;
    if (isSignupMode) {
        authTitle.textContent = 'Sign Up';
        authSubmitBtn.textContent = 'Sign Up';
        authToggleText.textContent = 'Already have an account?';
        authToggleBtn.textContent = 'Login';
        roleGroup.style.display = 'block';
    } else {
        authTitle.textContent = 'Login';
        authSubmitBtn.textContent = 'Login';
        authToggleText.textContent = "Don't have an account?";
        authToggleBtn.textContent = 'Sign up';
        roleGroup.style.display = 'none';
    }
    authMessage.style.display = 'none';
});

authForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const email = document.getElementById('authEmail').value;
    const password = document.getElementById('authPassword').value;
    const role = document.getElementById('authRole').value;
    
    const endpoint = isSignupMode ? '/auth/signup' : '/auth/login';
    const body = isSignupMode 
        ? { email, password, role }
        : { email, password };
    
    try {
        const response = await fetch(`${API_BASE}${endpoint}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(body)
        });
        
        if (response.ok) {
            const data = await response.json();
            authToken = data.token;
            currentUser = data.user;
            localStorage.setItem('authToken', authToken);
            showAuthenticatedUI();
            loadCourses();
        } else {
            const error = await response.text();
            showAuthMessage(error || 'Authentication failed', 'error');
        }
    } catch (error) {
        showAuthMessage('Error: ' + error.message, 'error');
    }
});

logoutBtn.addEventListener('click', () => {
    authToken = null;
    currentUser = null;
    localStorage.removeItem('authToken');
    showUnauthenticatedUI();
});

viewCoursesBtn.addEventListener('click', () => {
    showView('courses');
    loadCourses();
});

submitCourseBtn.addEventListener('click', () => {
    showView('submit');
});

statusFilter.addEventListener('change', loadCourses);

courseForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const formData = {
        title: document.getElementById('title').value,
        description: document.getElementById('description').value
    };
    
    try {
        const response = await fetch(`${API_BASE}/courses`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify(formData)
        });
        
        if (response.ok) {
            showMessage('Course submitted successfully! Awaiting approval.', 'success');
            courseForm.reset();
        } else {
            showMessage('Failed to submit course. Please try again.', 'error');
        }
    } catch (error) {
        showMessage('Error: ' + error.message, 'error');
    }
});

async function fetchCurrentUser() {
    try {
        const response = await fetch(`${API_BASE}/courses`, {
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        if (response.ok) {
            showAuthenticatedUI();
            loadCourses();
        } else {
            authToken = null;
            localStorage.removeItem('authToken');
            showUnauthenticatedUI();
        }
    } catch (error) {
        console.error('Error validating token:', error);
        showUnauthenticatedUI();
    }
}

function showAuthenticatedUI() {
    authView.classList.remove('active');
    coursesView.classList.add('active');
    mainNav.style.display = 'block';
    userInfo.style.display = 'block';
    
    if (currentUser) {
        userInfo.textContent = `Logged in as ${currentUser.email} (${currentUser.role})`;
    }
}

function showUnauthenticatedUI() {
    authView.classList.add('active');
    coursesView.classList.remove('active');
    submitView.classList.remove('active');
    mainNav.style.display = 'none';
    userInfo.style.display = 'none';
    authForm.reset();
}

function showView(view) {
    coursesView.classList.remove('active');
    submitView.classList.remove('active');
    viewCoursesBtn.classList.remove('active');
    submitCourseBtn.classList.remove('active');
    
    if (view === 'courses') {
        coursesView.classList.add('active');
        viewCoursesBtn.classList.add('active');
    } else {
        submitView.classList.add('active');
        submitCourseBtn.classList.add('active');
    }
}

async function loadCourses() {
    const status = statusFilter.value;
    const url = status ? `${API_BASE}/courses?status=${status}` : `${API_BASE}/courses`;
    
    try {
        const response = await fetch(url);
        const courses = await response.json();
        renderCourses(courses);
    } catch (error) {
        coursesList.innerHTML = '<p>Error loading courses</p>';
    }
}

function renderCourses(courses) {
    if (!courses || courses.length === 0) {
        coursesList.innerHTML = '<p>No courses found</p>';
        return;
    }
    
    coursesList.innerHTML = courses.map(course => `
        <div class="course-card">
            <h3>${escapeHtml(course.title)}</h3>
            <p>${escapeHtml(course.description)}</p>
            <div class="course-meta">
                <span>Instructor ID: ${course.instructor_id}</span>
                <span class="status-badge status-${course.status}">${course.status}</span>
            </div>
        </div>
    `).join('');
}

function showMessage(message, type) {
    formMessage.textContent = message;
    formMessage.className = type;
    setTimeout(() => {
        formMessage.style.display = 'none';
    }, 5000);
}

function showAuthMessage(message, type) {
    authMessage.textContent = message;
    authMessage.className = type;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
