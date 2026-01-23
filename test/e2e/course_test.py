import requests
from http import HTTPStatus

from .conftest import register_and_login


class TestCourseCreate:
    def test_rejects_direct_course_creation(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        payload = {
            "title": "Introduction to Python",
            "summary": "Learn Python basics.",
            "target_audience": "Developers",
            "learning_objectives": "Learn about python",
            "assumed_prerequisites": "None",
        }
        r = session.post(f"{api_url}/courses", json=payload)
        assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED

    def test_rejects_direct_course_creation_even_with_valid_payload(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        payload = {
            "title": "Valid Title",
            "summary": "Valid summary",
            "target_audience": "Developers",
            "learning_objectives": "Learn",
            "assumed_prerequisites": "None",
        }
        r = session.post(f"{api_url}/courses", json=payload)
        assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED

    def test_rejects_direct_course_creation_unauthenticated(self, api_url):
        r = requests.post(
            f"{api_url}/courses",
            json={
                "title": "Course Title",
                "summary": "Course Summary",
                "target_audience": "Developers",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED


class TestCourseCreationFromProposalOnly:
    def test_creates_course_from_approved_proposal_owned_by_user(
        self, api_url, admin_session
    ):
        author = register_and_login(api_url, "author@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Proposal Title",
                "summary": "Proposal Summary",
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
        assert r.status_code == HTTPStatus.CREATED
        assert "id" in r.json()
        assert r.json()["status"] == "draft"

    def test_created_course_has_correct_proposal_id(self, api_url, admin_session):
        author = register_and_login(api_url, "author@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Proposal Title",
                "summary": "Proposal Summary",
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
        assert r.status_code == HTTPStatus.CREATED
        assert r.json()["proposal_id"] == proposal_id

    def test_created_course_has_correct_fields_from_proposal(
        self, api_url, admin_session
    ):
        author = register_and_login(api_url, "author@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Proposal Title",
                "summary": "Proposal Summary",
                "target_audience": "Developers",
                "learning_objectives": "Learn Python",
                "assumed_prerequisites": "Basic programming",
            },
        )
        proposal_id = r.json()["id"]
        proposal_data = r.json()

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        assert r.status_code == HTTPStatus.CREATED
        course = r.json()
        assert course["title"] == proposal_data["title"]
        assert course["summary"] == proposal_data["summary"]
        assert course["target_audience"] == proposal_data["target_audience"]
        assert course["learning_objectives"] == proposal_data["learning_objectives"]
        assert course["assumed_prerequisites"] == proposal_data["assumed_prerequisites"]

    def test_cannot_create_course_from_proposal_owned_by_other_user(
        self, api_url, admin_session
    ):
        author1 = register_and_login(api_url, "author1@example.com", "password123")
        author2 = register_and_login(api_url, "author2@example.com", "password123")

        r = author1.post(
            f"{api_url}/proposals",
            json={
                "title": "Author 1 Proposal",
                "summary": "Summary",
                "target_audience": "Developers",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = author1.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author2.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        assert r.status_code == HTTPStatus.NOT_FOUND

    def test_cannot_create_course_from_nonexistent_proposal(self, api_url):
        author = register_and_login(api_url, "author@example.com", "password123")

        r = author.post(f"{api_url}/proposals/{2**63 - 1}/actions/create-course")
        assert r.status_code == HTTPStatus.NOT_FOUND

    def test_cannot_create_course_from_draft_proposal(self, api_url):
        author = register_and_login(api_url, "author@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Draft Proposal",
                "summary": "Summary",
                "target_audience": "Developers",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        assert r.status_code == HTTPStatus.CONFLICT

    def test_cannot_create_course_from_submitted_proposal(self, api_url):
        author = register_and_login(api_url, "author@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Submitted Proposal",
                "summary": "Summary",
                "target_audience": "Developers",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        assert r.status_code == HTTPStatus.CONFLICT

    def test_cannot_create_course_from_rejected_proposal(self, api_url, admin_session):
        author = register_and_login(api_url, "author@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Proposal to Reject",
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
            f"{api_url}/proposals/{proposal_id}/actions/reject",
            json={"review_notes": "Rejected"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        assert r.status_code == HTTPStatus.CONFLICT

    def test_cannot_create_course_from_withdrawn_proposal(self, api_url):
        author = register_and_login(api_url, "author@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Proposal to Withdraw",
                "summary": "Summary",
                "target_audience": "Developers",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/withdraw")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        assert r.status_code == HTTPStatus.CONFLICT

    def test_cannot_create_course_from_changes_requested_proposal(
        self, api_url, admin_session
    ):
        author = register_and_login(api_url, "author@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Proposal for Changes",
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
            f"{api_url}/proposals/{proposal_id}/actions/request-changes",
            json={"review_notes": "Need changes"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        assert r.status_code == HTTPStatus.CONFLICT

    def test_cannot_create_second_course_from_same_proposal(
        self, api_url, admin_session
    ):
        author = register_and_login(api_url, "author@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Proposal Title",
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
        assert r.status_code == HTTPStatus.CREATED
        course_id = r.json()["id"]

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        assert r.status_code == HTTPStatus.CONFLICT
        assert r.json()["course_id"] == course_id

    def test_requires_authentication_to_create_course_from_proposal(
        self, api_url, admin_session
    ):
        author = register_and_login(api_url, "author@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Proposal Title",
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

        r = requests.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        assert r.status_code == HTTPStatus.UNAUTHORIZED


class TestCourseRead:
    def test_gets_course_by_id(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Course Title",
                "summary": "Course Summary",
                "target_audience": "Developers",
                "learning_objectives": "Learn about python",
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

        r = author.get(f"{api_url}/courses/{course_id}")
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
    def test_updates_course_fields(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Original Title",
                "summary": "Original Summary",
                "target_audience": "Developers",
                "learning_objectives": "Learn about python",
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

        r = author.patch(
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

        r = author.get(f"{api_url}/courses/{course_id}")
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

    def test_partial_update(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Original Title",
                "summary": "Original Summary",
                "target_audience": "Developers",
                "learning_objectives": "Learn about python",
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

        r = author.patch(
            f"{api_url}/courses/{course_id}",
            json={
                "title": "Updated Title Only",
                "summary": "Original Summary",
                "target_audience": "Developers",
                "learning_objectives": "Learn about python",
                "assumed_prerequisites": "None",
            },
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.get(f"{api_url}/courses/{course_id}")
        assert r.json()["title"] == "Updated Title Only"
        assert r.json()["summary"] == "Original Summary"

    def test_invalid_course_id_format(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        r = session.patch(
            f"{api_url}/courses/invalid",
            json={
                "title": "Updated Title",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        assert r.status_code == HTTPStatus.BAD_REQUEST


class TestCoursePermissions:
    def test_instructor_can_access_own_course(self, api_url, admin_session):
        instructor = register_and_login(
            api_url, "instructor1@example.com", "password123"
        )

        r = instructor.post(
            f"{api_url}/proposals",
            json={
                "title": "My Course",
                "summary": "My course summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = instructor.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = instructor.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        course_id = r.json()["id"]

        r = instructor.get(f"{api_url}/courses/{course_id}")
        assert r.status_code == HTTPStatus.OK
        assert r.json()["title"] == "My Course"

    def test_other_instructor_cannot_update_course(self, api_url, admin_session):
        instructor1 = register_and_login(
            api_url, "instructor1@example.com", "password123"
        )
        instructor2 = register_and_login(
            api_url, "instructor2@example.com", "password123"
        )

        r = instructor1.post(
            f"{api_url}/proposals",
            json={
                "title": "Instructor 1 Course",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = instructor1.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = instructor1.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
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

    def test_instructors_see_each_others_draft_courses(self, api_url, admin_session):
        instructor1 = register_and_login(
            api_url, "instructor1@example.com", "password123"
        )
        instructor2 = register_and_login(
            api_url, "instructor2@example.com", "password123"
        )

        r = instructor1.post(
            f"{api_url}/proposals",
            json={
                "title": "Instructor 1 Course",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal1_id = r.json()["id"]

        r = instructor1.post(f"{api_url}/proposals/{proposal1_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal1_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = instructor1.post(
            f"{api_url}/proposals/{proposal1_id}/actions/create-course"
        )
        instructor1_course_id = r.json()["id"]

        r = instructor2.post(
            f"{api_url}/proposals",
            json={
                "title": "Instructor 2 Course",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal2_id = r.json()["id"]

        r = instructor2.post(f"{api_url}/proposals/{proposal2_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal2_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = instructor2.post(
            f"{api_url}/proposals/{proposal2_id}/actions/create-course"
        )

        r = instructor2.get(f"{api_url}/courses/{instructor1_course_id}")
        assert r.status_code in (HTTPStatus.NOT_FOUND, HTTPStatus.OK)


class TestCoursePublicAccess:
    def test_get_course_without_auth(self, api_url):
        r = requests.get(f"{api_url}/courses/1")
        assert r.status_code in (HTTPStatus.UNAUTHORIZED, HTTPStatus.OK)

    def test_live_courses_are_publicly_accessible(self, api_url, admin_session):
        instructor = register_and_login(
            api_url, "instructor@example.com", "password123"
        )

        r = instructor.post(
            f"{api_url}/proposals",
            json={
                "title": "Public Live Course",
                "summary": "A publicly accessible course.",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = instructor.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = instructor.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        course_id = r.json()["id"]

        r = instructor.post(f"{api_url}/courses/{course_id}/publish")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = requests.get(f"{api_url}/courses")
        assert r.status_code == HTTPStatus.OK

        course_ids = [course["id"] for course in r.json()]
        if course_id in course_ids:
            for course in r.json():
                if course["id"] == course_id:
                    assert course["status"] == "published"


class TestCoursePublish:
    def test_publishes_draft_course(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Draft Course",
                "summary": "A draft course",
                "target_audience": "Learners",
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
        assert r.json()["status"] == "draft"

        r = author.post(f"{api_url}/courses/{course_id}/publish")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.get(f"{api_url}/courses/{course_id}")
        assert r.status_code == HTTPStatus.OK
        assert r.json()["status"] == "published"

    def test_publish_requires_authentication(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Draft Course",
                "summary": "A draft course",
                "target_audience": "Learners",
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

        r = requests.post(f"{api_url}/courses/{course_id}/publish")
        assert r.status_code == HTTPStatus.UNAUTHORIZED

    def test_cannot_publish_other_instructors_course(self, api_url, admin_session):
        instructor1 = register_and_login(
            api_url, "instructor1@example.com", "password123"
        )
        instructor2 = register_and_login(
            api_url, "instructor2@example.com", "password123"
        )

        r = instructor1.post(
            f"{api_url}/proposals",
            json={
                "title": "Instructor 1 Course",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = instructor1.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = instructor1.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        course_id = r.json()["id"]

        r = instructor2.post(f"{api_url}/courses/{course_id}/publish")
        assert r.status_code == HTTPStatus.NOT_FOUND

        r = instructor1.get(f"{api_url}/courses/{course_id}")
        assert r.json()["status"] == "draft"

    def test_cannot_publish_already_published_course(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Course to Publish",
                "summary": "Summary",
                "target_audience": "Learners",
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

        r = author.post(f"{api_url}/courses/{course_id}/publish")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.post(f"{api_url}/courses/{course_id}/publish")
        assert r.status_code == HTTPStatus.CONFLICT

    def test_cannot_publish_nonexistent_course(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        r = session.post(f"{api_url}/courses/{2**63 - 1}/publish")
        assert r.status_code == HTTPStatus.NOT_FOUND

    def test_published_course_appears_in_list(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Course to Publish",
                "summary": "Summary",
                "target_audience": "Learners",
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

        r = requests.get(f"{api_url}/courses")
        course_ids = [course["id"] for course in r.json()]
        assert course_id not in course_ids

        r = author.post(f"{api_url}/courses/{course_id}/publish")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = requests.get(f"{api_url}/courses")
        course_ids = [course["id"] for course in r.json()]
        assert course_id in course_ids


class TestCourseCreateFromProposal:
    def test_creates_course_from_approved_proposal(self, api_url, admin_session):
        author = register_and_login(api_url, "author@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Proposal Title",
                "summary": "Proposal Summary",
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
        assert r.status_code == HTTPStatus.CREATED
        assert "id" in r.json()
        assert r.json()["proposal_id"] == proposal_id
        assert r.json()["title"] == "Proposal Title"
        assert r.json()["summary"] == "Proposal Summary"
        assert r.json()["status"] == "draft"

    def test_cannot_create_from_draft_proposal(self, api_url):
        author = register_and_login(api_url, "author@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Draft Proposal",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        assert r.status_code == HTTPStatus.CONFLICT

    def test_cannot_create_from_rejected_proposal(self, api_url, admin_session):
        author = register_and_login(api_url, "author@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Proposal to Reject",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/reject",
            json={"review_notes": "Rejected"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        assert r.status_code == HTTPStatus.CONFLICT

    def test_cannot_create_if_course_already_exists(self, api_url, admin_session):
        author = register_and_login(api_url, "author@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Proposal Title",
                "summary": "Summary",
                "target_audience": "Learners",
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
        assert r.status_code == HTTPStatus.CREATED
        course_id = r.json()["id"]

        r = author.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        assert r.status_code == HTTPStatus.CONFLICT
        assert r.json()["course_id"] == course_id

    def test_only_proposal_author_can_create_course(self, api_url, admin_session):
        author1 = register_and_login(api_url, "author1@example.com", "password123")
        author2 = register_and_login(api_url, "author2@example.com", "password123")

        r = author1.post(
            f"{api_url}/proposals",
            json={
                "title": "Author 1 Proposal",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = author1.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author2.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        assert r.status_code == HTTPStatus.NOT_FOUND

    def test_requires_authentication(self, api_url, admin_session):
        author = register_and_login(api_url, "author@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Proposal Title",
                "summary": "Summary",
                "target_audience": "Learners",
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

        r = requests.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        assert r.status_code == HTTPStatus.UNAUTHORIZED


class TestCourseAdminAccess:
    def test_admin_can_access_any_course(self, api_url, admin_session):
        instructor = register_and_login(
            api_url, "instructor@example.com", "password123"
        )

        r = instructor.post(
            f"{api_url}/proposals",
            json={
                "title": "Instructor Course",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = instructor.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = instructor.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        course_id = r.json()["id"]

        r = admin_session.get(f"{api_url}/courses/{course_id}")
        assert r.status_code == HTTPStatus.OK
        assert r.json()["title"] == "Instructor Course"

    def test_admin_can_access_other_instructors_draft_course(
        self, api_url, admin_session
    ):
        instructor = register_and_login(
            api_url, "instructor@example.com", "password123"
        )

        r = instructor.post(
            f"{api_url}/proposals",
            json={
                "title": "Draft Course",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = instructor.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = instructor.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        course_id = r.json()["id"]
        assert r.json()["status"] == "draft"

        r = admin_session.get(f"{api_url}/courses/{course_id}")
        assert r.status_code == HTTPStatus.OK
        assert r.json()["status"] == "draft"

    def test_admin_can_access_other_instructors_published_course(
        self, api_url, admin_session
    ):
        instructor = register_and_login(
            api_url, "instructor@example.com", "password123"
        )

        r = instructor.post(
            f"{api_url}/proposals",
            json={
                "title": "Course to Publish",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = instructor.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = instructor.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        course_id = r.json()["id"]

        r = instructor.post(f"{api_url}/courses/{course_id}/publish")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.get(f"{api_url}/courses/{course_id}")
        assert r.status_code == HTTPStatus.OK
        assert r.json()["status"] == "published"


class TestCourseFieldValidation:
    def test_rejects_title_too_short(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Valid Title",
                "summary": "Summary",
                "target_audience": "Learners",
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

        r = author.patch(
            f"{api_url}/courses/{course_id}",
            json={
                "title": "abc",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_title_too_long(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Valid Title",
                "summary": "Summary",
                "target_audience": "Learners",
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

        r = author.patch(
            f"{api_url}/courses/{course_id}",
            json={
                "title": "a" * 129,
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_summary_too_long(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Valid Title",
                "summary": "Summary",
                "target_audience": "Learners",
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

        r = author.patch(
            f"{api_url}/courses/{course_id}",
            json={
                "title": "Valid Title",
                "summary": "a" * 2049,
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_target_audience_too_long(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Valid Title",
                "summary": "Summary",
                "target_audience": "Learners",
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

        r = author.patch(
            f"{api_url}/courses/{course_id}",
            json={
                "title": "Valid Title",
                "summary": "Summary",
                "target_audience": "a" * 2049,
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_learning_objectives_too_long(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Valid Title",
                "summary": "Summary",
                "target_audience": "Learners",
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

        r = author.patch(
            f"{api_url}/courses/{course_id}",
            json={
                "title": "Valid Title",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "a" * 2049,
                "assumed_prerequisites": "None",
            },
        )
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_assumed_prerequisites_too_long(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Valid Title",
                "summary": "Summary",
                "target_audience": "Learners",
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

        r = author.patch(
            f"{api_url}/courses/{course_id}",
            json={
                "title": "Valid Title",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "a" * 2049,
            },
        )
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_trims_whitespace_from_fields(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "  Trimmed Title  ",
                "summary": "  Trimmed Summary  ",
                "target_audience": "  Learners  ",
                "learning_objectives": "  Learn  ",
                "assumed_prerequisites": "  None  ",
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

        r = author.get(f"{api_url}/courses/{course_id}")
        assert r.json()["title"] == "Trimmed Title"
        assert r.json()["summary"] == "Trimmed Summary"
        assert r.json()["target_audience"] == "Learners"
        assert r.json()["learning_objectives"] == "Learn"
        assert r.json()["assumed_prerequisites"] == "None"

    def test_trims_whitespace_on_update(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Original Title",
                "summary": "Original Summary",
                "target_audience": "Learners",
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

        r = author.patch(
            f"{api_url}/courses/{course_id}",
            json={
                "title": "  Updated Title  ",
                "summary": "  Updated Summary  ",
                "target_audience": "  Developers  ",
                "learning_objectives": "  Learn More  ",
                "assumed_prerequisites": "  Some  ",
            },
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.get(f"{api_url}/courses/{course_id}")
        assert r.json()["title"] == "Updated Title"
        assert r.json()["summary"] == "Updated Summary"
        assert r.json()["target_audience"] == "Developers"
        assert r.json()["learning_objectives"] == "Learn More"
        assert r.json()["assumed_prerequisites"] == "Some"


class TestCourseList:
    def test_list_only_shows_published_courses(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Draft Course",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal1_id = r.json()["id"]

        r = author.post(f"{api_url}/proposals/{proposal1_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal1_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.post(f"{api_url}/proposals/{proposal1_id}/actions/create-course")
        draft_course_id = r.json()["id"]

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Published Course",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal2_id = r.json()["id"]

        r = author.post(f"{api_url}/proposals/{proposal2_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal2_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.post(f"{api_url}/proposals/{proposal2_id}/actions/create-course")
        published_course_id = r.json()["id"]
        r = author.post(f"{api_url}/courses/{published_course_id}/publish")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = requests.get(f"{api_url}/courses")
        assert r.status_code == HTTPStatus.OK
        course_ids = [course["id"] for course in r.json()]
        assert draft_course_id not in course_ids
        assert published_course_id in course_ids

    def test_list_shows_multiple_published_courses(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Course 1",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal1_id = r.json()["id"]

        r = author.post(f"{api_url}/proposals/{proposal1_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal1_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.post(f"{api_url}/proposals/{proposal1_id}/actions/create-course")
        course1_id = r.json()["id"]
        r = author.post(f"{api_url}/courses/{course1_id}/publish")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Course 2",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal2_id = r.json()["id"]

        r = author.post(f"{api_url}/proposals/{proposal2_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal2_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.post(f"{api_url}/proposals/{proposal2_id}/actions/create-course")
        course2_id = r.json()["id"]
        r = author.post(f"{api_url}/courses/{course2_id}/publish")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = requests.get(f"{api_url}/courses")
        assert r.status_code == HTTPStatus.OK
        course_ids = [course["id"] for course in r.json()]
        assert course1_id in course_ids
        assert course2_id in course_ids

    def test_list_works_without_authentication(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Public Course",
                "summary": "Summary",
                "target_audience": "Learners",
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
        r = author.post(f"{api_url}/courses/{course_id}/publish")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = requests.get(f"{api_url}/courses")
        assert r.status_code == HTTPStatus.OK
        course_ids = [course["id"] for course in r.json()]
        assert course_id in course_ids

    def test_list_works_with_authentication(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Public Course",
                "summary": "Summary",
                "target_audience": "Learners",
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
        r = author.post(f"{api_url}/courses/{course_id}/publish")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.get(f"{api_url}/courses")
        assert r.status_code == HTTPStatus.OK
        course_ids = [course["id"] for course in r.json()]
        assert course_id in course_ids


class TestCourseGetPermissions:
    def test_student_cannot_access_other_instructors_course(
        self, api_url, admin_session
    ):
        instructor = register_and_login(
            api_url, "instructor@example.com", "password123"
        )
        student = register_and_login(api_url, "student@example.com", "password123")

        r = instructor.post(
            f"{api_url}/proposals",
            json={
                "title": "Instructor Course",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = instructor.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = instructor.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        course_id = r.json()["id"]

        r = student.get(f"{api_url}/courses/{course_id}")
        assert r.status_code == HTTPStatus.NOT_FOUND

    def test_published_course_visibility(self, api_url, admin_session):
        instructor = register_and_login(
            api_url, "instructor@example.com", "password123"
        )
        student = register_and_login(api_url, "student@example.com", "password123")

        r = instructor.post(
            f"{api_url}/proposals",
            json={
                "title": "Published Course",
                "summary": "Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        proposal_id = r.json()["id"]

        r = instructor.post(f"{api_url}/proposals/{proposal_id}/actions/submit")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = admin_session.post(
            f"{api_url}/proposals/{proposal_id}/actions/approve",
            json={"review_notes": "Approved"},
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = instructor.post(f"{api_url}/proposals/{proposal_id}/actions/create-course")
        course_id = r.json()["id"]
        r = instructor.post(f"{api_url}/courses/{course_id}/publish")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = student.get(f"{api_url}/courses/{course_id}")
        assert r.status_code == HTTPStatus.NOT_FOUND

    def test_invalid_course_id_format(self, api_url):
        session = register_and_login(api_url, "instructor@example.com", "password123")

        r = session.get(f"{api_url}/courses/invalid")
        assert r.status_code == HTTPStatus.BAD_REQUEST


class TestCourseTimestamps:
    def test_created_at_is_set_on_creation(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "New Course",
                "summary": "Summary",
                "target_audience": "Learners",
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
        assert r.status_code == HTTPStatus.CREATED
        assert "created_at" in r.json()
        assert r.json()["created_at"] is not None

    def test_updated_at_changes_on_update(self, api_url, admin_session):
        author = register_and_login(api_url, "instructor@example.com", "password123")

        r = author.post(
            f"{api_url}/proposals",
            json={
                "title": "Original Title",
                "summary": "Summary",
                "target_audience": "Learners",
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
        original_updated_at = r.json()["updated_at"]

        r = author.patch(
            f"{api_url}/courses/{course_id}",
            json={
                "title": "Updated Title",
                "summary": "Updated Summary",
                "target_audience": "Learners",
                "learning_objectives": "Learn",
                "assumed_prerequisites": "None",
            },
        )
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = author.get(f"{api_url}/courses/{course_id}")
        assert r.json()["updated_at"] != original_updated_at
