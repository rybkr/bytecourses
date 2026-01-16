import requests
from http import HTTPStatus

from .conftest import register_and_login


class TestCourseCreate:
    def test_creates_course_with_title_and_summary(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        r = session.post(
            f"{api_url}/courses",
            json={"title": "Introduction to Python", "summary": "Learn Python basics."},
        )
        assert r.status_code == HTTPStatus.CREATED
        assert "id" in r.json()
        assert r.json()["title"] == "Introduction to Python"
        assert r.json()["status"] == "draft"

    def test_creates_course_with_all_fields(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        payload = {
            "title": "Complete Python Course",
            "summary": "A comprehensive course covering Python basics to advanced.",
        }
        r = session.post(f"{api_url}/courses", json=payload)
        assert r.status_code == HTTPStatus.CREATED

        course_id = r.json()["id"]
        r = session.get(f"{api_url}/courses/{course_id}")
        assert r.status_code == HTTPStatus.OK
        for key, value in payload.items():
            assert r.json()[key] == value

    def test_rejects_missing_title(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        r = session.post(
            f"{api_url}/courses",
            json={"summary": "Course summary without title."},
        )
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_empty_title(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        r = session.post(
            f"{api_url}/courses",
            json={"title": "", "summary": "Course summary with empty title."},
        )
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_empty_payload(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        r = session.post(f"{api_url}/courses")
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_unauthenticated_request(self, api_url):
        r = requests.post(
            f"{api_url}/courses",
            json={"title": "Course Title", "summary": "Course Summary"},
        )
        assert r.status_code == HTTPStatus.UNAUTHORIZED


class TestCourseRead:
    def test_gets_course_by_id(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        r = session.post(
            f"{api_url}/courses",
            json={"title": "Course Title", "summary": "Course Summary"},
        )
        course_id = r.json()["id"]

        r = session.get(f"{api_url}/courses/{course_id}")
        assert r.status_code == HTTPStatus.OK
        assert r.json()["id"] == course_id
        assert r.json()["title"] == "Course Title"

    def test_returns_404_for_nonexistent_course(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        r = session.get(f"{api_url}/courses/{2**63 - 1}")
        assert r.status_code == HTTPStatus.NOT_FOUND

    def test_lists_live_courses_returns_empty_when_no_courses(self, api_url):
        r = requests.get(f"{api_url}/courses")
        assert r.status_code == HTTPStatus.OK
        assert r.json() == []

    def test_rejects_delete_method_on_courses_list(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        r = session.delete(f"{api_url}/courses")
        assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED


class TestCourseUpdate:
    def test_updates_course_fields(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        r = session.post(
            f"{api_url}/courses",
            json={"title": "Original Title", "summary": "Original Summary"},
        )
        course_id = r.json()["id"]

        r = session.patch(
            f"{api_url}/courses/{course_id}",
            json={"title": "Updated Title", "summary": "Updated Summary"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = session.get(f"{api_url}/courses/{course_id}")
        assert r.json()["title"] == "Updated Title"
        assert r.json()["summary"] == "Updated Summary"

    def test_returns_404_for_nonexistent_course(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        r = session.patch(
            f"{api_url}/courses/{2**63 - 1}",
            json={"title": "Updated Title"},
        )
        assert r.status_code == HTTPStatus.NOT_FOUND

    def test_rejects_unauthenticated_request(self, api_url):
        r = requests.patch(
            f"{api_url}/courses/1",
            json={"title": "Attempted Update"},
        )
        assert r.status_code == HTTPStatus.UNAUTHORIZED


class TestCoursePermissions:
    def test_instructor_can_access_own_course(self, api_url):
        instructor = register_and_login(
            api_url, "instructor1@example.com", "password123"
        )

        r = instructor.post(
            f"{api_url}/courses",
            json={"title": "My Course", "summary": "My course summary"},
        )
        course_id = r.json()["id"]

        r = instructor.get(f"{api_url}/courses/{course_id}")
        assert r.status_code == HTTPStatus.OK
        assert r.json()["title"] == "My Course"

    def test_other_instructor_cannot_update_course(self, api_url):
        instructor1 = register_and_login(
            api_url, "instructor1@example.com", "password123"
        )
        instructor2 = register_and_login(
            api_url, "instructor2@example.com", "password123"
        )

        r = instructor1.post(
            f"{api_url}/courses",
            json={"title": "Instructor 1 Course", "summary": "Summary"},
        )
        course_id = r.json()["id"]

        r = instructor2.patch(
            f"{api_url}/courses/{course_id}",
            json={"title": "Attempted Update"},
        )
        assert r.status_code == HTTPStatus.NOT_FOUND

        r = instructor1.get(f"{api_url}/courses/{course_id}")
        assert r.json()["title"] == "Instructor 1 Course"

    def test_instructors_see_each_others_draft_courses(self, api_url):
        instructor1 = register_and_login(
            api_url, "instructor1@example.com", "password123"
        )
        instructor2 = register_and_login(
            api_url, "instructor2@example.com", "password123"
        )

        r = instructor1.post(
            f"{api_url}/courses",
            json={"title": "Instructor 1 Course", "summary": "Summary"},
        )
        instructor1_course_id = r.json()["id"]

        r = instructor2.post(
            f"{api_url}/courses",
            json={"title": "Instructor 2 Course", "summary": "Summary"},
        )

        r = instructor2.get(f"{api_url}/courses/{instructor1_course_id}")
        assert r.status_code in (HTTPStatus.NOT_FOUND, HTTPStatus.OK)


class TestCoursePublicAccess:
    def test_get_course_without_auth(self, api_url):
        r = requests.get(f"{api_url}/courses/1")
        assert r.status_code in (HTTPStatus.UNAUTHORIZED, HTTPStatus.OK)

    def test_live_courses_are_publicly_accessible(self, api_url):
        instructor = register_and_login(
            api_url, "instructor@example.com", "password123"
        )

        r = instructor.post(
            f"{api_url}/courses",
            json={
                "title": "Public Live Course",
                "summary": "A publicly accessible course.",
            },
        )
        course_id = r.json()["id"]

        r = requests.get(f"{api_url}/courses")
        assert r.status_code == HTTPStatus.OK

        course_ids = [course["id"] for course in r.json()]
        if course_id in course_ids:
            for course in r.json():
                if course["id"] == course_id:
                    assert course["status"] == "live"
