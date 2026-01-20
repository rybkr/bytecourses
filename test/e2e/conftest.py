import pytest
import requests
from http import HTTPStatus

from test.conftest import USER_EMAIL, USER_PASSWORD, ADMIN_EMAIL, ADMIN_PASSWORD


@pytest.fixture(scope="function")
def user_session(api_url):
    session = requests.Session()
    r = session.post(
        f"{api_url}/login",
        json={"email": USER_EMAIL, "password": USER_PASSWORD},
    )
    assert r.status_code == HTTPStatus.OK, "Failed to login as seeded user"
    return session


@pytest.fixture(scope="function")
def admin_session(api_url):
    session = requests.Session()
    r = session.post(
        f"{api_url}/login",
        json={"email": ADMIN_EMAIL, "password": ADMIN_PASSWORD},
    )
    assert r.status_code == HTTPStatus.OK, "Failed to login as seeded admin"
    return session


def register_and_login(api_url: str, email: str, password: str, name: str):
    session = requests.Session()
    payload = {"email": email, "password": password, "name": name}
    r = session.post(f"{api_url}/register", json=payload)
    assert r.status_code == HTTPStatus.CREATED, f"Failed to register user {email}"
    r = session.post(f"{api_url}/login", json={"email": email, "password": password})
    assert r.status_code == HTTPStatus.OK, f"Failed to login as {email}"
    return session
