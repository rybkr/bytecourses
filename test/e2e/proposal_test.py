from http import HTTPStatus

from .conftest import register_and_login


class TestProposalCreate:
    def test_creates_proposal_with_title_and_summary(self, api_url):
        session = register_and_login(api_url, "author@example.com", "password123")

        r = session.post(
            f"{api_url}/proposals",
            json={"title": "My Course", "summary": "Course summary."},
        )
        assert r.status_code == HTTPStatus.CREATED
        assert "id" in r.json()

    def test_creates_proposal_with_all_fields(self, api_url):
        session = register_and_login(api_url, "author@example.com", "password123")

        payload = {
            "title": "Complete Course",
            "summary": "A comprehensive course.",
            "target_audience": "Developers",
            "learning_objectives": "Learn everything",
            "outline": "- item 1\n- item 2",
            "assumed_prerequisites": "Basic programming",
        }
        r = session.post(f"{api_url}/proposals", json=payload)
        assert r.status_code == HTTPStatus.CREATED

        proposal_id = r.json()["id"]
        r = session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.status_code == HTTPStatus.OK
        for key, value in payload.items():
            assert r.json()[key] == value

    def test_creates_proposal_with_qualifications(self, api_url):
        session = register_and_login(api_url, "author@example.com", "password123")

        r = session.post(
            f"{api_url}/proposals",
            json={
                "title": "Course",
                "summary": "Summary",
                "qualifications": "10 years experience",
            },
        )
        assert r.status_code == HTTPStatus.CREATED

        proposal_id = r.json()["id"]
        r = session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.status_code == HTTPStatus.OK
        assert r.json()["qualifications"] == "10 years experience"

    def test_new_proposal_has_draft_status(self, api_url, user_session):
        r = user_session.post(
            f"{api_url}/proposals",
            json={"title": "Draft Proposal", "summary": "Summary"},
        )
        assert r.status_code == HTTPStatus.CREATED

        proposal_id = r.json()["id"]
        r = user_session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.json()["status"] == "draft"


class TestProposalRead:
    def test_lists_own_proposals(self, api_url):
        session = register_and_login(api_url, "author@example.com", "password123")

        session.post(
            f"{api_url}/proposals",
            json={"title": "First Course", "summary": "First summary"},
        )
        session.post(
            f"{api_url}/proposals",
            json={"title": "Second Course", "summary": "Second summary"},
        )

        r = session.get(f"{api_url}/proposals")
        assert r.status_code == HTTPStatus.OK

        proposals = sorted(r.json(), key=lambda p: p["id"])
        assert len(proposals) == 2
        assert proposals[0]["title"] == "First Course"
        assert proposals[1]["title"] == "Second Course"

    def test_returns_empty_list_when_no_proposals(self, api_url):
        session = register_and_login(api_url, "newauthor@example.com", "password123")

        r = session.get(f"{api_url}/proposals")
        assert r.status_code == HTTPStatus.OK
        assert r.json() == []

    def test_gets_proposal_by_id(self, api_url):
        session = register_and_login(api_url, "author@example.com", "password123")

        r = session.post(
            f"{api_url}/proposals",
            json={"title": "My Course", "summary": "Summary"},
        )
        proposal_id = r.json()["id"]

        r = session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.status_code == HTTPStatus.OK
        assert r.json()["title"] == "My Course"

    def test_returns_404_for_nonexistent_proposal(self, api_url):
        session = register_and_login(api_url, "author@example.com", "password123")

        r = session.get(f"{api_url}/proposals/{2**63 - 1}")
        assert r.status_code == HTTPStatus.NOT_FOUND

    def test_rejects_delete_method_on_proposals_list(self, api_url):
        session = register_and_login(api_url, "author@example.com", "password123")

        r = session.delete(f"{api_url}/proposals")
        assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED


class TestProposalUpdate:
    def test_updates_proposal_title(self, api_url):
        session = register_and_login(api_url, "author@example.com", "password123")

        r = session.post(
            f"{api_url}/proposals",
            json={"title": "Original Title", "summary": "Summary"},
        )
        proposal_id = r.json()["id"]

        r = session.patch(
            f"{api_url}/proposals/{proposal_id}",
            json={"title": "New Title"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.json()["title"] == "New Title"

    def test_patch_replaces_all_fields(self, api_url, user_session):
        r = user_session.post(
            f"{api_url}/proposals",
            json={"title": "Title", "summary": "Summary", "outline": "Outline"},
        )
        proposal_id = r.json()["id"]

        r = user_session.patch(
            f"{api_url}/proposals/{proposal_id}",
            json={"title": "New Title"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = user_session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.json()["title"] == "New Title"
        assert r.json()["summary"] == ""
        assert r.json()["outline"] == ""

    def test_rejects_update_on_submitted_proposal(self, api_url, user_session):
        r = user_session.post(
            f"{api_url}/proposals",
            json={"title": "Title", "summary": "Summary"},
        )
        proposal_id = r.json()["id"]

        user_session.post(f"{api_url}/proposals/{proposal_id}/actions/submit")

        r = user_session.patch(
            f"{api_url}/proposals/{proposal_id}",
            json={"title": "New Title"},
        )
        assert r.status_code == HTTPStatus.CONFLICT


class TestProposalDelete:
    def test_deletes_draft_proposal(self, api_url, user_session):
        r = user_session.post(
            f"{api_url}/proposals",
            json={"title": "To Delete", "summary": "Summary"},
        )
        proposal_id = r.json()["id"]

        r = user_session.delete(f"{api_url}/proposals/{proposal_id}")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = user_session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.status_code == HTTPStatus.NOT_FOUND


class TestProposalActions:
    def test_submit_changes_status_to_submitted(self, api_url):
        session = register_and_login(api_url, "author@example.com", "password123")

        r = session.post(
            f"{api_url}/proposals",
            json={"title": "Title", "summary": "Summary"},
        )
        proposal_id = r.json()["id"]

        r = session.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.json()["status"] == "submitted"

    def test_unknown_action_returns_400(self, api_url):
        session = register_and_login(api_url, "author@example.com", "password123")

        r = session.post(
            f"{api_url}/proposals",
            json={"title": "Title", "summary": "Summary"},
        )
        proposal_id = r.json()["id"]

        r = session.post(f"{api_url}/proposals/{proposal_id}/actions/unknown")
        assert r.status_code == HTTPStatus.BAD_REQUEST


class TestProposalPermissions:
    def test_user_cannot_view_other_users_proposal(self, api_url):
        user_a = register_and_login(api_url, "usera@example.com", "password")
        user_b = register_and_login(api_url, "userb@example.com", "password")

        r = user_a.post(
            f"{api_url}/proposals",
            json={"title": "A's Proposal", "summary": "Summary"},
        )
        proposal_a_id = r.json()["id"]

        r = user_b.get(f"{api_url}/proposals/{proposal_a_id}")
        assert r.status_code == HTTPStatus.NOT_FOUND

    def test_users_see_only_their_own_proposals(self, api_url):
        user_a = register_and_login(api_url, "usera@example.com", "password")
        user_b = register_and_login(api_url, "userb@example.com", "password")

        user_a.post(
            f"{api_url}/proposals",
            json={"title": "A's Proposal", "summary": "Summary"},
        )
        user_b.post(
            f"{api_url}/proposals",
            json={"title": "B's Proposal", "summary": "Summary"},
        )

        r = user_a.get(f"{api_url}/proposals")
        assert len(r.json()) == 1
        assert r.json()[0]["title"] == "A's Proposal"

        r = user_b.get(f"{api_url}/proposals")
        assert len(r.json()) == 1
        assert r.json()[0]["title"] == "B's Proposal"


class TestAdminProposalAccess:
    def test_admin_sees_submitted_proposals(self, api_url, admin_session):
        user = register_and_login(api_url, "author@example.com", "password123")

        r = user.post(
            f"{api_url}/proposals",
            json={"title": "Draft Proposal", "summary": "Summary"},
        )
        draft_id = r.json()["id"]

        r = user.post(
            f"{api_url}/proposals",
            json={"title": "Submitted Proposal", "summary": "Summary"},
        )
        submitted_id = r.json()["id"]
        user.post(f"{api_url}/proposals/{submitted_id}/actions/submit")

        r = admin_session.get(f"{api_url}/proposals")
        assert r.status_code == HTTPStatus.OK
        ids = [p["id"] for p in r.json()]
        assert submitted_id in ids
        assert draft_id not in ids

    def test_admin_cannot_view_draft_proposals(self, api_url, admin_session):
        user = register_and_login(api_url, "author@example.com", "password123")

        r = user.post(
            f"{api_url}/proposals",
            json={"title": "Draft", "summary": "Summary"},
        )
        draft_id = r.json()["id"]

        r = admin_session.get(f"{api_url}/proposals/{draft_id}")
        assert r.status_code == HTTPStatus.NOT_FOUND

    def test_admin_can_view_submitted_proposal(self, api_url, admin_session):
        user = register_and_login(api_url, "author@example.com", "password123")

        r = user.post(
            f"{api_url}/proposals",
            json={"title": "To Submit", "summary": "Summary"},
        )
        proposal_id = r.json()["id"]
        user.post(f"{api_url}/proposals/{proposal_id}/actions/submit")

        r = admin_session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.status_code == HTTPStatus.OK
        assert r.json()["title"] == "To Submit"

    def test_admin_sees_own_proposals_via_mine_endpoint(self, api_url, admin_session):
        r = admin_session.post(
            f"{api_url}/proposals",
            json={"title": "Admin's Proposal", "summary": "Summary"},
        )
        admin_proposal_id = r.json()["id"]

        r = admin_session.get(f"{api_url}/proposals/mine")
        assert r.status_code == HTTPStatus.OK
        assert any(p["id"] == admin_proposal_id for p in r.json())


class TestProposalWorkflowHappyPath:
    def test_draft_submit_approve_workflow(self, api_url, user_session, admin_session):
        r = user_session.post(
            f"{api_url}/proposals",
            json={"title": "Title", "summary": "Summary", "outline": "Outline"},
        )
        proposal_id = r.json()["id"]

        r = user_session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.json()["status"] == "draft"

        r = admin_session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.status_code == HTTPStatus.NOT_FOUND

        r = user_session.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.status_code == HTTPStatus.OK

        r = user_session.patch(
            f"{api_url}/proposals/{proposal_id}",
            json={"title": "New Title"},
        )
        assert r.status_code == HTTPStatus.CONFLICT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/approve",
            json={"review_notes": "Looks good!"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = user_session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.json()["status"] == "approved"


class TestProposalWorkflowChangesRequested:
    def test_changes_requested_allows_editing(
        self, api_url, user_session, admin_session
    ):
        r = user_session.post(
            f"{api_url}/proposals",
            json={
                "title": "Original Title",
                "summary": "Summary",
                "outline": "Outline",
            },
        )
        proposal_id = r.json()["id"]

        user_session.post(f"{api_url}/proposals/{proposal_id}/actions/submit")

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/request-changes",
            json={"review_notes": "Please add more detail"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = user_session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.json()["status"] == "changes_requested"
        assert r.json()["review_notes"] == "Please add more detail"

        r = user_session.patch(
            f"{api_url}/proposals/{proposal_id}",
            json={
                "title": "Updated Title",
                "summary": "Better summary",
                "outline": "Outline",
            },
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = user_session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.json()["title"] == "Updated Title"

        user_session.post(f"{api_url}/proposals/{proposal_id}/actions/submit")

        r = admin_session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.json()["status"] == "submitted"

        admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/approve",
            json={"review_notes": "Approved"},
        )

        r = user_session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.json()["status"] == "approved"


class TestProposalWorkflowRejection:
    def test_reject_sets_rejected_status(self, api_url, user_session, admin_session):
        r = user_session.post(
            f"{api_url}/proposals",
            json={"title": "Title", "summary": "Summary", "outline": "Outline"},
        )
        proposal_id = r.json()["id"]

        user_session.post(f"{api_url}/proposals/{proposal_id}/actions/submit")

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/reject",
            json={"review_notes": "Not suitable"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = user_session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.json()["status"] == "rejected"

    def test_rejected_proposal_cannot_be_edited(
        self, api_url, user_session, admin_session
    ):
        r = user_session.post(
            f"{api_url}/proposals",
            json={"title": "Title", "summary": "Summary"},
        )
        proposal_id = r.json()["id"]

        user_session.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/reject",
            json={"review_notes": "Rejected"},
        )

        r = user_session.patch(
            f"{api_url}/proposals/{proposal_id}",
            json={"title": "New Title"},
        )
        assert r.status_code == HTTPStatus.CONFLICT


class TestProposalWorkflowWithdrawal:
    def test_withdraw_sets_withdrawn_status(self, api_url, user_session, admin_session):
        r = user_session.post(
            f"{api_url}/proposals",
            json={"title": "Title", "summary": "Summary", "outline": "Outline"},
        )
        proposal_id = r.json()["id"]

        user_session.post(f"{api_url}/proposals/{proposal_id}/actions/submit")

        r = user_session.post(f"{api_url}/proposals/{proposal_id}/actions/withdraw")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = user_session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.json()["status"] == "withdrawn"

    def test_admin_cannot_see_withdrawn_proposal(
        self, api_url, user_session, admin_session
    ):
        r = user_session.post(
            f"{api_url}/proposals",
            json={"title": "Title", "summary": "Summary"},
        )
        proposal_id = r.json()["id"]

        user_session.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        user_session.post(f"{api_url}/proposals/{proposal_id}/actions/withdraw")

        r = admin_session.get(f"{api_url}/proposals/{proposal_id}")
        assert r.status_code == HTTPStatus.NOT_FOUND

    def test_withdrawn_proposal_cannot_be_edited(self, api_url, user_session):
        r = user_session.post(
            f"{api_url}/proposals",
            json={"title": "Title", "summary": "Summary"},
        )
        proposal_id = r.json()["id"]

        user_session.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        user_session.post(f"{api_url}/proposals/{proposal_id}/actions/withdraw")

        r = user_session.patch(
            f"{api_url}/proposals/{proposal_id}",
            json={"title": "New Title"},
        )
        assert r.status_code == HTTPStatus.CONFLICT
