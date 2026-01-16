from playwright.sync_api import Page
from .base_page import BasePage


class ProposalsPage(BasePage):
    PATH = "/proposals"

    PAGE_HEADER = ".page-header h1"
    NEW_PROPOSAL_BTN = "a[href='/proposals/new']"
    PROPOSALS_LIST = "#proposals-list"
    LOADING_INDICATOR = ".loading"
    PROPOSAL_CARD = ".proposal-card"
    PROPOSAL_TITLE = ".proposal-card h3"
    PROPOSAL_STATUS = ".proposal-status"

    def __init__(self, page: Page, base_url: str):
        super().__init__(page, base_url)

    def navigate(self):
        self.goto(self.PATH)

    def wait_for_proposals_loaded(self):
        self.page.wait_for_selector(self.LOADING_INDICATOR, state="hidden")

    def get_page_title(self) -> str:
        return self.page.locator(self.PAGE_HEADER).text_content() or ""

    def click_new_proposal(self):
        self.page.click(self.NEW_PROPOSAL_BTN)

    def get_proposal_count(self) -> int:
        self.wait_for_proposals_loaded()
        return self.page.locator(self.PROPOSAL_CARD).count()

    def get_proposal_titles(self) -> list[str]:
        self.wait_for_proposals_loaded()
        titles = self.page.locator(self.PROPOSAL_TITLE).all_text_contents()
        return titles

    def click_proposal_by_title(self, title: str):
        self.wait_for_proposals_loaded()
        self.page.locator(self.PROPOSAL_CARD, has_text=title).click()
