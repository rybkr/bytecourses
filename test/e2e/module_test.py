import requests
from http import HTTPStatus

from .conftest import register_and_login


def create_course(session, api_url, title="Test Course"):
    """Helper to create a course and return its ID."""
    r = session.post(
        f"{api_url}/courses",
        json={"title": title, "summary": "Test course summary"},
    )
    assert r.status_code == HTTPStatus.CREATED
    return r.json()["id"]


class TestModuleCreate:
    def test_creates_module_with_title(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        r = session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Introduction"},
        )
        assert r.status_code == HTTPStatus.CREATED
        assert "id" in r.json()
        assert r.json()["title"] == "Introduction"
        assert r.json()["course_id"] == course_id
        assert r.json()["position"] == 1

    def test_auto_increments_position(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        r1 = session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 1"},
        )
        r2 = session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 2"},
        )
        r3 = session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 3"},
        )

        assert r1.json()["position"] == 1
        assert r2.json()["position"] == 2
        assert r3.json()["position"] == 3

    def test_rejects_empty_title(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        r = session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": ""},
        )
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_missing_title(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        r = session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={},
        )
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_nonexistent_course(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        r = session.post(
            f"{api_url}/courses/{2**63 - 1}/modules",
            json={"title": "Module 1"},
        )
        assert r.status_code == HTTPStatus.NOT_FOUND

    def test_rejects_unauthenticated_request(self, api_url):
        r = requests.post(
            f"{api_url}/courses/1/modules",
            json={"title": "Module 1"},
        )
        assert r.status_code == HTTPStatus.UNAUTHORIZED


class TestModuleRead:
    def test_gets_module_by_id(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        r = session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 1"},
        )
        module_id = r.json()["id"]

        r = session.get(f"{api_url}/courses/{course_id}/modules/{module_id}")
        assert r.status_code == HTTPStatus.OK
        assert r.json()["id"] == module_id
        assert r.json()["title"] == "Module 1"

    def test_returns_404_for_nonexistent_module(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        r = session.get(f"{api_url}/courses/{course_id}/modules/{2**63 - 1}")
        assert r.status_code == HTTPStatus.NOT_FOUND

    def test_lists_modules_for_course(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 1"},
        )
        session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 2"},
        )

        r = session.get(f"{api_url}/courses/{course_id}/modules")
        assert r.status_code == HTTPStatus.OK
        modules = r.json()
        assert len(modules) == 2
        assert modules[0]["title"] == "Module 1"
        assert modules[1]["title"] == "Module 2"

    def test_lists_modules_empty_course(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        r = session.get(f"{api_url}/courses/{course_id}/modules")
        assert r.status_code == HTTPStatus.OK
        assert r.json() == []

    def test_lists_modules_sorted_by_position(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 1"},
        )
        session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 2"},
        )
        session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 3"},
        )

        r = session.get(f"{api_url}/courses/{course_id}/modules")
        modules = r.json()
        positions = [m["position"] for m in modules]
        assert positions == sorted(positions)


class TestModuleUpdate:
    def test_updates_module_title(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        r = session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Original Title"},
        )
        module_id = r.json()["id"]

        r = session.patch(
            f"{api_url}/courses/{course_id}/modules/{module_id}",
            json={"title": "Updated Title"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = session.get(f"{api_url}/courses/{course_id}/modules/{module_id}")
        assert r.json()["title"] == "Updated Title"

    def test_returns_404_for_nonexistent_module(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        r = session.patch(
            f"{api_url}/courses/{course_id}/modules/{2**63 - 1}",
            json={"title": "Updated Title"},
        )
        assert r.status_code == HTTPStatus.NOT_FOUND


class TestModuleDelete:
    def test_deletes_module(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        r = session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module to Delete"},
        )
        module_id = r.json()["id"]

        r = session.delete(f"{api_url}/courses/{course_id}/modules/{module_id}")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = session.get(f"{api_url}/courses/{course_id}/modules/{module_id}")
        assert r.status_code == HTTPStatus.NOT_FOUND

    def test_returns_404_for_nonexistent_module(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        r = session.delete(f"{api_url}/courses/{course_id}/modules/{2**63 - 1}")
        assert r.status_code == HTTPStatus.NOT_FOUND


class TestModuleReorder:
    def test_reorders_modules(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        r1 = session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 1"},
        )
        r2 = session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 2"},
        )
        r3 = session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 3"},
        )
        id1, id2, id3 = r1.json()["id"], r2.json()["id"], r3.json()["id"]

        # Reorder: 3, 1, 2
        r = session.post(
            f"{api_url}/courses/{course_id}/modules/reorder",
            json={"module_ids": [id3, id1, id2]},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = session.get(f"{api_url}/courses/{course_id}/modules")
        modules = r.json()
        assert len(modules) == 3
        assert modules[0]["title"] == "Module 3"
        assert modules[0]["position"] == 1
        assert modules[1]["title"] == "Module 1"
        assert modules[1]["position"] == 2
        assert modules[2]["title"] == "Module 2"
        assert modules[2]["position"] == 3

    def test_rejects_empty_module_ids(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        r = session.post(
            f"{api_url}/courses/{course_id}/modules/reorder",
            json={"module_ids": []},
        )
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_nonexistent_module_ids(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        r = session.post(
            f"{api_url}/courses/{course_id}/modules/reorder",
            json={"module_ids": [2**63 - 1]},
        )
        assert r.status_code in (HTTPStatus.NOT_FOUND, HTTPStatus.INTERNAL_SERVER_ERROR)


class TestModulePermissions:
    def test_instructor_can_manage_own_course_modules(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")
        course_id = create_course(session, api_url)

        # Create
        r = session.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 1"},
        )
        assert r.status_code == HTTPStatus.CREATED
        module_id = r.json()["id"]

        # Read
        r = session.get(f"{api_url}/courses/{course_id}/modules/{module_id}")
        assert r.status_code == HTTPStatus.OK

        # Update
        r = session.patch(
            f"{api_url}/courses/{course_id}/modules/{module_id}",
            json={"title": "Updated"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        # Delete
        r = session.delete(f"{api_url}/courses/{course_id}/modules/{module_id}")
        assert r.status_code == HTTPStatus.NO_CONTENT

    def test_other_instructor_cannot_create_module(self, api_url):
        instructor1 = register_and_login(
            api_url, "instructor1@example.com", "password123"
        )
        instructor2 = register_and_login(
            api_url, "instructor2@example.com", "password123"
        )

        course_id = create_course(instructor1, api_url)

        r = instructor2.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Unauthorized Module"},
        )
        assert r.status_code == HTTPStatus.FORBIDDEN

    def test_other_instructor_cannot_update_module(self, api_url):
        instructor1 = register_and_login(
            api_url, "instructor1@example.com", "password123"
        )
        instructor2 = register_and_login(
            api_url, "instructor2@example.com", "password123"
        )

        course_id = create_course(instructor1, api_url)
        r = instructor1.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 1"},
        )
        module_id = r.json()["id"]

        r = instructor2.patch(
            f"{api_url}/courses/{course_id}/modules/{module_id}",
            json={"title": "Unauthorized Update"},
        )
        assert r.status_code == HTTPStatus.FORBIDDEN

    def test_other_instructor_cannot_delete_module(self, api_url):
        instructor1 = register_and_login(
            api_url, "instructor1@example.com", "password123"
        )
        instructor2 = register_and_login(
            api_url, "instructor2@example.com", "password123"
        )

        course_id = create_course(instructor1, api_url)
        r = instructor1.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 1"},
        )
        module_id = r.json()["id"]

        r = instructor2.delete(f"{api_url}/courses/{course_id}/modules/{module_id}")
        assert r.status_code == HTTPStatus.FORBIDDEN

    def test_other_instructor_cannot_reorder_modules(self, api_url):
        instructor1 = register_and_login(
            api_url, "instructor1@example.com", "password123"
        )
        instructor2 = register_and_login(
            api_url, "instructor2@example.com", "password123"
        )

        course_id = create_course(instructor1, api_url)
        r1 = instructor1.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 1"},
        )
        r2 = instructor1.post(
            f"{api_url}/courses/{course_id}/modules",
            json={"title": "Module 2"},
        )
        id1, id2 = r1.json()["id"], r2.json()["id"]

        r = instructor2.post(
            f"{api_url}/courses/{course_id}/modules/reorder",
            json={"module_ids": [id2, id1]},
        )
        assert r.status_code == HTTPStatus.FORBIDDEN
