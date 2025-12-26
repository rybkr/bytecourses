import pytest
import requests
from http import HTTPStatus

API_ROOT: str = "http://localhost:8080/api"


def test_create_proposal(go_server):
    register_payload: dict[str, str] = {
        "email": "usr@example.com",
        "password": "password123",
    }
    r = requests.post(f"{API_ROOT}/register", json=register_payload)
    assert r.status_code == HTTPStatus.OK

    login_payload: dict[str, str] = {
        "email": "usr@example.com",
        "password": "password123",
    }
    r = requests.post(f"{API_ROOT}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK

    proposal_payload: dict[str, str] = {
        "title": "Some Course Title",
        "summary": "A summary of some course.",
    }
    r = requests.post(f"{API_ROOT}/proposals", json=proposal_payload)
    assert r.status_code == HTTPStatus.OK


def test_get_proposal_by_id(go_server)
