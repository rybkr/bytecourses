from playwright.sync_api import Page
from .base_page import BasePage


class RegisterPage(BasePage):
    PATH = "/register"

    NAME_INPUT = "#name"
    EMAIL_INPUT = "#email"
    PASSWORD_INPUT = "#password"
    SUBMIT_BTN = "button[type='submit']"
    LOGIN_LINK = "a[href='/login']"

    def __init__(self, page: Page, base_url: str):
        super().__init__(page, base_url)

    def navigate(self):
        self.goto(self.PATH)

    def fill_name(self, name: str):
        self.page.fill(self.NAME_INPUT, name)

    def fill_email(self, email: str):
        self.page.fill(self.EMAIL_INPUT, email)

    def fill_password(self, password: str):
        self.page.fill(self.PASSWORD_INPUT, password)

    def submit(self):
        self.page.click(self.SUBMIT_BTN)

    def register(self, name: str, email: str, password: str):
        self.fill_name(name)
        self.fill_email(email)
        self.fill_password(password)
        self.submit()

    def go_to_login(self):
        self.page.click(self.LOGIN_LINK)
