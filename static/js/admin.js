const adminModule = {
    async load() {
        await this.loadUsers();
        await this.loadPendingCourses();
    },
    
    async loadUsers() {
        try {
            const users = await api.admin.getUsers();
            this.renderUsers(users);
        } catch (error) {
            document.getElementById('usersList').innerHTML = '<p>Error loading users</p>';
        }
    },
    
    renderUsers(users) {
        const usersList = document.getElementById('usersList');
        
        if (!users || users.length === 0) {
            usersList.innerHTML = '<p>No users found</p>';
            return;
        }
        
        usersList.innerHTML = users.map(user => `
            <div class="user-card">
                <div class="user-info">
                    <div class="user-email">${escapeHtml(user.email)}</div>
                    <div class="user-role">Role: ${user.role}</div>
                </div>
                <div>ID: ${user.id}</div>
            </div>
        `).join('');
    },
    
    async loadPendingCourses() {
        try {
            const courses = await api.admin.getCourses('pending');
            this.renderPendingCourses(courses);
        } catch (error) {
            document.getElementById('pendingCoursesList').innerHTML = '<p>Error loading pending courses</p>';
        }
    },
    
    renderPendingCourses(courses) {
        const pendingCoursesList = document.getElementById('pendingCoursesList');
        
        if (!courses || courses.length === 0) {
            pendingCoursesList.innerHTML = '<p>No pending courses</p>';
            return;
        }
        
        pendingCoursesList.innerHTML = courses.map(course => `
            <div class="pending-course-card">
                <h4>${escapeHtml(course.title)}</h4>
                <p>${escapeHtml(course.description)}</p>
                <div>Instructor: ${escapeHtml(course.instructor_name || course.instructor_email)}</div>
                <div class="course-actions">
                    <button class="approve-btn" onclick="adminModule.approveCourse(${course.id})">Approve</button>
                    <button class="reject-btn" onclick="adminModule.rejectCourse(${course.id})">Reject</button>
                </div>
            </div>
        `).join('');
    },
    
    async approveCourse(id) {
        try {
            await api.admin.approveCourse(id);
            this.loadPendingCourses();
        } catch (error) {
            alert(error.message);
        }
    },
    
    async rejectCourse(id) {
        try {
            await api.admin.rejectCourse(id);
            this.loadPendingCourses();
        } catch (error) {
            alert(error.message);
        }
    }
};
