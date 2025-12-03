const contentEditorModule = {
	sections: [],

	init(initialContent) {
		this.sections = [];
		if (initialContent) {
			try {
				const parsed = JSON.parse(initialContent);
				if (parsed.sections && Array.isArray(parsed.sections)) {
					this.sections = parsed.sections.map((section, index) => ({
						id: `section-${Date.now()}-${index}`,
						title: section.title || "",
						type: section.type || "lesson",
						content: section.content || "",
						order: section.order !== undefined ? section.order : index + 1,
					}));
				}
			} catch (error) {
				console.error("Error parsing initial content:", error);
			}
		}
		this.render();
	},

	render() {
		const container = document.getElementById("sectionsContainer");
		if (!container) return;

		if (this.sections.length === 0) {
			container.innerHTML = `
				<div class="empty-sections">
					<p>No sections yet. Click "Add Section" to get started.</p>
				</div>
			`;
			return;
		}

		const sortedSections = [...this.sections].sort((a, b) => (a.order || 0) - (b.order || 0));

		container.innerHTML = sortedSections
			.map(
				(section) => `
			<div class="section-editor-item" data-section-id="${section.id}">
				<div class="section-editor-header">
					<div class="section-editor-drag-handle">
						<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="9" cy="12" r="1"></circle>
							<circle cx="9" cy="5" r="1"></circle>
							<circle cx="9" cy="19" r="1"></circle>
							<circle cx="15" cy="12" r="1"></circle>
							<circle cx="15" cy="5" r="1"></circle>
							<circle cx="15" cy="19" r="1"></circle>
						</svg>
					</div>
					<input type="text" class="section-title-input" value="${this.escapeHtml(section.title)}" 
						placeholder="Section Title" data-section-id="${section.id}">
					<select class="section-type-select" data-section-id="${section.id}">
						<option value="lesson" ${section.type === "lesson" ? "selected" : ""}>Lesson</option>
						<option value="material" ${section.type === "material" ? "selected" : ""}>Material</option>
						<option value="assignment" ${section.type === "assignment" ? "selected" : ""}>Assignment</option>
					</select>
					<button type="button" class="btn-remove-section" data-section-id="${section.id}" title="Remove Section">
						<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="18" y1="6" x2="6" y2="18"></line>
							<line x1="6" y1="6" x2="18" y2="18"></line>
						</svg>
					</button>
				</div>
				<div class="section-editor-content">
					<textarea class="section-content-textarea" rows="6" 
						placeholder="Enter section content here..." data-section-id="${section.id}">${this.escapeHtml(section.content)}</textarea>
				</div>
			</div>
		`,
			)
			.join("");

		this.attachEventListeners();
		this.updateOrderNumbers();
	},

	attachEventListeners() {
		document.querySelectorAll(".section-title-input").forEach((input) => {
			input.addEventListener("input", (e) => {
				const sectionId = e.target.dataset.sectionId;
				const section = this.sections.find((s) => s.id === sectionId);
				if (section) {
					section.title = e.target.value;
				}
			});
		});

		document.querySelectorAll(".section-type-select").forEach((select) => {
			select.addEventListener("change", (e) => {
				const sectionId = e.target.dataset.sectionId;
				const section = this.sections.find((s) => s.id === sectionId);
				if (section) {
					section.type = e.target.value;
				}
			});
		});

		document.querySelectorAll(".section-content-textarea").forEach((textarea) => {
			textarea.addEventListener("input", (e) => {
				const sectionId = e.target.dataset.sectionId;
				const section = this.sections.find((s) => s.id === sectionId);
				if (section) {
					section.content = e.target.value;
				}
			});
		});

		document.querySelectorAll(".btn-remove-section").forEach((btn) => {
			btn.addEventListener("click", (e) => {
				const sectionId = e.target.closest(".btn-remove-section").dataset.sectionId;
				this.removeSection(sectionId);
			});
		});
	},

	addSection() {
		const newOrder = this.sections.length > 0 ? Math.max(...this.sections.map((s) => s.order || 0)) + 1 : 1;
		const newSection = {
			id: `section-${Date.now()}`,
			title: "",
			type: "lesson",
			content: "",
			order: newOrder,
		};
		this.sections.push(newSection);
		this.render();
		this.scrollToSection(newSection.id);
	},

	removeSection(sectionId) {
		if (!confirm("Are you sure you want to remove this section?")) {
			return;
		}
		this.sections = this.sections.filter((s) => s.id !== sectionId);
		this.updateOrderNumbers();
		this.render();
	},

	updateOrderNumbers() {
		const sortedSections = [...this.sections].sort((a, b) => (a.order || 0) - (b.order || 0));
		sortedSections.forEach((section, index) => {
			section.order = index + 1;
		});
	},

	scrollToSection(sectionId) {
		setTimeout(() => {
			const element = document.querySelector(`[data-section-id="${sectionId}"]`);
			if (element) {
				element.scrollIntoView({ behavior: "smooth", block: "nearest" });
				const input = element.querySelector(".section-title-input");
				if (input) {
					input.focus();
				}
			}
		}, 100);
	},

	getJSON() {
		this.updateOrderNumbers();
		return JSON.stringify(
			{
				sections: this.sections.map((s) => ({
					title: s.title,
					type: s.type,
					content: s.content,
					order: s.order,
				})),
			},
			null,
			2,
		);
	},

	escapeHtml(text) {
		const div = document.createElement("div");
		div.textContent = text;
		return div.innerHTML;
	},
};

