const authModule = {
    init() {
        const authForm = document.getElementById('authForm');
        const authToggleBtn = document.getElementById('authToggleBtn');
        
        authToggleBtn.addEventListener('click', this.toggleAuthMode.bind(this));
        authForm.addEventListener('submit', this.handleSubmit.bind(this));
        
        this.isSignupMode = false;
    },
    
    toggleAuthMode() {
        this.isSignupMode = !this.isSignupMode;
        
        const authTitle = document.getElementById('authTitle');
        const authSubmitBtn = document.getElementById('authSubmitBtn');
        const authToggleText = document.getElementById('authToggleText');
        const authToggleBtn = document.getElementById('authToggleBtn');
        const roleGroup = document.getElementById('roleGroup');
        const authMessage = document.getElementById('authMessage');
        
        if (this.isSignupMode) {
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
    },
    
    async handleSubmit(e) {
        e.preventDefault();
        
        const email = document.getElementById('authEmail').value;
        const password = document.getElementById('authPassword').value;
        const role = document.getElementById('authRole').value;
        
        try {
            const data = this.isSignupMode 
                ? await api.auth.signup({ email, password, role })
                : await api.auth.login({ email, password });
            
            sessionStorage.setItem('authToken', data.token);
            app.currentUser = data.user;
            app.showAuthenticatedUI();
            coursesModule.load();
        } catch (error) {
            this.showMessage(error.message, 'error');
        }
    },
    
    showMessage(message, type) {
        const authMessage = document.getElementById('authMessage');
        authMessage.textContent = message;
        authMessage.className = type;
    },
    
    logout() {
        sessionStorage.removeItem('authToken');
        app.currentUser = null;
        app.showUnauthenticatedUI();
    },
    
    parseJWT(token) {
        try {
            const base64Url = token.split('.')[1];
            const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
            const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
                return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
            }).join(''));
            return JSON.parse(jsonPayload);
        } catch (e) {
            return null;
        }
    },
    
    async validateToken() {
        const token = sessionStorage.getItem('authToken');
        if (!token) {
            app.showUnauthenticatedUI();
            return;
        }
        
        const payload = this.parseJWT(token);
        if (!payload) {
            sessionStorage.removeItem('authToken');
            app.showUnauthenticatedUI();
            return;
        }
        
        if (payload.exp && payload.exp * 1000 < Date.now()) {
            sessionStorage.removeItem('authToken');
            app.showUnauthenticatedUI();
            return;
        }
        
        app.currentUser = {
            id: payload.user_id,
            email: payload.email,
            role: payload.role
        };
        
        app.showAuthenticatedUI();
        coursesModule.load();
    }
};
