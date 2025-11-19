const profileModule = {
	init() {
		const profileForm = document.getElementById("profileForm");
		if (profileForm) {
			profileForm.addEventListener("submit", this.handleSubmit.bind(this));
		}
	},

	initStandalone() {
		const token = sessionStorage.getItem("authToken");
		if (!token) {
			this.showMessage("You must be logged in to view your profile.", "error");
			setTimeout(() => {
				window.location.href = "/";
			}, 2000);
			return;
		}

		this.init();
		this.load();
	},

	async load() {
		try {
			const profile = await api.profile.get();
			document.getElementById("profileName").value = profile.name || "";
			document.getElementById("profileBio").value = profile.bio || "";
			document.getElementById("profileEmail").value = profile.email;
			document.getElementById("profileRole").value = profile.role;
			document.getElementById("profileCreatedAt").value = new Date(
				profile.created_at,
			).toLocaleDateString();
			document.getElementById("profileUpdatedAt").value = new Date(
				profile.updated_at,
			).toLocaleDateString();
		} catch (error) {
			console.error("Error loading profile:", error);
			if (error.status === 401 || error.status === 403) {
				this.showMessage("You must be logged in to view your profile.", "error");
				setTimeout(() => {
					window.location.href = "/";
				}, 2000);
			} else {
				this.showMessage("Failed to load profile. Please try again.", "error");
			}
		}
	},

	async handleSubmit(e) {
		e.preventDefault();

		const formData = {
			name: document.getElementById("profileName").value,
			bio: document.getElementById("profileBio").value,
		};

		try {
			await api.profile.update(formData);
			this.showMessage("Profile updated successfully!", "success");
			this.load();
		} catch (error) {
			if (error.type === "validation" && error.fields) {
				const fieldMessages = Object.entries(error.fields)
					.map(([field, msg]) => `${field}: ${msg}`)
					.join(", ");
				this.showMessage(fieldMessages, "error");
			} else {
				this.showMessage(error.message, "error");
			}
		}
	},

	showMessage(message, type) {
		const profileMessage = document.getElementById("profileMessage");
		if (profileMessage) {
			profileMessage.textContent = message;
			profileMessage.className = type;
			profileMessage.style.display = "block";
			if (type !== "error") {
				setTimeout(() => {
					profileMessage.style.display = "none";
				}, 3000);
			}
		}
	},
};
