from playwright.sync_api import Page, expect

from .pages import ProposalsPage, ProposalEditPage


def test_proposals_list_shows_user_proposals(logged_in_user: Page, go_server: str):
    page = logged_in_user
    proposals_page = ProposalsPage(page, go_server)
    proposals_page.navigate()

    proposals_page.wait_for_proposals_loaded()

    assert proposals_page.get_proposal_count() == 4

    titles = proposals_page.get_proposal_titles()
    assert "Practical Distributed Systems in Go" in titles


def test_create_new_proposal(logged_in_user: Page, go_server: str):
    page = logged_in_user
    proposals_page = ProposalsPage(page, go_server)
    proposals_page.navigate()
    proposals_page.wait_for_proposals_loaded()

    proposals_page.click_new_proposal()

    page.wait_for_url(f"{go_server}/proposals/*/edit")

    edit_page = ProposalEditPage(page, go_server)
    edit_page.fill_all_fields(
        title="Test Proposal Title",
        summary="This is a test proposal summary for UI testing.",
        qualifications="I have experience in testing and quality assurance.",
        target_audience="Developers who want to learn about testing.",
        learning_objectives="- Learn how to write tests\n- Understand testing patterns",
        outline="1. Introduction\n2. Writing Tests\n3. Best Practices",
        prerequisites="Basic programming knowledge",
    )

    edit_page.wait_for_autosave(timeout=10000)

    edit_page.click_save_draft()
    page.wait_for_function(
        "() => !window.location.pathname.endsWith('/edit')", timeout=10000
    )

    proposals_page.navigate()
    proposals_page.wait_for_proposals_loaded()

    titles = proposals_page.get_proposal_titles()
    assert "Test Proposal Title" in titles


def test_autosave_triggers_after_delay(logged_in_user: Page, go_server: str):
    page = logged_in_user
    proposals_page = ProposalsPage(page, go_server)
    proposals_page.navigate()
    proposals_page.wait_for_proposals_loaded()

    proposals_page.click_proposal_by_title("Practical Distributed Systems in Go")
    page.wait_for_url(f"{go_server}/proposals/*")

    page.click("a.btn:has-text('Edit')")
    page.wait_for_url(f"{go_server}/proposals/*/edit")

    edit_page = ProposalEditPage(page, go_server)

    original_title = edit_page.get_title()
    edit_page.fill_title(original_title + " Updated")

    edit_page.wait_for_autosave(timeout=10000)

    status = edit_page.get_save_status()
    assert "Saved at" in status


def test_edit_proposal_fields(logged_in_user: Page, go_server: str):
    page = logged_in_user
    proposals_page = ProposalsPage(page, go_server)
    proposals_page.navigate()
    proposals_page.wait_for_proposals_loaded()

    proposals_page.click_proposal_by_title("Practical Distributed Systems in Go")
    page.wait_for_url(f"{go_server}/proposals/*")

    page.click("a.btn:has-text('Edit')")
    page.wait_for_url(f"{go_server}/proposals/*/edit")

    edit_page = ProposalEditPage(page, go_server)

    new_title = "Updated Distributed Systems Course"
    edit_page.fill_title(new_title)

    page.locator("#summary").focus()

    edit_page.wait_for_autosave(timeout=10000)

    edit_page.click_save_draft()
    page.wait_for_url(f"{go_server}/proposals/*", wait_until="networkidle")

    proposals_page.navigate()
    proposals_page.wait_for_proposals_loaded()

    titles = proposals_page.get_proposal_titles()
    assert new_title in titles


def test_submit_proposal_changes_status(logged_in_user: Page, go_server: str):
    page = logged_in_user

    proposals_page = ProposalsPage(page, go_server)
    proposals_page.navigate()
    proposals_page.wait_for_proposals_loaded()

    proposals_page.click_new_proposal()
    page.wait_for_url(f"{go_server}/proposals/*/edit")

    edit_page = ProposalEditPage(page, go_server)
    edit_page.fill_all_fields(
        title="Proposal to Submit",
        summary="This proposal will be submitted for review.",
        qualifications="Qualified instructor with relevant experience.",
        target_audience="All developers.",
        learning_objectives="- Master the subject matter",
        outline="1. Introduction\n2. Main content\n3. Conclusion",
        prerequisites="None",
    )

    edit_page.wait_for_autosave(timeout=10000)

    edit_page.click_submit()

    page.wait_for_function(
        "() => !window.location.pathname.endsWith('/edit')", timeout=10000
    )

    expect(page.locator(".status-badge")).to_contain_text("submitted")
