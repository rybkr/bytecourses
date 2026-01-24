from http import HTTPStatus

from .conftest import register_and_login


class TestContentEndpoints:
    """Basic tests for content endpoints using new nested route structure."""

    def test_content_endpoints_use_nested_structure(self, api_url, admin_session):
        """Verify content endpoints are accessible via new nested route structure."""
        author = register_and_login(api_url, "author@example.com", "password123")

        # Create proposal and course
        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Test Course",
                "summary": "Summary",
                "target_audience": "Developers",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        course_id = r.json()["id"]

        # Create module
        r = author.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Test Module", "description": "Description", "order": 1},
        )
        module_id = r.json()["id"]

        # Test content creation with new nested route
        r = author.post(
            f"{api_url}/courses/{course_id}/modules/{module_id}/content",
            json={
                "title": "Test Reading",
                "order": 1,
                "format": "markdown",
                "content": "# Test Content",
            },
        )
        assert r.status_code == HTTPStatus.CREATED
        assert "id" in r.json()
        content_id = r.json()["id"]

        # Test content listing
        r = author.get(f"{api_url}/courses/{course_id}/modules/{module_id}/content")
        assert r.status_code == HTTPStatus.OK
        assert len(r.json()) == 1
        assert r.json()[0]["id"] == content_id

        # Test content retrieval
        r = author.get(
            f"{api_url}/courses/{course_id}/modules/{module_id}/content/{content_id}"
        )
        assert r.status_code == HTTPStatus.OK
        assert r.json()["id"] == content_id
        assert r.json()["title"] == "Test Reading"

        # Test content update
        r = author.patch(
            f"{api_url}/courses/{course_id}/modules/{module_id}/content/{content_id}",
            json={
                "title": "Updated Reading",
                "order": 1,
                "format": "markdown",
                "content": "# Updated Content",
            },
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        # Verify update
        r = author.get(
            f"{api_url}/courses/{course_id}/modules/{module_id}/content/{content_id}"
        )
        assert r.json()["title"] == "Updated Reading"

        # Test content publish
        r = author.post(
            f"{api_url}/courses/{course_id}/modules/{module_id}/content/{content_id}/actions/publish"
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        # Test content deletion
        r = author.delete(
            f"{api_url}/courses/{course_id}/modules/{module_id}/content/{content_id}"
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        # Verify deletion
        r = author.get(
            f"{api_url}/courses/{course_id}/modules/{module_id}/content/{content_id}"
        )
        assert r.status_code == HTTPStatus.NOT_FOUND
