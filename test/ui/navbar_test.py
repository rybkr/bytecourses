import re
from playwright.sync_api import Page, expect

from test.ui.pages import BasePage


def test_navbar_shows_about_and_browse_courses(page: Page, go_server: str):
    base_page = BasePage(page, go_server)
    base_page.goto()

    expect(page.locator(base_page.ABOUT_LINK)).to_be_visible()
    expect(page.locator(base_page.BROWSE_COURSES_LINK)).to_be_visible()


def test_navbar_shows_login_when_logged_out(page: Page, go_server: str):
    base_page = BasePage(page, go_server)
    base_page.goto()

    expect(page.locator(base_page.LOGIN_LINK)).to_be_visible()
    expect(page.locator(base_page.USER_MENU_BTN)).not_to_be_visible()


def test_navbar_shows_user_menu_when_logged_in(logged_in_user: Page, go_server: str):
    page = logged_in_user
    base_page = BasePage(page, go_server)

    expect(page.locator(base_page.USER_MENU_BTN)).to_be_visible()
    expect(page.locator(base_page.LOGIN_LINK)).not_to_be_visible()


def test_teach_dropdown_shows_for_logged_in_users(logged_in_user: Page, go_server: str):
    page = logged_in_user
    base_page = BasePage(page, go_server)

    expect(page.locator(base_page.TEACH_DROPDOWN_TRIGGER)).to_be_visible()

    base_page.open_teach_dropdown()

    expect(page.locator(base_page.TEACH_NEW_PROPOSAL)).to_be_visible()
    expect(page.locator(base_page.TEACH_MY_PROPOSALS)).to_be_visible()


def test_teach_dropdown_all_proposals_shows_for_admin(logged_in_admin: Page, go_server: str):
    page = logged_in_admin
    base_page = BasePage(page, go_server)

    base_page.open_teach_dropdown()

    expect(page.locator(base_page.TEACH_ALL_PROPOSALS)).to_be_visible()


def test_teach_dropdown_all_proposals_hidden_for_regular_user(logged_in_user: Page, go_server: str):
    page = logged_in_user
    base_page = BasePage(page, go_server)

    base_page.open_teach_dropdown()

    expect(page.locator(base_page.TEACH_ALL_PROPOSALS)).not_to_be_visible()


def test_user_dropdown_contains_profile_and_logout(logged_in_user: Page, go_server: str):
    page = logged_in_user
    base_page = BasePage(page, go_server)

    base_page.open_user_menu()

    expect(page.locator(base_page.USER_PROFILE_LINK)).to_be_visible()
    expect(page.locator(base_page.LOGOUT_BTN)).to_be_visible()


def test_navbar_brand_links_to_home(page: Page, go_server: str):
    base_page = BasePage(page, go_server)
    base_page.goto("/courses")

    base_page.go_home()

    expect(page).to_have_url(f"{go_server}/")


def test_about_link_navigates_to_about(page: Page, go_server: str):
    base_page = BasePage(page, go_server)
    base_page.goto()

    base_page.go_to_about()

    expect(page).to_have_url(f"{go_server}/about")


def test_browse_courses_link_navigates_to_courses(page: Page, go_server: str):
    base_page = BasePage(page, go_server)
    base_page.goto()

    base_page.go_to_browse_courses()

    expect(page).to_have_url(f"{go_server}/courses")


def test_teach_dropdown_new_proposal_navigates(logged_in_user: Page, go_server: str):
    page = logged_in_user
    base_page = BasePage(page, go_server)

    base_page.go_to_new_proposal()

    current_url = page.url
    assert re.match(rf"{re.escape(go_server)}/proposals/\d+/edit", current_url), f"URL {current_url} doesn't match expected pattern"


def test_teach_dropdown_my_proposals_navigates(logged_in_user: Page, go_server: str):
    page = logged_in_user
    base_page = BasePage(page, go_server)

    base_page.go_to_my_proposals()

    expect(page).to_have_url(f"{go_server}/proposals/mine")


def test_teach_dropdown_all_proposals_navigates_for_admin(logged_in_admin: Page, go_server: str):
    page = logged_in_admin
    base_page = BasePage(page, go_server)

    base_page.go_to_all_proposals()

    expect(page).to_have_url(f"{go_server}/proposals")


def test_mobile_menu_opens_and_closes(page: Page, go_server: str):
    page.set_viewport_size({"width": 600, "height": 800})
    base_page = BasePage(page, go_server)
    base_page.goto()

    expect(page.locator(base_page.HAMBURGER_BTN)).to_be_visible()
    expect(page.locator(f"{base_page.MOBILE_MENU}.active")).not_to_be_visible()

    base_page.open_mobile_menu()

    expect(page.locator(f"{base_page.MOBILE_MENU}.active")).to_be_visible()
    expect(page.locator(f"{base_page.MOBILE_MENU_OVERLAY}.active")).to_be_visible()

    base_page.close_mobile_menu()

    expect(page.locator(f"{base_page.MOBILE_MENU}.active")).not_to_be_visible()


def test_mobile_menu_shows_navigation_links(page: Page, go_server: str):
    page.set_viewport_size({"width": 600, "height": 800})
    base_page = BasePage(page, go_server)
    base_page.goto()

    base_page.open_mobile_menu()

    expect(page.locator(".mobile-menu-item[href='/about']")).to_be_visible()
    expect(page.locator(".mobile-menu-item[href='/courses']")).to_be_visible()
    expect(page.locator(".mobile-menu-item[href='/login']")).to_be_visible()


def test_mobile_menu_shows_proposals_for_logged_in_user(logged_in_user: Page, go_server: str):
    page = logged_in_user
    page.set_viewport_size({"width": 600, "height": 800})
    base_page = BasePage(page, go_server)

    base_page.open_mobile_menu()

    expect(page.locator(".mobile-menu-item[href='/proposals/new']")).to_be_visible()
    expect(page.locator(".mobile-menu-item[href='/proposals/mine']")).to_be_visible()
    expect(page.locator(".mobile-menu-item[href='/proposals']")).not_to_be_visible()


def test_mobile_menu_shows_all_proposals_for_admin(logged_in_admin: Page, go_server: str):
    page = logged_in_admin
    page.set_viewport_size({"width": 600, "height": 800})
    base_page = BasePage(page, go_server)

    base_page.open_mobile_menu()

    expect(page.locator(".mobile-menu-item[href='/proposals']")).to_be_visible()
