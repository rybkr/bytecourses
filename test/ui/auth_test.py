from playwright.sync_api import Page, expect

from test.conftest import USER_EMAIL, USER_PASSWORD
from .pages import LoginPage, RegisterPage


def test_login_success(page: Page, go_server: str):
    login_page = LoginPage(page, go_server)
    login_page.navigate()

    login_page.login(USER_EMAIL, USER_PASSWORD)

    expect(page).to_have_url(f"{go_server}/")
    expect(page.locator(".user-menu-btn")).to_be_visible()


def test_login_failure_wrong_password(page: Page, go_server: str):
    login_page = LoginPage(page, go_server)
    login_page.navigate()

    login_page.login(USER_EMAIL, "wrongpassword")

    expect(page).to_have_url(f"{go_server}/login")
    expect(page.locator("#error-message")).to_be_visible()
    expect(page.locator("#error-message")).to_contain_text("invalid credentials")


def test_logout_redirects_to_home(logged_in_user: Page, go_server: str):
    page = logged_in_user

    expect(page.locator(".user-menu-btn")).to_be_visible()

    page.click(".user-menu-btn")
    page.wait_for_selector("#userDropdownMenu:visible")
    page.click(".user-dropdown-logout")

    expect(page).to_have_url(f"{go_server}/")
    expect(page.locator(".nav-links a[href='/login']")).to_be_visible()
    expect(page.locator(".user-menu-btn")).not_to_be_visible()


def test_register_success(page: Page, go_server: str):
    register_page = RegisterPage(page, go_server)
    register_page.navigate()

    register_page.register(
        name="Test User",
        email="newuser@example.com",
        password="securepassword123",
    )

    expect(page).to_have_url(f"{go_server}/login")

    login_page = LoginPage(page, go_server)
    login_page.login("newuser@example.com", "securepassword123")

    expect(page).to_have_url(f"{go_server}/")
    expect(page.locator(".user-menu-btn")).to_be_visible()
