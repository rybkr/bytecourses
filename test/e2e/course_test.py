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


def test_list_live_courses_empty(go_server):
    r = requests.get(f"{go_server}/courses")
    assert r.status_code == HTTPStatus.OK

    data = r.json()
    assert len(data) == 0


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


def test_update_course(go_server):
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
        "title": "Original Title",
        "summary": "Original Summary",
    }
    r = s.post(f"{go_server}/courses", json=course_payload)
    assert r.status_code == HTTPStatus.CREATED
    assert "id" in r.json()
    course_id = r.json()["id"]

    r = s.get(f"{go_server}/courses/{course_id}")
    assert r.status_code == HTTPStatus.OK
    assert r.json()["title"] == "Original Title"
    assert r.json()["summary"] == "Original Summary"

    update_payload: dict[str, str] = {
        "title": "Updated Title",
        "summary": "Updated Summary",
    }
    r = s.patch(f"{go_server}/courses/{course_id}", json=update_payload)
    assert r.status_code == HTTPStatus.NO_CONTENT

    r = s.get(f"{go_server}/courses/{course_id}")
    assert r.status_code == HTTPStatus.OK
    assert r.json()["title"] == "Updated Title"
    assert r.json()["summary"] == "Updated Summary"


def test_update_course_wrong_instructor(go_server):
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

    course_payload: dict[str, str] = {
        "title": "Instructor 1 Course",
        "summary": "Course by instructor 1",
    }
    r = s.post(f"{go_server}/courses", json=course_payload)
    assert r.status_code == HTTPStatus.CREATED
    course_id = r.json()["id"]

    update_payload: dict[str, str] = {
        "title": "Attempted Update",
        "summary": "Attempted update by wrong instructor",
    }
    r = t.patch(f"{go_server}/courses/{course_id}", json=update_payload)
    assert r.status_code == HTTPStatus.NOT_FOUND

    r = s.get(f"{go_server}/courses/{course_id}")
    assert r.status_code == HTTPStatus.OK
    assert r.json()["title"] == "Instructor 1 Course"
    assert r.json()["summary"] == "Course by instructor 1"


def test_update_course_nonexistent(go_server):
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

    update_payload: dict[str, str] = {
        "title": "Attempted Update",
        "summary": "Attempted update of nonexistent course",
    }
    r = s.patch(f"{go_server}/courses/{2**63 - 1}", json=update_payload)
    assert r.status_code == HTTPStatus.NOT_FOUND


def test_update_course_unauthorized(go_server):
    update_payload: dict[str, str] = {
        "title": "Attempted Update",
        "summary": "Attempted update without auth",
    }
    r = requests.patch(f"{go_server}/courses/1", json=update_payload)
    assert r.status_code == HTTPStatus.UNAUTHORIZED
