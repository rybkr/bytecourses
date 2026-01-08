import pytest
import requests
from http import HTTPStatus

go_server: str = "http://localhost:8080/api"


def test_create_proposal(go_server):
    s = requests.Session()

    register_payload: dict[str, str] = {
        "email": "usr@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=register_payload)
    assert r.status_code == HTTPStatus.CREATED

    login_payload: dict[str, str] = {
        "email": "usr@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    proposal_payload: dict[str, str] = {
        "title": "Some Course Title",
        "summary": "A summary of some course.",
    }
    r = s.post(f"{go_server}/proposals", json=proposal_payload)
    assert r.status_code == HTTPStatus.CREATED

    proposal_payload = {
        "title": "Another Course Title",
        "summary": "A summary of another course.",
    }
    r = s.post(f"{go_server}/proposals", json=proposal_payload)
    assert r.status_code == HTTPStatus.CREATED


def test_get_proposals(go_server):
    s = requests.Session()

    login_payload: dict[str, str] = {
        "email": "usr@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=login_payload)
    assert r.status_code == HTTPStatus.CREATED
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    proposal_payload: dict[str, str] = {
        "title": "Some Course Title",
        "summary": "A summary of some course.",
    }
    r = s.post(f"{go_server}/proposals", json=proposal_payload)
    assert r.status_code == HTTPStatus.CREATED

    proposal_payload = {
        "title": "Another Course Title",
        "summary": "A summary of another course.",
    }
    r = s.post(f"{go_server}/proposals", json=proposal_payload)
    assert r.status_code == HTTPStatus.CREATED

    r = s.get(f"{go_server}/proposals")
    assert r.status_code == HTTPStatus.OK

    data = sorted(r.json(), key=lambda p: p["id"])
    assert len(data) == 2
    assert "title" in data[0] and "summary" in data[0]
    assert "title" in data[1] and "summary" in data[1]

    assert data[0]["title"] == "Some Course Title"
    assert data[0]["summary"] == "A summary of some course."
    assert data[1]["title"] == "Another Course Title"
    assert data[1]["summary"] == "A summary of another course."


def test_get_proposals_empty(go_server):
    s = requests.Session()

    login_payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=login_payload)
    assert r.status_code == HTTPStatus.CREATED
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    r = s.get(f"{go_server}/proposals")
    assert r.status_code == HTTPStatus.OK

    data = r.json()
    assert len(data) == 0


def test_proposals_invalid_method(go_server):
    r = requests.delete(f"{go_server}/proposals")
    assert r.status_code == HTTPStatus.UNAUTHORIZED

    s = requests.Session()

    login_payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=login_payload)
    assert r.status_code == HTTPStatus.CREATED
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    r = s.delete(f"{go_server}/proposals")
    assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED


def test_get_proposals_by_id(go_server):
    s = requests.Session()

    login_payload: dict[str, str] = {
        "email": "usr@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=login_payload)
    assert r.status_code == HTTPStatus.CREATED
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    proposal_payload: dict[str, str] = {
        "title": "Some Course Title",
        "summary": "A summary of some course.",
    }
    r = s.post(f"{go_server}/proposals", json=proposal_payload)
    assert r.status_code == HTTPStatus.CREATED

    proposal_payload = {
        "title": "Another Course Title",
        "summary": "A summary of another course.",
    }
    r = s.post(f"{go_server}/proposals", json=proposal_payload)
    assert r.status_code == HTTPStatus.CREATED

    r = s.get(f"{go_server}/proposals")
    assert r.status_code == HTTPStatus.OK
    assert len(r.json()) == 2

    p1, p2 = r.json()[0], r.json()[1]
    assert "id" in p1 and "id" in p2

    id1, id2 = int(p1["id"]), int(p2["id"])
    r = s.get(f"{go_server}/proposals/{id1}")
    assert r.status_code == HTTPStatus.OK
    assert r.json() == p1

    r = s.get(f"{go_server}/proposals/{id2}")
    assert r.status_code == HTTPStatus.OK
    assert r.json() == p2


def test_get_proposal_nonexistent(go_server):
    s = requests.Session()

    login_payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=login_payload)
    assert r.status_code == HTTPStatus.CREATED
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    r = s.get(f"{go_server}/proposals")
    assert r.status_code == HTTPStatus.OK

    data = r.json()
    assert len(data) == 0

    r = s.get(f"{go_server}/proposals/{2**63 - 1}")
    assert r.status_code == HTTPStatus.NOT_FOUND


def test_update_proposal(go_server):
    s = requests.Session()

    login_payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=login_payload)
    assert r.status_code == HTTPStatus.CREATED
    r = s.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    r = s.get(f"{go_server}/me")
    assert r.status_code == HTTPStatus.OK
    assert "id" in r.json()
    author_id = r.json()["id"]

    proposal_payload: dict[str, str] = {
        "title": "Title",
        "summary": "Summary",
        "author_id": author_id,
    }
    r = s.post(f"{go_server}/proposals", json=proposal_payload)
    assert r.status_code == HTTPStatus.CREATED
    assert "id" in r.json()

    pid: int = r.json()["id"]
    r = s.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "title" in r.json() and "summary" in r.json()
    assert r.json()["title"] == "Title"

    p = r.json()
    p["title"] = "New Title"
    r = s.patch(f"{go_server}/proposals/{pid}", json=p)
    assert r.status_code == HTTPStatus.NO_CONTENT

    r = s.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "title" in r.json() and "summary" in r.json()
    assert r.json()["title"] == "New Title"


def test_get_proposal_wrong_user(go_server):
    s = requests.Session()
    t = requests.Session()

    s_login_payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    t_login_payload: dict[str, str] = {
        "email": "user@example.org",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=s_login_payload)
    assert r.status_code == HTTPStatus.CREATED
    r = t.post(f"{go_server}/register", json=t_login_payload)
    assert r.status_code == HTTPStatus.CREATED
    r = s.post(f"{go_server}/login", json=s_login_payload)
    assert r.status_code == HTTPStatus.OK
    r = t.post(f"{go_server}/login", json=t_login_payload)
    assert r.status_code == HTTPStatus.OK

    s_proposal_payload: dict[str, str] = {
        "title": "S Title",
        "summary": "S Summary",
    }
    t_proposal_payload: dict[str, str] = {
        "title": "T Title",
        "summary": "T Summary",
    }
    r = s.post(f"{go_server}/proposals", json=s_proposal_payload)
    assert r.status_code == HTTPStatus.CREATED
    assert "id" in r.json()
    s_id = r.json()["id"]

    r = t.post(f"{go_server}/proposals", json=t_proposal_payload)
    assert r.status_code == HTTPStatus.CREATED
    assert "id" in r.json()
    t_id = r.json()["id"]

    r = s.get(f"{go_server}/proposals")
    assert r.status_code == HTTPStatus.OK
    assert len(r.json()) == 1
    assert r.json()[0]["title"] == "S Title"

    r = t.get(f"{go_server}/proposals")
    assert r.status_code == HTTPStatus.OK
    assert len(r.json()) == 1
    assert r.json()[0]["title"] == "T Title"

    r = s.get(f"{go_server}/proposals/{s_id}")
    assert r.status_code == HTTPStatus.OK
    r = t.get(f"{go_server}/proposals/{s_id}")
    assert r.status_code == HTTPStatus.NOT_FOUND
    r = s.get(f"{go_server}/proposals/{t_id}")
    assert r.status_code == HTTPStatus.NOT_FOUND
    r = t.get(f"{go_server}/proposals/{t_id}")
    assert r.status_code == HTTPStatus.OK


def test_create_proposal_rich(go_server):
    s = requests.Session()

    register_payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=register_payload)
    r = s.post(f"{go_server}/login", json=register_payload)
    assert r.status_code == HTTPStatus.OK

    proposal_payload: dict[str, str] = {
        "title": "Some Course Title",
        "summary": "A summary of some course.",
        "target_audience": "The target audience for some course.",
        "learning_objectives": "The learning objectives of some course.",
        "outline": """
            - item 1
            - item 2
        """,
        "assumed_prerequisites": "Some older course, some other older course, some skill.",
    }
    r = s.post(f"{go_server}/proposals", json=proposal_payload)
    assert r.status_code == HTTPStatus.CREATED
    assert "id" in r.json()

    r = s.get(f"{go_server}/proposals/{r.json()['id']}")
    assert r.status_code == HTTPStatus.OK

    for key, value in proposal_payload.items():
        assert key in r.json()
        assert r.json()[key] == value


def test_submit_proposal(go_server):
    s = requests.Session()

    register_payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=register_payload)
    r = s.post(f"{go_server}/login", json=register_payload)
    assert r.status_code == HTTPStatus.OK

    proposal_payload: dict[str, str] = {
        "title": "S Title",
        "summary": "S Summary",
    }
    r = s.post(f"{go_server}/proposals", json=proposal_payload)
    assert r.status_code == HTTPStatus.CREATED
    assert "id" in r.json()
    pid = r.json()["id"]

    r = s.post(f"{go_server}/proposals/{pid}/actions/submit")
    assert r.status_code == HTTPStatus.NO_CONTENT

    r = s.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert r.json()["status"] == "submitted"


def test_unknown_action(go_server):
    s = requests.Session()

    register_payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    r = s.post(f"{go_server}/register", json=register_payload)
    r = s.post(f"{go_server}/login", json=register_payload)
    assert r.status_code == HTTPStatus.OK

    proposal_payload: dict[str, str] = {
        "title": "S Title",
        "summary": "S Summary",
    }
    r = s.post(f"{go_server}/proposals", json=proposal_payload)
    assert r.status_code == HTTPStatus.CREATED
    assert "id" in r.json()
    pid = r.json()["id"]

    r = s.post(f"{go_server}/proposals/{pid}/actions/jump")
    assert r.status_code == HTTPStatus.NOT_FOUND


def test_admin_view_proposals(go_server):
    s = requests.Session()
    t = requests.Session()

    s_register_payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    t_register_payload: dict[str, str] = {
        "email": "admin@local.bytecourses.org",
        "password": "admin",
    }
    s.post(f"{go_server}/register", json=s_register_payload)
    s.post(f"{go_server}/login", json=s_register_payload)
    t.post(f"{go_server}/login", json=t_register_payload)

    s_proposal_payload: dict[str, str] = {
        "title": "Title",
        "summary": "Summary",
    }
    r = s.post(f"{go_server}/proposals", json=s_proposal_payload)
    assert r.status_code == HTTPStatus.CREATED
    assert "id" in r.json()
    s_pid_1 = r.json()["id"]

    s_proposal_payload = {
        "title": "Title 2",
        "summary": "Summary 2",
    }
    r = s.post(f"{go_server}/proposals", json=s_proposal_payload)
    assert r.status_code == HTTPStatus.CREATED
    assert "id" in r.json()
    s_pid_2 = r.json()["id"]

    t_proposal_payload: dict[str, str] = {
        "title": "Title 3",
        "summary": "Summary 3",
    }
    r = t.post(f"{go_server}/proposals", json=t_proposal_payload)
    assert r.status_code == HTTPStatus.CREATED
    assert "id" in r.json()
    t_pid_1 = r.json()["id"]

    r = s.post(f"{go_server}/proposals/{s_pid_2}/actions/submit")
    assert r.status_code == HTTPStatus.NO_CONTENT

    r = t.get(f"{go_server}/proposals")
    assert r.status_code == HTTPStatus.OK
    assert len(r.json()) == 1
    assert "id" in r.json()[0]
    assert r.json()[0]["id"] == s_pid_2

    r = t.get(f"{go_server}/proposals/mine")
    assert r.status_code == HTTPStatus.OK
    assert len(r.json()) == 1
    assert "id" in r.json()[0]
    assert r.json()[0]["id"] == t_pid_1

    assert s_pid_1 != s_pid_2 and s_pid_2 != t_pid_1


def test_admin_view_submitted_proposal(go_server):
    s = requests.Session()
    t = requests.Session()

    s_register_payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    t_register_payload: dict[str, str] = {
        "email": "admin@local.bytecourses.org",
        "password": "admin",
    }
    s.post(f"{go_server}/register", json=s_register_payload)
    s.post(f"{go_server}/login", json=s_register_payload)
    t.post(f"{go_server}/login", json=t_register_payload)

    s_proposal_payload: dict[str, str] = {
        "title": "Title",
        "summary": "Summary",
    }
    r = s.post(f"{go_server}/proposals", json=s_proposal_payload)
    assert r.status_code == HTTPStatus.CREATED
    assert "id" in r.json()
    s_pid = r.json()["id"]

    r = t.get(f"{go_server}/proposals")
    assert r.status_code == HTTPStatus.OK
    assert len(r.json()) == 0

    r = t.get(f"{go_server}/proposals/{s_pid}")
    assert r.status_code == HTTPStatus.NOT_FOUND

    r = s.post(f"{go_server}/proposals/{s_pid}/actions/submit")
    assert r.status_code == HTTPStatus.NO_CONTENT

    r = t.get(f"{go_server}/proposals")
    assert r.status_code == HTTPStatus.OK
    assert len(r.json()) == 1
    assert "id" in r.json()[0]
    assert r.json()[0]["id"] == s_pid

    r = t.get(f"{go_server}/proposals/{r.json()[0]['id']}")
    assert r.status_code == HTTPStatus.OK
    assert "id" in r.json()
    assert r.json()["id"] == s_pid

    assert "title" in r.json()
    assert r.json()["title"] == "Title"


def test_proposal_workflow_happy_path(go_server):
    u = requests.Session()
    a = requests.Session()

    u.post(
        f"{go_server}/login",
        json={
            "email": "user@local.bytecourses.org",
            "password": "user",
        },
    )
    a.post(
        f"{go_server}/login",
        json={
            "email": "admin@local.bytecourses.org",
            "password": "admin",
        },
    )

    r = u.post(
        f"{go_server}/proposals",
        json={
            "title": "Title",
            "summary": "Summary",
            "outline": "Outline",
        },
    )
    assert "id" in r.json()
    pid = r.json()["id"]

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json()
    assert r.json()["status"] == "draft"

    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.NOT_FOUND

    r = u.post(f"{go_server}/proposals/{pid}/actions/submit")
    assert r.status_code == HTTPStatus.NO_CONTENT
    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK

    r = u.patch(
        f"{go_server}/proposals/{pid}",
        json={
            "title": "New Title",
        },
    )
    assert r.status_code == HTTPStatus.CONFLICT

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "submitted"
    assert "title" in r.json() and r.json()["title"] == "Title"

    r = a.post(f"{go_server}/proposals/{pid}/actions/approve")
    assert r.status_code == HTTPStatus.NO_CONTENT

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "approved"

    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "approved"

    r = u.patch(
        f"{go_server}/proposals/{pid}",
        json={
            "title": "New Title",
        },
    )
    assert r.status_code == HTTPStatus.CONFLICT

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "approved"
    assert "title" in r.json() and r.json()["title"] == "Title"


def test_proposal_workflow_with_changes_requested(go_server):
    u = requests.Session()
    a = requests.Session()

    u.post(
        f"{go_server}/login",
        json={
            "email": "user@local.bytecourses.org",
            "password": "user",
        },
    )
    a.post(
        f"{go_server}/login",
        json={
            "email": "admin@local.bytecourses.org",
            "password": "admin",
        },
    )

    r = u.post(
        f"{go_server}/proposals",
        json={
            "title": "Title",
            "summary": "Summary",
            "outline": "Outline",
        },
    )
    assert "id" in r.json()
    pid = r.json()["id"]

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json()
    assert r.json()["status"] == "draft"

    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.NOT_FOUND

    r = u.post(f"{go_server}/proposals/{pid}/actions/submit")
    assert r.status_code == HTTPStatus.NO_CONTENT
    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK

    r = u.patch(
        f"{go_server}/proposals/{pid}",
        json={
            "title": "New Title",
        },
    )
    assert r.status_code == HTTPStatus.CONFLICT

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "submitted"
    assert "title" in r.json() and r.json()["title"] == "Title"

    r = a.post(
        f"{go_server}/proposals/{pid}/actions/request-changes",
        json={"review_notes": "Changes requested"},
    )
    assert r.status_code == HTTPStatus.NO_CONTENT

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "changes_requested"
    assert "title" in r.json() and r.json()["title"] == "Title"
    assert (
        "review_notes" in r.json() and r.json()["review_notes"] == "Changes requested"
    )

    r = u.patch(
        f"{go_server}/proposals/{pid}",
        json={"title": "New Title", "summary": "New Summary", "outline": "Outline"},
    )
    assert r.status_code == HTTPStatus.NO_CONTENT

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "changes_requested"
    assert "title" in r.json() and r.json()["title"] == "New Title"
    assert "outline" in r.json() and r.json()["outline"] == "Outline"

    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "changes_requested"

    r = u.post(f"{go_server}/proposals/{pid}/actions/submit")
    assert r.status_code == HTTPStatus.NO_CONTENT
    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "submitted"

    r = u.patch(
        f"{go_server}/proposals/{pid}",
        json={
            "title": "New Title",
        },
    )
    assert r.status_code == HTTPStatus.CONFLICT

    r = a.post(f"{go_server}/proposals/{pid}/actions/approve")
    assert r.status_code == HTTPStatus.NO_CONTENT

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "approved"

    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "approved"

    r = u.patch(
        f"{go_server}/proposals/{pid}",
        json={
            "title": "Old Title",
        },
    )
    assert r.status_code == HTTPStatus.CONFLICT

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "approved"
    assert "title" in r.json() and r.json()["title"] == "New Title"


def test_proposal_workflow_with_patch(go_server):
    u = requests.Session()
    a = requests.Session()

    u.post(
        f"{go_server}/login",
        json={
            "email": "user@local.bytecourses.org",
            "password": "user",
        },
    )
    a.post(
        f"{go_server}/login",
        json={
            "email": "admin@local.bytecourses.org",
            "password": "admin",
        },
    )

    r = u.post(
        f"{go_server}/proposals",
        json={
            "title": "Title",
            "summary": "Summary",
            "outline": "Outline",
        },
    )
    assert "id" in r.json()
    pid = r.json()["id"]

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json()
    assert r.json()["status"] == "draft"
    assert "outline" in r.json() and r.json()["outline"] == "Outline"

    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.NOT_FOUND

    r = u.patch(
        f"{go_server}/proposals/{pid}",
        json={
            "title": "New Title",
        },
    )
    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "draft"
    assert "title" in r.json() and r.json()["title"] == "New Title"
    assert "outline" in r.json() and r.json()["outline"] == ""
    assert "summary" in r.json() and r.json()["summary"] == ""

    r = u.patch(
        f"{go_server}/proposals/{pid}",
        json={
            "title": "New Title",
            "summary": "New Summary",
        },
    )
    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "draft"
    assert "title" in r.json() and r.json()["title"] == "New Title"
    assert "summary" in r.json() and r.json()["summary"] == "New Summary"
    assert "outline" in r.json() and r.json()["outline"] == ""

    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.NOT_FOUND

    r = u.post(f"{go_server}/proposals/{pid}/actions/submit")
    assert r.status_code == HTTPStatus.NO_CONTENT
    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK

    r = u.patch(
        f"{go_server}/proposals/{pid}",
        json={
            "title": "Old Title",
        },
    )
    assert r.status_code == HTTPStatus.CONFLICT

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "submitted"
    assert "title" in r.json() and r.json()["title"] == "New Title"

    r = a.post(f"{go_server}/proposals/{pid}/actions/approve")
    assert r.status_code == HTTPStatus.NO_CONTENT

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "approved"

    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "approved"

    r = u.patch(
        f"{go_server}/proposals/{pid}",
        json={
            "title": "New Title",
        },
    )
    assert r.status_code == HTTPStatus.CONFLICT

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "approved"
    assert "title" in r.json() and r.json()["title"] == "New Title"
    assert "outline" in r.json() and r.json()["outline"] == ""


def test_proposal_workflow_delete_draft(go_server):
    u = requests.Session()
    u.post(
        f"{go_server}/login",
        json={
            "email": "user@local.bytecourses.org",
            "password": "user",
        },
    )

    r = u.post(
        f"{go_server}/proposals",
        json={
            "title": "Title",
            "summary": "Summary",
            "outline": "Outline",
        },
    )
    assert "id" in r.json()
    pid = r.json()["id"]

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json()
    assert r.json()["status"] == "draft"
    assert "outline" in r.json() and r.json()["outline"] == "Outline"

    r = u.delete(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.NO_CONTENT

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.NOT_FOUND
    r = u.patch(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.NOT_FOUND


def test_proposal_workflow_reject_submission(go_server):
    u = requests.Session()
    a = requests.Session()

    u.post(
        f"{go_server}/login",
        json={
            "email": "user@local.bytecourses.org",
            "password": "user",
        },
    )
    a.post(
        f"{go_server}/login",
        json={
            "email": "admin@local.bytecourses.org",
            "password": "admin",
        },
    )

    r = u.post(
        f"{go_server}/proposals",
        json={
            "title": "Title",
            "summary": "Summary",
            "outline": "Outline",
        },
    )
    assert "id" in r.json()
    pid = r.json()["id"]

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json()
    assert r.json()["status"] == "draft"

    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.NOT_FOUND

    r = u.post(f"{go_server}/proposals/{pid}/actions/submit")
    assert r.status_code == HTTPStatus.NO_CONTENT
    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK

    r = u.patch(
        f"{go_server}/proposals/{pid}",
        json={
            "title": "New Title",
        },
    )
    assert r.status_code == HTTPStatus.CONFLICT

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "submitted"
    assert "title" in r.json() and r.json()["title"] == "Title"

    r = a.post(f"{go_server}/proposals/{pid}/actions/reject")
    assert r.status_code == HTTPStatus.NO_CONTENT

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "rejected"

    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "rejected"

    r = u.patch(
        f"{go_server}/proposals/{pid}",
        json={
            "title": "New Title",
        },
    )
    assert r.status_code == HTTPStatus.CONFLICT

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "rejected"
    assert "title" in r.json() and r.json()["title"] == "Title"


def test_proposal_workflow_withdraw_submission(go_server):
    u = requests.Session()
    a = requests.Session()

    u.post(
        f"{go_server}/login",
        json={
            "email": "user@local.bytecourses.org",
            "password": "user",
        },
    )
    a.post(
        f"{go_server}/login",
        json={
            "email": "admin@local.bytecourses.org",
            "password": "admin",
        },
    )

    r = u.post(
        f"{go_server}/proposals",
        json={
            "title": "Title",
            "summary": "Summary",
            "outline": "Outline",
        },
    )
    assert "id" in r.json()
    pid = r.json()["id"]

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json()
    assert r.json()["status"] == "draft"

    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.NOT_FOUND

    r = u.post(f"{go_server}/proposals/{pid}/actions/submit")
    assert r.status_code == HTTPStatus.NO_CONTENT
    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK

    r = u.patch(
        f"{go_server}/proposals/{pid}",
        json={
            "title": "New Title",
        },
    )
    assert r.status_code == HTTPStatus.CONFLICT

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "submitted"
    assert "title" in r.json() and r.json()["title"] == "Title"

    r = u.post(f"{go_server}/proposals/{pid}/actions/withdraw")
    assert r.status_code == HTTPStatus.NO_CONTENT
    r = a.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.NOT_FOUND

    r = u.get(f"{go_server}/proposals/{pid}")
    assert r.status_code == HTTPStatus.OK
    assert "status" in r.json() and r.json()["status"] == "withdrawn"
    assert "title" in r.json() and r.json()["title"] == "Title"

    r = u.patch(
        f"{go_server}/proposals/{pid}",
        json={
            "title": "New Title",
        },
    )
    assert r.status_code == HTTPStatus.CONFLICT
