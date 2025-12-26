import pytest
import requests
from http import HTTPStatus

API_ROOT: str = "http://localhost:8080/api"


def test_register(go_server):
    payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    r = requests.post(f"{API_ROOT}/register", json=payload)
    assert r.status_code == HTTPStatus.OK

    payload["email"], payload["password"] = "jane.doe@example.com", "qwerty"
    r = requests.post(f"{API_ROOT}/register", json=payload)
    assert r.status_code == HTTPStatus.OK


def test_register_duplicate(go_server):
    payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    r = requests.post(f"{API_ROOT}/register", json=payload)
    assert r.status_code == HTTPStatus.BAD_REQUEST

    payload["email"], payload["password"] = "jane.doe@example.com", "qwerty"
    r = requests.post(f"{API_ROOT}/register", json=payload)
    assert r.status_code == HTTPStatus.BAD_REQUEST


def test_register_no_email(go_server):
    payload: dict[str, str] = {
        "password": "password1234",
    }
    r = requests.post(f"{API_ROOT}/register", json=payload)
    assert r.status_code == HTTPStatus.BAD_REQUEST


def test_register_empty_email(go_server):
    payload: dict[str, str] = {
        "email": "",
        "password": "password1234",
    }
    r = requests.post(f"{API_ROOT}/register", json=payload)
    assert r.status_code == HTTPStatus.BAD_REQUEST


def test_register_no_password(go_server):
    payload: dict[str, str] = {
        "email": "newuser@example.com",
    }
    r = requests.post(f"{API_ROOT}/register", json=payload)
    assert r.status_code == HTTPStatus.BAD_REQUEST


def test_register_empty_password(go_server):
    payload: dict[str, str] = {
        "email": "newuser@example.com",
        "password": "",
    }
    r = requests.post(f"{API_ROOT}/register", json=payload)
    assert r.status_code == HTTPStatus.BAD_REQUEST


def test_register_invalid_method(go_server):
    r = requests.get(f"{API_ROOT}/register")
    assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED
    r = requests.delete(f"{API_ROOT}/register")
    assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED


def test_register_no_payload(go_server):
    r = requests.post(f"{API_ROOT}/register")
    assert r.status_code == HTTPStatus.BAD_REQUEST


def test_login(go_server):
    payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    r = requests.post(f"{API_ROOT}/login", json=payload)
    assert r.status_code == HTTPStatus.OK

    payload["email"], payload["password"] = "jane.doe@example.com", "qwerty"
    r = requests.post(f"{API_ROOT}/login", json=payload)
    assert r.status_code == HTTPStatus.OK


def test_login_incorrect_password(go_server):
    payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password1234",
    }
    r = requests.post(f"{API_ROOT}/login", json=payload)
    assert r.status_code == HTTPStatus.UNAUTHORIZED


def test_login_duplicate(go_server):
    payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    r = requests.post(f"{API_ROOT}/login", json=payload)
    assert r.status_code == HTTPStatus.OK


def test_login_no_password(go_server):
    payload: dict[str, str] = {
        "email": "user@example.com",
    }
    r = requests.post(f"{API_ROOT}/login", json=payload)
    assert r.status_code == HTTPStatus.UNAUTHORIZED


def test_login_empty_password(go_server):
    payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "",
    }
    r = requests.post(f"{API_ROOT}/login", json=payload)
    assert r.status_code == HTTPStatus.UNAUTHORIZED


def test_login_no_email(go_server):
    payload: dict[str, str] = {
        "password": "password123",
    }
    r = requests.post(f"{API_ROOT}/login", json=payload)
    assert r.status_code == HTTPStatus.UNAUTHORIZED


def test_login_empty_email(go_server):
    payload: dict[str, str] = {
        "email": "",
        "password": "password123",
    }
    r = requests.post(f"{API_ROOT}/login", json=payload)
    assert r.status_code == HTTPStatus.UNAUTHORIZED


def test_login_invalid_method(go_server):
    r = requests.get(f"{API_ROOT}/login")
    assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED
    r = requests.delete(f"{API_ROOT}/login")
    assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED


def test_login_no_payload(go_server):
    r = requests.post(f"{API_ROOT}/register")
    assert r.status_code == HTTPStatus.BAD_REQUEST


def test_auth_flow(go_server):
    s = requests.Session()

    r = s.post(
        f"{API_ROOT}/register",
        json={"email": "u@example.com", "password": "secret"},
    )
    assert r.status_code == HTTPStatus.OK

    r = s.post(
        f"{API_ROOT}/login",
        json={"email": "u@example.com", "password": "secret"},
    )
    assert r.status_code == HTTPStatus.OK
    assert "session" in s.cookies

    r = s.get(f"{API_ROOT}/me")
    assert r.status_code == HTTPStatus.OK
    data = r.json()
    assert data["email"] == "u@example.com"

    r = s.post(f"{API_ROOT}/logout")
    assert r.status_code == HTTPStatus.NO_CONTENT

    r = s.get(f"{API_ROOT}/me")
    assert r.status_code == HTTPStatus.UNAUTHORIZED


def test_logout_invalid_method():
    r = requests.get(f"{API_ROOT}/logout")
    assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED
    r = requests.delete(f"{API_ROOT}/logout")
    assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED


def test_admin_user():
    s = requests.Session()

    payload: dict[str, str] = {
        "email": "admin@local.bytecourses.org",
        "password": "admin",
    }
    r = s.post(f"{API_ROOT}/login", json=payload)
    assert r.status_code == HTTPStatus.OK

    r = s.get(f"{API_ROOT}/me")
    assert r.status_code == HTTPStatus.OK

    data = r.json()
    assert "role" in data
    assert data["role"] == "admin"
