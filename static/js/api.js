const API_BASE = "/api";

async function request(endpoint, options = {}) {
	const token = sessionStorage.getItem("authToken");

	const config = {
		headers: {
			"Content-Type": "application/json",
			...(token && { Authorization: `Bearer ${token}` }),
			...options.headers,
		},
		...options,
	};

	try {
		const response = await fetch(`${API_BASE}${endpoint}`, config);

		if (!response.ok) {
			const errorData = await response
				.json()
				.catch(() => ({ error: "Request failed" }));

			if (errorData.fields && errorData.error === "validation_failed") {
				const error = new Error(errorData.error);
				error.type = "validation";
				error.fields = errorData.fields;
				error.message = "validation_failed";
				error.status = response.status;
				throw error;
			} else {
				const error = new Error(errorData.error || "Request failed");
				error.type = "simple";
				error.message = errorData.error || "Request failed";
				error.status = response.status;
				throw error;
			}
		}

		if (response.status === 204) {
			return null;
		}

		return await response.json();
	} catch (error) {
		console.error("API request failed:", error);
		throw error;
	}
}

const api = {
	auth: {
		signup: (data) =>
			request("/auth/signup", {
				method: "POST",
				body: JSON.stringify(data),
			}),
		login: (data) =>
			request("/auth/login", {
				method: "POST",
				body: JSON.stringify(data),
			}),
	},

	profile: {
		get: () => request("/profile"),
		update: (data) =>
			request("/profile", {
				method: "PATCH",
				body: JSON.stringify(data),
			}),
	},

	courses: {
		list: () => request("/courses"),
		get: (id) => request(`/courses/${id}`),
	},

	applications: {
		create: (data) =>
			request("/applications", {
				method: "POST",
				body: JSON.stringify(data),
			}),
		list: () => request("/instructor/applications"),
		update: (id, data) =>
			request(`/instructor/applications?id=${id}`, {
				method: "PATCH",
				body: JSON.stringify(data),
			}),
		delete: (id) =>
			request(`/instructor/applications?id=${id}`, {
				method: "DELETE",
			}),
		submit: (id, data) =>
			request(`/instructor/applications/submit?id=${id}`, {
				method: "PATCH",
				body: JSON.stringify(data),
			}),
	},

	instructor: {
		getApplications: () => request("/instructor/applications"),
		getCourses: () => request("/instructor/courses"),
		updateApplication: (id, data) =>
			request(`/instructor/applications?id=${id}`, {
				method: "PATCH",
				body: JSON.stringify(data),
			}),
		updateCourse: (id, data) =>
			request(`/instructor/courses?id=${id}`, {
				method: "PATCH",
				body: JSON.stringify(data),
			}),
		deleteCourse: (id) =>
			request(`/instructor/courses?id=${id}`, {
				method: "DELETE",
			}),
	},

	drafts: {
		create: (data) =>
			request("/applications", {
				method: "POST",
				body: JSON.stringify({ ...data, status: "draft" }),
			}),
		update: (id, data) =>
			request(`/instructor/applications?id=${id}`, {
				method: "PATCH",
				body: JSON.stringify({ ...data, status: "draft" }),
			}),
		submit: (id, data) =>
			request(`/instructor/applications/submit?id=${id}`, {
				method: "PATCH",
				body: JSON.stringify(data),
			}),
	},

	admin: {
		getUsers: () => request("/admin/users"),
		getApplications: () => request("/admin/applications"),
		approveApplication: (id) =>
			request(`/admin/applications/approve?id=${id}`, {
				method: "PATCH",
			}),
		rejectApplication: (id) =>
			request(`/admin/applications/reject?id=${id}`, {
				method: "PATCH",
			}),
	},
};
