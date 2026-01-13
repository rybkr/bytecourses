import requests
from http import HTTPStatus


def test_create_course(go_server):
    s = requests.Session()

    register_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=register_payload)
    assert r.status_code == HTTPStatus.CREATED

    login_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    course_payload: dict[str, str] = {
        "title": "Introduction to Python",
        "summary": "Learn the basics of Python programming.",
    }
    r = s.post(f"{go_server}/courses", json=course_payload)
    assert r.status_code == HTTPStatus.CREATED
    assert "id" in r.json()
    assert "title" in r.json()
    assert r.json()["title"] == "Introduction to Python"
    assert r.json()["status"] == "draft"

    course_payload = {
        "title": "Advanced JavaScript",
        "summary": "Master advanced JavaScript concepts.",
    }
    r = s.post(f"{go_server}/courses", json=course_payload)
    assert r.status_code == HTTPStatus.CREATED
    assert "id" in r.json()


def test_get_course_by_id(go_server):
    s = requests.Session()

    register_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=register_payload)
    assert r.status_code == HTTPStatus.CREATED

    login_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    course_payload: dict[str, str] = {
        "title": "Course Title",
        "summary": "Course Summary",
    }
    r = s.post(f"{go_server}/courses", json=course_payload)
    assert r.status_code == HTTPStatus.CREATED
    assert "id" in r.json()
    course_id = r.json()["id"]

    r = s.get(f"{go_server}/courses/{course_id}")
    assert r.status_code == HTTPStatus.OK
    assert r.json()["id"] == course_id
    assert r.json()["title"] == "Course Title"
    assert r.json()["summary"] == "Course Summary"


def test_get_course_nonexistent(go_server):
    s = requests.Session()

    register_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=register_payload)
    assert r.status_code == HTTPStatus.CREATED

    login_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    r = s.get(f"{go_server}/courses/{2**63 - 1}")
    assert r.status_code == HTTPStatus.NOT_FOUND


"""
def test_list_live_courses_empty(go_server):
    r = requests.get(f"{go_server}/courses")
    assert r.status_code == HTTPStatus.OK

    data = r.json()
    assert len(data) == 0


def test_list_live_courses(go_server):
    s = requests.Session()

    register_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=register_payload)
    assert r.status_code == HTTPStatus.CREATED

    login_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    course_payload: dict[str, str] = {
        "title": "Draft Course",
        "summary": "This is a draft course.",
    }
    r = s.post(f"{go_server}/courses", json=course_payload)
    assert r.status_code == HTTPStatus.CREATED
    draft_id = r.json()["id"]

    course_payload = {
        "title": "Live Course 1",
        "summary": "This is a live course.",
    }
    r = s.post(f"{go_server}/courses", json=course_payload)
    assert r.status_code == HTTPStatus.CREATED
    live_id_1 = r.json()["id"]

    course_payload = {
        "title": "Live Course 2",
        "summary": "This is another live course.",
    }
    r = s.post(f"{go_server}/courses", json=course_payload)
    assert r.status_code == HTTPStatus.CREATED
    live_id_2 = r.json()["id"]

    r = requests.get(f"{go_server}/courses")
    assert r.status_code == HTTPStatus.OK

    data = r.json()
    live_titles = [course["title"] for course in data]
    live_ids = [course["id"] for course in data]

    assert draft_id not in live_ids
    assert live_id_1 in live_ids or live_id_2 in live_ids
    for course in data:
        assert course["status"] == "live"


def test_courses_invalid_method(go_server):
    r = requests.delete(f"{go_server}/courses")
    assert r.status_code in (HTTPStatus.UNAUTHORIZED, HTTPStatus.METHOD_NOT_ALLOWED)

    s = requests.Session()

    register_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=register_payload)
    assert r.status_code == HTTPStatus.CREATED

    login_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    r = s.delete(f"{go_server}/courses")
    assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED


def test_create_course_no_title(go_server):
    s = requests.Session()

    register_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=register_payload)
    assert r.status_code == HTTPStatus.CREATED

    login_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    course_payload: dict[str, str] = {
        "summary": "Course summary without title.",
    }
    r = s.post(f"{go_server}/courses", json=course_payload)
    assert r.status_code == HTTPStatus.BAD_REQUEST


def test_create_course_empty_title(go_server):
    s = requests.Session()

    register_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=register_payload)
    assert r.status_code == HTTPStatus.CREATED

    login_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    course_payload: dict[str, str] = {
        "title": "",
        "summary": "Course summary with empty title.",
    }
    r = s.post(f"{go_server}/courses", json=course_payload)
    assert r.status_code == HTTPStatus.BAD_REQUEST


def test_create_course_no_payload(go_server):
    s = requests.Session()

    register_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=register_payload)
    assert r.status_code == HTTPStatus.CREATED

    login_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    r = s.post(f"{go_server}/courses")
    assert r.status_code == HTTPStatus.BAD_REQUEST


def test_create_course_unauthorized(go_server):
    course_payload: dict[str, str] = {
        "title": "Course Title",
        "summary": "Course Summary",
    }
    r = requests.post(f"{go_server}/courses", json=course_payload)
    assert r.status_code == HTTPStatus.UNAUTHORIZED


def test_get_course_unauthorized(go_server):
    r = requests.get(f"{go_server}/courses/1")
    assert r.status_code in (HTTPStatus.UNAUTHORIZED, HTTPStatus.OK)


def test_course_instructor_isolation(go_server):
    s = requests.Session()
    t = requests.Session()

    s_register_payload: dict[str, str] = {
        "email": "instructor1@example.com",
        "password": "password123",
    }
    t_register_payload: dict[str, str] = {
        "email": "instructor2@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=s_register_payload)
    assert r.status_code == HTTPStatus.CREATED
    r = t.post(f"{go_server}/register", json=t_register_payload)
    assert r.status_code == HTTPStatus.CREATED

    s_login_payload: dict[str, str] = {
        "email": "instructor1@example.com",
        "password": "password123",
    }
    t_login_payload: dict[str, str] = {
        "email": "instructor2@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/login", json=s_login_payload)
    assert r.status_code == HTTPStatus.OK
    r = t.post(f"{go_server}/login", json=t_login_payload)
    assert r.status_code == HTTPStatus.OK

    s_course_payload: dict[str, str] = {
        "title": "Instructor 1 Course",
        "summary": "Course by instructor 1",
    }
    t_course_payload: dict[str, str] = {
        "title": "Instructor 2 Course",
        "summary": "Course by instructor 2",
    }

    r = s.post(f"{go_server}/courses", json=s_course_payload)
    assert r.status_code == HTTPStatus.CREATED
    s_course_id = r.json()["id"]

    r = t.post(f"{go_server}/courses", json=t_course_payload)
    assert r.status_code == HTTPStatus.CREATED
    t_course_id = r.json()["id"]

    r = s.get(f"{go_server}/courses/{s_course_id}")
    assert r.status_code == HTTPStatus.OK
    assert r.json()["title"] == "Instructor 1 Course"

    r = t.get(f"{go_server}/courses/{t_course_id}")
    assert r.status_code == HTTPStatus.OK
    assert r.json()["title"] == "Instructor 2 Course"

    r = s.get(f"{go_server}/courses/{t_course_id}")
    assert r.status_code in (HTTPStatus.NOT_FOUND, HTTPStatus.OK)

    r = t.get(f"{go_server}/courses/{s_course_id}")
    assert r.status_code in (HTTPStatus.NOT_FOUND, HTTPStatus.OK)


def test_list_live_courses_public_access(go_server):
    s = requests.Session()

    register_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=register_payload)
    assert r.status_code == HTTPStatus.CREATED

    login_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    course_payload: dict[str, str] = {
        "title": "Public Live Course",
        "summary": "A course that should be publicly accessible.",
    }
    r = s.post(f"{go_server}/courses", json=course_payload)
    assert r.status_code == HTTPStatus.CREATED
    course_id = r.json()["id"]

    r = requests.get(f"{go_server}/courses")
    assert r.status_code == HTTPStatus.OK

    data = r.json()
    course_ids = [course["id"] for course in data]

    if course_id in course_ids:
        for course in data:
            if course["id"] == course_id:
                assert course["status"] == "live"


def test_create_course_rich_content(go_server):
    s = requests.Session()

    register_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=register_payload)
    assert r.status_code == HTTPStatus.CREATED

    login_payload: dict[str, str] = {
        "email": "instructor@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    course_payload: dict[str, str] = {
        "title": "Complete Python Course",
        "summary": "A comprehensive course covering Python from basics to advanced topics.",
    }
    r = s.post(f"{go_server}/courses", json=course_payload)
    assert r.status_code == HTTPStatus.CREATED
    assert "id" in r.json()
    course_id = r.json()["id"]

    r = s.get(f"{go_server}/courses/{course_id}")
    assert r.status_code == HTTPStatus.OK

    for key, value in course_payload.items():
        assert key in r.json()
        assert r.json()[key] == value
        """
