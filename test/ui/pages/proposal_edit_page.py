from playwright.sync_api import Page, expect
from .base_page import BasePage


class ProposalEditPage(BasePage):
    FORM = "#proposal-form"
    TITLE_INPUT = "#title"
    SUMMARY_INPUT = "#summary"
    QUALIFICATIONS_INPUT = "#qualifications"
    TARGET_AUDIENCE_INPUT = "#target_audience"
    LEARNING_OBJECTIVES_INPUT = "#learning_objectives"
    OUTLINE_INPUT = "#outline"
    PREREQUISITES_INPUT = "#assumed_prerequisites"
    SUBMIT_BTN = "#submitBtn"
    SAVE_DRAFT_BTN = "#saveDraftBtn"
    SAVE_STATUS = "#save-status"
    BACK_TO_PROPOSALS_LINK = "a[href='/proposals']"

    def __init__(self, page: Page, base_url: str):
        super().__init__(page, base_url)

    def navigate_new(self):
        self.goto("/proposals/new")

    def navigate_edit(self, proposal_id: int):
        self.goto(f"/proposals/{proposal_id}/edit")

    def fill_title(self, title: str):
        self.page.fill(self.TITLE_INPUT, title)

    def fill_summary(self, summary: str):
        self.page.fill(self.SUMMARY_INPUT, summary)

    def fill_qualifications(self, qualifications: str):
        self.page.fill(self.QUALIFICATIONS_INPUT, qualifications)

    def fill_target_audience(self, audience: str):
        self.page.fill(self.TARGET_AUDIENCE_INPUT, audience)

    def fill_learning_objectives(self, objectives: str):
        self.page.fill(self.LEARNING_OBJECTIVES_INPUT, objectives)

    def fill_outline(self, outline: str):
        self.page.fill(self.OUTLINE_INPUT, outline)

    def fill_prerequisites(self, prereqs: str):
        self.page.fill(self.PREREQUISITES_INPUT, prereqs)

    def fill_all_fields(
        self,
        title: str,
        summary: str,
        qualifications: str,
        target_audience: str,
        learning_objectives: str,
        outline: str,
        prerequisites: str,
    ):
        self.fill_title(title)
        self.fill_summary(summary)
        self.fill_qualifications(qualifications)
        self.fill_target_audience(target_audience)
        self.fill_learning_objectives(learning_objectives)
        self.fill_outline(outline)
        self.fill_prerequisites(prerequisites)

    def get_title(self) -> str:
        return self.page.locator(self.TITLE_INPUT).input_value()

    def get_summary(self) -> str:
        return self.page.locator(self.SUMMARY_INPUT).input_value()

    def click_submit(self):
        self.page.click(self.SUBMIT_BTN)

    def click_save_draft(self):
        self.page.click(self.SAVE_DRAFT_BTN)

    def get_save_status(self) -> str:
        return self.page.locator(self.SAVE_STATUS).text_content() or ""

    def wait_for_autosave(self, timeout: float = 10000):
        expect(self.page.locator(self.SAVE_STATUS)).to_contain_text(
            "Saved at", timeout=timeout
        )

    def trigger_autosave_and_wait(self, timeout: float = 10000):
        current_title = self.get_title()
        self.fill_title(current_title + " ")
        self.wait_for_autosave(timeout)
