const instructorModule = {
    init() {
        const editCourseForm = document.getElementById('editCourseForm');
        const closeModal = document.getElementsByClassName('close')[0];
        const editModal = document.getElementById('editModal');
        
        editCourseForm.addEventListener('submit', this.handleEdit.bind(this));
        
        closeModal.onclick = () => editModal.style.display = 'none';
        window.onclick = (event) => {
            if (event.target == editModal) {
                editModal.style.display = 'none';
            }
        };
    },
    
    async load() {
        try {
            const courses = await api.instructor.getCourses();
            this.render(courses);
        } catch (error) {
            document.getElementById('myCoursesList').innerHTML = '<p>Error loading your courses</p>';
        }
    },
    
    render(courses) {
        const myCoursesList = document.getElementById('myCoursesList');
        
        if (!courses || courses.length === 0) {
            myCoursesList.innerHTML = '<p>You haven\'t created any courses yet</p>';
            return;
        }
        
        myCoursesList.innerHTML = courses.map(course => `
            <div class="my-course-card">
                <h3>${escapeHtml(course.title)}</h3>
                <p>${escapeHtml(course.description)}</p>
                <div class="my-course-meta">
                    <span class="status-badge status-${course.status}">${course.status}</span>
                    <div class="my-course-actions">
                        <button class="edit-btn" onclick="instructorModule.openEditModal(${course.id}, '${escapeHtml(course.title).replace(/'/g, "\\'")}', '${escapeHtml(course.description).replace(/'/g, "\\'")}')">Edit</button>
                        <button class="delete-btn-small" onclick="instructorModule.deleteCourse(${course.id})">Delete</button>
                    </div>
                </div>
            </div>
        `).join('');
    },
    
    openEditModal(id, title, description) {
        document.getElementById('editCourseId').value = id;
        document.getElementById('editTitle').value = title;
        document.getElementById('editDescription').value = description;
        document.getElementById('editModal').style.display = 'block';
    },
    
    async handleEdit(e) {
        e.preventDefault();
        
        const courseId = document.getElementById('editCourseId').value;
        const formData = {
            title: document.getElementById('editTitle').value,
            description: document.getElementById('editDescription').value
        };
        
        try {
            await api.instructor.updateCourse(courseId, formData);
            document.getElementById('editModal').style.display = 'none';
            this.showMessage('Course updated successfully!', 'success');
            this.load();
        } catch (error) {
            alert(error.message);
        }
    },
    
    async deleteCourse(id) {
        if (!confirm('Are you sure you want to delete this course?')) {
            return;
        }
        
        try {
            await api.instructor.deleteCourse(id);
            this.showMessage('Course deleted successfully', 'success');
            this.load();
        } catch (error) {
            alert(error.message);
        }
    },
    
    showMessage(message, type) {
        const myCoursesMessage = document.getElementById('myCoursesMessage');
        myCoursesMessage.textContent = message;
        myCoursesMessage.className = type;
        setTimeout(() => {
            myCoursesMessage.style.display = 'none';
        }, 3000);
    }
};
