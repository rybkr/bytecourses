import pytest
import requests
from http import HTTPStatus

from test.conftest import USER_EMAIL, USER_PASSWORD, ADMIN_EMAIL, ADMIN_PASSWORD

_original_init = requests.Session.__init__
_original_post = requests.Session.post
_original_patch = requests.Session.patch
_original_delete = requests.Session.delete

_original_module_post = requests.post
_original_module_patch = requests.patch
_original_module_delete = requests.delete


def _csrf_init(self, base_url=None, *args, **kwargs):
    _original_init(self, *args, **kwargs)
    self._csrf_base_url = base_url
    self._csrf_token_initialized = False


def _ensure_csrf_token(self, url=None):
    if not self.cookies.get("csrf-token") and not self._csrf_token_initialized:
        if self._csrf_base_url:
            self.get(f"{self._csrf_base_url}/api/health")
        elif url:
            if url.startswith("http"):
                from urllib.parse import urlparse

                parsed = urlparse(url)
                base = f"{parsed.scheme}://{parsed.netloc}"
                self.get(f"{base}/api/health")
            else:
                parts = url.split("/")
                if len(parts) >= 3:
                    base = "/".join(parts[:3])
                    self.get(f"{base}/api/health")
        self._csrf_token_initialized = True


def _add_csrf_header(self, kwargs):
    if "headers" not in kwargs:
        kwargs["headers"] = {}
    if "X-CSRF-Token" not in kwargs["headers"]:
        csrf_token = self.cookies.get("csrf-token")
        if csrf_token:
            kwargs["headers"]["X-CSRF-Token"] = csrf_token


def _csrf_post(self, url, *args, **kwargs):
    _ensure_csrf_token(self, url)
    _add_csrf_header(self, kwargs)
    return _original_post(self, url, *args, **kwargs)


def _csrf_patch(self, url, *args, **kwargs):
    _ensure_csrf_token(self, url)
    _add_csrf_header(self, kwargs)
    return _original_patch(self, url, *args, **kwargs)


def _csrf_delete(self, url, *args, **kwargs):
    _ensure_csrf_token(self, url)
    _add_csrf_header(self, kwargs)
    return _original_delete(self, url, *args, **kwargs)


requests.Session.__init__ = _csrf_init
requests.Session.post = _csrf_post
requests.Session.patch = _csrf_patch
requests.Session.delete = _csrf_delete


def _csrf_module_post(url, *args, **kwargs):
    with requests.Session() as session:
        return session.post(url, *args, **kwargs)


def _csrf_module_patch(url, *args, **kwargs):
    with requests.Session() as session:
        return session.patch(url, *args, **kwargs)


def _csrf_module_delete(url, *args, **kwargs):
    with requests.Session() as session:
        return session.delete(url, *args, **kwargs)


requests.post = _csrf_module_post
requests.patch = _csrf_module_patch
requests.delete = _csrf_module_delete


@pytest.fixture(scope="function")
def user_session(api_url):
    session = requests.Session(base_url=api_url.replace("/api", ""))
    r = session.post(
        f"{api_url}/login",
        json={"email": USER_EMAIL, "password": USER_PASSWORD},
    )
    assert r.status_code == HTTPStatus.OK, "Failed to login as seeded user"
    return session


@pytest.fixture(scope="function")
def admin_session(api_url):
    session = requests.Session(base_url=api_url.replace("/api", ""))
    r = session.post(
        f"{api_url}/login",
        json={"email": ADMIN_EMAIL, "password": ADMIN_PASSWORD},
    )
    assert r.status_code == HTTPStatus.OK, "Failed to login as seeded admin"
    return session


def register_and_login(api_url: str, email: str, password: str, name: str = "Name"):
    base_url = api_url.replace("/api", "")
    session = requests.Session(base_url=base_url)
    payload = {"email": email, "password": password, "name": name}
    r = session.post(f"{api_url}/register", json=payload)
    assert r.status_code == HTTPStatus.CREATED, f"Failed to register user {email}"
    assert "id" in r.json(), f"Failed to register user {email}"
    session.user_id = r.json()["id"]
    r = session.post(f"{api_url}/login", json={"email": email, "password": password})
    assert r.status_code == HTTPStatus.OK, f"Failed to login as {email}"
    return session
