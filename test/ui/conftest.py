import pytest
from playwright.sync_api import Page

from test.conftest import USER_EMAIL, USER_PASSWORD, ADMIN_EMAIL, ADMIN_PASSWORD


@pytest.fixture(scope="function")
def logged_in_user(page: Page, go_server: str):
    page.goto(f"{go_server}/login")
    page.fill("#email", USER_EMAIL)
    page.fill("#password", USER_PASSWORD)
    page.click("button[type='submit']")
    page.wait_for_url(f"{go_server}/")
    return page


@pytest.fixture(scope="function")
def logged_in_admin(page: Page, go_server: str):
    page.goto(f"{go_server}/login")
    page.fill("#email", ADMIN_EMAIL)
    page.fill("#password", ADMIN_PASSWORD)
    page.click("button[type='submit']")
    page.wait_for_url(f"{go_server}/")
    return page
