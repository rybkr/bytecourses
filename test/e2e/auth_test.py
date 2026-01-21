import requests
from http import HTTPStatus

from .conftest import register_and_login


class TestRegister:
    def test_creates_user_with_valid_credentials(self, api_url):
        payload = {
            "email": "newuser@example.com",
            "password": "password123",
            "name": "New User",
        }
        r = requests.post(f"{api_url}/register", json=payload)
        assert r.status_code == HTTPStatus.CREATED
        assert "id" in r.json()

    def test_creates_user_with_name(self, api_url):
        session = register_and_login(
            api_url,
            email="user@example.com",
            password="secret",
            name="Test User",
        )
        r = session.get(f"{api_url}/me")
        assert r.status_code == HTTPStatus.OK
        assert "name" in r.json() and r.json()["name"] == "Test User"
        assert "email" in r.json() and r.json()["email"] == "user@example.com"

    def test_rejects_duplicate_email(self, api_url):
        payload = {
            "email": "duplicate@example.com",
            "password": "password123",
            "name": "Name",
        }
        r = requests.post(f"{api_url}/register", json=payload)
        assert r.status_code == HTTPStatus.CREATED

        r = requests.post(f"{api_url}/register", json=payload)
        assert r.status_code == HTTPStatus.CONFLICT

    def test_rejects_missing_email(self, api_url):
        payload = {"password": "password1234", "name": "Name"}
        r = requests.post(f"{api_url}/register", json=payload)
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_empty_email(self, api_url):
        payload = {"email": "", "password": "password1234", "name": "Name"}
        r = requests.post(f"{api_url}/register", json=payload)
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_missing_password(self, api_url):
        payload = {"email": "nopass@example.com", "name": "Name"}
        r = requests.post(f"{api_url}/register", json=payload)
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_empty_password(self, api_url):
        payload = {"email": "emptypass@example.com", "password": "", "name": "Name"}
        r = requests.post(f"{api_url}/register", json=payload)
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_empty_payload(self, api_url):
        r = requests.post(f"{api_url}/register")
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_get_method(self, api_url):
        r = requests.get(f"{api_url}/register")
        assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED

    def test_rejects_delete_method(self, api_url):
        r = requests.delete(f"{api_url}/register")
        assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED

    def test_trims_whitespace(self, api_url):
        session = register_and_login(
            api_url,
            email="   user@example.com  ",
            password="secret",
            name="Test User   \
            ",
        )
        r = session.get(f"{api_url}/me")
        assert r.status_code == HTTPStatus.OK
        assert "name" in r.json() and r.json()["name"] == "Test User"
        assert "email" in r.json() and r.json()["email"] == "user@example.com"


class TestLogin:
    def test_succeeds_with_valid_credentials(self, api_url):
        payload = {
            "email": "logintest@example.com",
            "password": "password123",
            "name": "Name",
        }
        requests.post(f"{api_url}/register", json=payload)

        del payload["name"]
        r = requests.post(f"{api_url}/login", json=payload)
        assert r.status_code == HTTPStatus.OK

    def test_allows_multiple_logins(self, api_url):
        payload = {
            "email": "multilogin@example.com",
            "password": "password123",
            "name": "Name",
        }
        requests.post(f"{api_url}/register", json=payload)

        del payload["name"]
        for _ in range(3):
            r = requests.post(f"{api_url}/login", json=payload)
            assert r.status_code == HTTPStatus.OK

    def test_rejects_wrong_password(self, api_url):
        payload = {
            "email": "logintest@example.com",
            "password": "password123",
            "name": "Name",
        }
        requests.post(f"{api_url}/register", json=payload)

        payload = {
            "email": "logintest@example.com",
            "password": "wrongpassword",
        }
        r = requests.post(f"{api_url}/login", json=payload)
        assert r.status_code == HTTPStatus.UNAUTHORIZED

    def test_rejects_missing_password(self, api_url):
        payload = {
            "email": "logintest@example.com",
            "password": "password123",
            "name": "Name",
        }
        requests.post(f"{api_url}/register", json=payload)

        payload = {
            "email": "logintest@example.com",
        }
        r = requests.post(f"{api_url}/login", json=payload)
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_empty_password(self, api_url):
        payload = {
            "email": "logintest@example.com",
            "password": "password123",
            "name": "Name",
        }
        requests.post(f"{api_url}/register", json=payload)

        payload = {
            "email": "logintest@example.com",
            "password": "",
        }
        r = requests.post(f"{api_url}/login", json=payload)
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_missing_email(self, api_url):
        r = requests.post(f"{api_url}/login", json={"password": "password123"})
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_empty_email(self, api_url):
        r = requests.post(
            f"{api_url}/login",
            json={"email": "", "password": "password123"},
        )
        assert r.status_code == HTTPStatus.BAD_REQUEST

    def test_rejects_get_method(self, api_url):
        r = requests.get(f"{api_url}/login")
        assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED

    def test_rejects_delete_method(self, api_url):
        r = requests.delete(f"{api_url}/login")
        assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED


class TestLogout:
    def test_invalidates_session(self, api_url, user_session):
        r = user_session.post(f"{api_url}/logout")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = user_session.get(f"{api_url}/me")
        assert r.status_code == HTTPStatus.UNAUTHORIZED

    def test_rejects_get_method(self, api_url):
        r = requests.get(f"{api_url}/logout")
        assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED

    def test_rejects_delete_method(self, api_url):
        r = requests.delete(f"{api_url}/logout")
        assert r.status_code == HTTPStatus.METHOD_NOT_ALLOWED


class TestAuthFlow:
    def test_register_login_me_logout_flow(self, api_url):
        session = requests.Session()

        r = session.post(
            f"{api_url}/register",
            json={
                "email": "flowtest@example.com",
                "password": "secret",
                "name": "Name",
            },
        )
        assert r.status_code == HTTPStatus.CREATED

        r = session.post(
            f"{api_url}/login",
            json={"email": "flowtest@example.com", "password": "secret"},
        )
        assert r.status_code == HTTPStatus.OK
        assert "session" in session.cookies

        r = session.get(f"{api_url}/me")
        assert r.status_code == HTTPStatus.OK
        assert r.json()["email"] == "flowtest@example.com"

        r = session.post(f"{api_url}/logout")
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = session.get(f"{api_url}/me")
        assert r.status_code == HTTPStatus.UNAUTHORIZED


class TestUserRoles:
    def test_admin_user_has_admin_role(self, api_url, admin_session):
        r = admin_session.get(f"{api_url}/me")
        assert r.status_code == HTTPStatus.OK
        assert r.json()["role"] == "admin"

    def test_regular_user_has_student_role(self, api_url, user_session):
        r = user_session.get(f"{api_url}/me")
        assert r.status_code == HTTPStatus.OK
        assert r.json()["role"] == "student"


class TestUpdateProfile:
    def test_update_user_name(self, api_url, user_session):
        r = user_session.get(f"{api_url}/me")
        assert r.json()["name"] == "Guest User"
        userID = r.json()["id"]

        r = user_session.patch(f"{api_url}/me", json={"name": "New Name"})
        assert r.status_code == HTTPStatus.NO_CONTENT

        r = user_session.get(f"{api_url}/me")
        assert r.json()["name"] == "New Name"
