import os
import signal
import socket
import subprocess
import time
from http import HTTPStatus

import pytest
import requests

USER_EMAIL = "user@bytecourses.org"
USER_PASSWORD = "user"
ADMIN_EMAIL = "admin@bytecourses.org"
ADMIN_PASSWORD = "admin"


def get_free_port():
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


@pytest.fixture(scope="function")
def go_server():
    port = get_free_port()
    base_url = f"http://127.0.0.1:{port}"
    env = os.environ.copy()
    env["PORT"] = f"{port}"

    proc = subprocess.Popen(
        [
            "go",
            "run",
            "cmd/server/main.go",
            "--bcrypt-cost=5",
            "--email-service=none",
            "--storage=memory",
            "--seed-users=./test/fixtures/users.json",
            "--seed-proposals=./test/fixtures/proposals.json",
        ],
        env=env,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        start_new_session=True,
    )

    for _ in range(30):
        try:
            r = requests.get(f"{base_url}/api/health")
            if r.status_code == HTTPStatus.OK:
                break
        except requests.exceptions.ConnectionError:
            time.sleep(0.2)
    else:
        os.killpg(proc.pid, signal.SIGTERM)
        raise RuntimeError("Go server did not start")

    yield base_url

    os.killpg(proc.pid, signal.SIGTERM)
    proc.wait()


@pytest.fixture(scope="function")
def api_url(go_server):
    return f"{go_server}/api"
