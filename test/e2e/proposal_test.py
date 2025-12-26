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


def test_get
