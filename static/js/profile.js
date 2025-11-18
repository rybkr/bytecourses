const profileModule = {
	init() {
		const profileForm = document.getElementById("profileForm");
		profileForm.addEventListener("submit", this.handleSubmit.bind(this));
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
		profileMessage.textContent = message;
		profileMessage.className = type;
		setTimeout(() => {
			profileMessage.style.display = "none";
		}, 3000);
	},
};
