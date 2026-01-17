from playwright.sync_api import Page


class BasePage:
    def __init__(self, page: Page, base_url: str):
        self.page = page
        self.base_url = base_url

    NAV_BRAND = ".nav-brand a"
    ABOUT_LINK = ".nav-links a[href='/about']"
    BROWSE_COURSES_LINK = ".nav-links a[href='/courses']"
    TEACH_DROPDOWN_TRIGGER = ".teach-menu-trigger"
    TEACH_DROPDOWN_MENU = "#teachDropdownMenu"
    TEACH_NEW_PROPOSAL = ".teach-dropdown-item[href='/proposals/new']"
    TEACH_MY_PROPOSALS = ".teach-dropdown-item[href='/proposals/mine']"
    TEACH_ALL_PROPOSALS = ".teach-dropdown-item[href='/proposals']"
    USER_MENU_BTN = ".user-menu-btn"
    USER_DROPDOWN_MENU = "#userDropdownMenu"
    USER_PROFILE_LINK = ".user-dropdown-item[href='/profile']"
    LOGOUT_BTN = ".user-dropdown-logout"
    LOGIN_LINK = ".nav-links a[href='/login']"
    HAMBURGER_BTN = ".hamburger-btn"
    MOBILE_MENU = "#mobileMenu"
    MOBILE_MENU_OVERLAY = "#mobileMenuOverlay"
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
        self.page.wait_for_selector(f"{self.USER_DROPDOWN_MENU}.show", state="visible")

    def logout(self):
        self.open_user_menu()
        self.page.click(self.LOGOUT_BTN)

    def go_to_profile(self):
        self.open_user_menu()
        self.page.click(self.USER_PROFILE_LINK)

    def open_teach_dropdown(self):
        self.page.click(self.TEACH_DROPDOWN_TRIGGER)
        self.page.wait_for_selector(f"{self.TEACH_DROPDOWN_MENU}.show", state="visible")

    def go_to_new_proposal(self):
        self.open_teach_dropdown()
        self.page.click(self.TEACH_NEW_PROPOSAL)

    def go_to_my_proposals(self):
        self.open_teach_dropdown()
        self.page.click(self.TEACH_MY_PROPOSALS)

    def go_to_all_proposals(self):
        self.open_teach_dropdown()
        self.page.click(self.TEACH_ALL_PROPOSALS)

    def go_to_about(self):
        self.page.click(self.ABOUT_LINK)

    def go_to_browse_courses(self):
        self.page.click(self.BROWSE_COURSES_LINK)

    def go_home(self):
        self.page.click(self.NAV_BRAND)

    def open_mobile_menu(self):
        self.page.click(self.HAMBURGER_BTN)
        self.page.wait_for_selector(f"{self.MOBILE_MENU}.active", state="visible")

    def close_mobile_menu(self):
        if self.page.locator(f"{self.MOBILE_MENU}.active").is_visible():
            self.page.evaluate("closeMobileMenu()")
            self.page.wait_for_selector(f"{self.MOBILE_MENU}.active", state="hidden")

    def is_mobile_menu_visible(self) -> bool:
        return self.page.locator(f"{self.MOBILE_MENU}.active").is_visible()
