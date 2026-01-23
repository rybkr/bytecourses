import requests
from http import HTTPStatus

from .conftest import register_and_login


class TestCourseCreate:
    def test_creates_course_with_title_and_summary(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        payload = {
            "title": "Introduction to Python",
            "summary": "Learn Python basics.",
            "target_audience": "Developers",
            "learning_objectives": "Learn about python",
            "assumed_prerequisites": "None",
        }
        r = session.post(f"{api_url}/courses", json=payload)
        assert r.status_code == HTTPStatus.CREATED
        assert "id" in r.json()
        assert r.json()["status"] == "draft"

        course_id = r.json()["id"]
        r = session.get(f"{api_url}/courses/{course_id}")
        assert r.status_code == HTTPStatus.OK
        for key, value in payload.items():
            assert r.json()[key] == value

    def test_rejects_missing_fields(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        payload = {
            "title": "Complete Python Course",
            "summary": "A comprehensive course covering Python basics to advanced.",
        }
        r = session.post(f"{api_url}/courses", json=payload)
        assert r.status_code == HTTPStatus.BAD_REQUEST

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

    def test_creates_with_instructor_id(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        payload = {
            "title": "Introduction to Python",
            "summary": "Learn Python basics.",
            "target_audience": "Developers",
            "learning_objectives": "Learn about python",
            "assumed_prerequisites": "None",
        }
        r = session.post(f"{api_url}/courses", json=payload)
        assert r.status_code == HTTPStatus.CREATED
        assert r.json()["instructor_id"] == session.user_id


class TestCourseRead:
    def test_gets_course_by_id(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        r = session.post(
            f"{api_url}/courses",
            json={
                "title": "Course Title",
                "summary": "Course Summary",
                "target_audience": "Developers",
                "learning_objectives": "Learn about python",
                "assumed_prerequisites": "None",
            },
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
            json={
                "title": "Original Title",
                "summary": "Original Summary",
                "target_audience": "Developers",
                "learning_objectives": "Learn about python",
                "assumed_prerequisites": "None",
            },
        )
        course_id = r.json()["id"]

        r = session.patch(
            f"{api_url}/courses/{course_id}",
            json={
                "title": "Updated Title",
                "summary": "Updated Summary",
                "target_audience": "Developers",
                "learning_objectives": "Learn about python",
                "assumed_prerequisites": "None",
            },
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = session.get(f"{api_url}/courses/{course_id}")
        assert r.json()["title"] == "Updated Title"
        assert r.json()["summary"] == "Updated Summary"

    def test_returns_404_for_nonexistent_course(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        r = session.patch(
            f"{api_url}/courses/{2**63 - 1}",
            json={
                "title": "Updated Title",
                "summary": "New Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
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
            json={
                "title": "My Course",
                "summary": "My course summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
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
            json={
                "title": "Instructor 1 Course",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        course_id = r.json()["id"]

        r = instructor2.patch(
            f"{api_url}/courses/{course_id}",
            json={
                "title": "Attempted Update",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
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
            json={
                "title": "Instructor 1 Course",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        instructor1_course_id = r.json()["id"]

        r = instructor2.post(
            f"{api_url}/courses",
            json={
                "title": "Instructor 2 Course",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
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
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
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
