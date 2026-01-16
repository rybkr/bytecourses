from playwright.sync_api import Page
from .base_page import BasePage


class LoginPage(BasePage):
    PATH = "/login"

    EMAIL_INPUT = "#email"
    PASSWORD_INPUT = "#password"
    SUBMIT_BTN = "button[type='submit']"
    REGISTER_LINK = "a[href='/register']"
    FORGOT_PASSWORD_LINK = "a[href='/forgot-password']"

    def __init__(self, page: Page, base_url: str):
        super().__init__(page, base_url)

    def navigate(self):
        self.goto(self.PATH)

    def fill_email(self, email: str):
        self.page.fill(self.EMAIL_INPUT, email)

    def fill_password(self, password: str):
        self.page.fill(self.PASSWORD_INPUT, password)

    def submit(self):
        self.page.click(self.SUBMIT_BTN)

    def login(self, email: str, password: str):
        self.fill_email(email)
        self.fill_password(password)
        self.submit()

    def go_to_register(self):
        self.page.click(self.REGISTER_LINK)
