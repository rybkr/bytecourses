from playwright.sync_api import Page


class BasePage:
    def __init__(self, page: Page, base_url: str):
        self.page = page
        self.base_url = base_url

    NAV_BRAND = ".nav-brand a"
    USER_MENU_BTN = ".user-menu-btn"
    USER_DROPDOWN_MENU = "#userDropdownMenu"
    LOGOUT_BTN = ".user-dropdown-logout"
    LOGIN_LINK = ".nav-links a[href='/login']"
    PROPOSALS_LINK = "a[href='/proposals']"
    PROFILE_LINK = "a[href='/profile']"
    ERROR_MESSAGE = "#error-message"

    def goto(self, path: str = "/"):
        self.page.goto(f"{self.base_url}{path}")

    def get_error_message(self) -> str | None:
        error = self.page.locator(self.ERROR_MESSAGE)
        if error.is_visible():
            return error.text_content()
        return None

    def is_logged_in(self) -> bool:
        return self.page.locator(self.USER_MENU_BTN).is_visible()

    def open_user_menu(self):
        self.page.click(self.USER_MENU_BTN)
        self.page.wait_for_selector(f"{self.USER_DROPDOWN_MENU}:visible")

    def logout(self):
        self.open_user_menu()
        self.page.click(self.LOGOUT_BTN)

    def go_to_proposals(self):
        self.open_user_menu()
        self.page.click(self.PROPOSALS_LINK)

    def go_home(self):
        self.page.click(self.NAV_BRAND)
