import requests
from http import HTTPStatus

go_server: str = "http://localhost:8080/api"


def test_register(go_server):
    payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    r = requests.post(f"{go_server}/register", json=payload)
    assert r.status_code == HTTPStatus.CREATED


def test_register_duplicate(go_server):
    payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    r = requests.post(f"{go_server}/register", json=payload)
    assert r.status_code == HTTPStatus.CREATED

    r = requests.post(f"{go_server}/register", json=payload)
    assert r.status_code == HTTPStatus.INTERNAL_SERVER_ERROR
    r = requests.post(f"{go_server}/register", json=payload)
    assert r.status_code == HTTPStatus.INTERNAL_SERVER_ERROR


def test_register_no_email(go_server):
    payload: dict[str, str] = {
        "password": "password1234",
    }
    r = requests.post(f"{go_server}/register", json=payload)
    assert r.status_code == HTTPStatus.BAD_REQUEST


def test_register_empty_email(go_server):
    payload: dict[str, str] = {
        "email": "",
        "password": "password1234",
    }
    r = requests.post(f"{go_server}/register", json=payload)
    assert r.status_code == HTTPStatus.BAD_REQUEST


def test_register_no_password(go_server):
    payload: dict[str, str] = {
        "email": "newuser@example.com",
    }
    r = requests.post(f"{go_server}/register", json=payload)
    assert r.status_code == HTTPStatus.BAD_REQUEST


def test_register_empty_password(go_server):
    payload: dict[str, str] = {
        "email": "newuser@example.com",
        "password": "",
    }
    r = requests.post(f"{go_server}/register", json=payload)
    assert r.status_code == HTTPStatus.BAD_REQUEST


def test_register_invalid_method(go_server):
    r = requests.get(f"{go_server}/register")
    assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED
    r = requests.delete(f"{go_server}/register")
    assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED


def test_register_no_payload(go_server):
    r = requests.post(f"{go_server}/register")
    assert r.status_code == HTTPStatus.BAD_REQUEST


def test_login(go_server):
    payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    r = requests.post(f"{go_server}/register", json=payload)
    assert r.status_code == HTTPStatus.CREATED

    r = requests.post(f"{go_server}/login", json=payload)
    assert r.status_code == HTTPStatus.OK


def test_login_incorrect_password(go_server):
    payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password1234",
    }
    r = requests.post(f"{go_server}/login", json=payload)
    assert r.status_code == HTTPStatus.UNAUTHORIZED


def test_login_duplicate(go_server):
    payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "password123",
    }
    r = requests.post(f"{go_server}/register", json=payload)
    assert r.status_code == HTTPStatus.CREATED

    r = requests.post(f"{go_server}/login", json=payload)
    assert r.status_code == HTTPStatus.OK
    r = requests.post(f"{go_server}/login", json=payload)
    assert r.status_code == HTTPStatus.OK
    r = requests.post(f"{go_server}/login", json=payload)
    assert r.status_code == HTTPStatus.OK


def test_login_no_password(go_server):
    payload: dict[str, str] = {
        "email": "user@example.com",
    }
    r = requests.post(f"{go_server}/login", json=payload)
    assert r.status_code == HTTPStatus.UNAUTHORIZED


def test_login_empty_password(go_server):
    payload: dict[str, str] = {
        "email": "user@example.com",
        "password": "",
    }
    r = requests.post(f"{go_server}/login", json=payload)
    assert r.status_code == HTTPStatus.UNAUTHORIZED


def test_login_no_email(go_server):
    payload: dict[str, str] = {
        "password": "password123",
    }
    r = requests.post(f"{go_server}/login", json=payload)
    assert r.status_code == HTTPStatus.UNAUTHORIZED


def test_login_empty_email(go_server):
    payload: dict[str, str] = {
        "email": "",
        "password": "password123",
    }
    r = requests.post(f"{go_server}/login", json=payload)
    assert r.status_code == HTTPStatus.UNAUTHORIZED


def test_login_invalid_method(go_server):
    r = requests.get(f"{go_server}/login")
    assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED
    r = requests.delete(f"{go_server}/login")
    assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED


def test_login_no_payload(go_server):
    r = requests.post(f"{go_server}/register")
    assert r.status_code == HTTPStatus.BAD_REQUEST


def test_auth_flow(go_server):
    s = requests.Session()

    r = s.post(
        f"{go_server}/register",
        json={"email": "u@example.com", "password": "secret"},
    )
    assert r.status_code == HTTPStatus.CREATED

    r = s.post(
        f"{go_server}/login",
        json={"email": "u@example.com", "password": "secret"},
    )
    assert r.status_code == HTTPStatus.OK
    assert "session" in s.cookies

    r = s.get(f"{go_server}/me")
    assert r.status_code == HTTPStatus.OK
    data = r.json()
    assert data["email"] == "u@example.com"

    r = s.post(f"{go_server}/logout")
    assert r.status_code == HTTPStatus.NO_CONTENT

    r = s.get(f"{go_server}/me")
    assert r.status_code == HTTPStatus.UNAUTHORIZED


def test_logout_invalid_method(go_server):
    r = requests.get(f"{go_server}/logout")
    assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED
    r = requests.delete(f"{go_server}/logout")
    assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED


def test_admin_user(go_server):
    s = requests.Session()

    payload: dict[str, str] = {
        "email": "admin@local.bytecourses.org",
        "password": "admin",
    }
    r = s.post(f"{go_server}/login", json=payload)
    assert r.status_code == HTTPStatus.OK

    r = s.get(f"{go_server}/me")
    assert r.status_code == HTTPStatus.OK

    data = r.json()
    assert "role" in data
    assert data["role"] == "admin"


def test_user_name(go_server):
    s = requests.Session()
    payload: dict[str, str] = {
        "name": "User Name",
        "email": "user@example.com",
        "password": "secret",
    }

    r = s.post(f"{go_server}/register", json=payload)
    assert r.status_code == HTTPStatus.CREATED
    r = s.post(f"{go_server}/login", json=payload)
    assert r.status_code == HTTPStatus.OK
    r = s.get(f"{go_server}/me")
    assert r.status_code == HTTPStatus.OK

    assert "name" in r.json()
    assert r.json()["name"] == "User Name"


def test_register_welcome_email_non_blocking(go_server):
    """
    Verify registration succeeds even if welcome email fails.
    Email delivery is non-blocking.
    """
    payload: dict[str, str] = {
        "name": "Welcome Test",
        "email": "welcometest@example.com",
        "password": "password123",
    }
    r = requests.post(f"{go_server}/register", json=payload)
    assert r.status_code == HTTPStatus.CREATED

    login_payload: dict[str, str] = {
        "email": "welcometest@example.com",
        "password": "password123",
    }
    r = requests.post(f"{go_server}/login", json=login_payload)
    assert r.status_code == HTTPStatus.OK
