import pytest
import requests
from http import HTTPStatus

API_ROOT: str = "http://localhost:8080/api"


def test_create_proposal(go_server):
    s = requests.Session()

    register_payload: dict[str, str] = {
        "email": "usr@example.com",
        "password": "password123",
    }
    r = s.post(f"{API_ROOT}/register", json=register_payload)
    assert r.status_code == HTTPStatus.OK

    login_payload: dict[str, str] = {
        "email": "usr@example.com",
        "password": "password123",
    }
    r = s.post(f"{API_ROOT}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    proposal_payload: dict[str, str] = {
        "title": "Some Course Title",
        "summary": "A summary of some course.",
    }
    r = s.post(f"{API_ROOT}/proposals", json=proposal_payload)
    assert r.status_code == HTTPStatus.OK

    proposal_payload = {
        "title": "Another Course Title",
        "summary": "A summary of another course.",
    }
    r = s.post(f"{API_ROOT}/proposals", json=proposal_payload)
    assert r.status_code == HTTPStatus.OK


def test_get_proposals(go_server):
    s = requests.Session()

    login_payload: dict[str, str] = {
        "email": "usr@example.com",
        "password": "password123",
    }
    r = s.post(f"{API_ROOT}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    r = s.get(f"{API_ROOT}/proposals")
    assert r.status_code == HTTPStatus.OK

    data = r.json()
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
    r = s.post(f"{API_ROOT}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    r = s.get(f"{API_ROOT}/proposals")
    assert r.status_code == HTTPStatus.OK

    data = r.json()
    assert len(data) == 0


def test_proposals_invalid_method(go_server):
    r = requests.put(f"{API_ROOT}/proposals")
    assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED
    r = requests.delete(f"{API_ROOT}/proposals")
    assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED


def test_get_proposals_by_id(go_server):
    s = requests.Session()

    login_payload: dict[str, str] = {
        "email": "usr@example.com",
        "password": "password123",
    }
    r = s.post(f"{API_ROOT}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    r = s.get(f"{API_ROOT}/proposals")
    assert r.status_code == HTTPStatus.OK
    assert len(r.json()) == 2

    p1, p2 = r.json()[0], r.json()[1]
    assert "id" in p1 and "id" in p2

    id1, id2 = int(p1["id"]), int(p2["id"])
    r = s.get(f"{API_ROOT}/proposals/{id1}")
    assert r.status_code == HTTPStatus.OK
    assert r.json() == p1

    r = s.get(f"{API_ROOT}/proposals/{id2}")
    assert r.status_code == HTTPStatus.OK
    assert r.json() == p2


def test_get_proposal_nonexistent(go_server):
    s = requests.Session()

    login_payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    r = s.post(f"{API_ROOT}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    r = s.get(f"{API_ROOT}/proposals")
    assert r.status_code == HTTPStatus.OK

    data = r.json()
    assert len(data) == 0

    r = s.get(f"{API_ROOT}/proposals/{2**63 - 1}")
    assert r.status_code == HTTPStatus.NOT_FOUND
