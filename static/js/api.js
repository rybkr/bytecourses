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
				throw error;
			} else {
				const error = new Error(errorData.error || "Request failed");
				error.type = "simple";
				error.message = errorData.error || "Request failed";
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
		list: (status) => {
			const url = status ? `/courses?status=${status}` : "/courses";
			return request(url);
		},
		create: (data) =>
			request("/courses", {
				method: "POST",
				body: JSON.stringify(data),
			}),
	},

	instructor: {
		getCourses: () => request("/instructor/courses"),
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

	admin: {
		getUsers: () => request("/admin/users"),
		getCourses: (status) => request(`/admin/courses?status=${status}`),
		approveCourse: (id) =>
			request(`/admin/courses/approve?id=${id}`, {
				method: "PATCH",
			}),
		rejectCourse: (id) =>
			request(`/admin/courses/reject?id=${id}`, {
				method: "PATCH",
			}),
	},
};
