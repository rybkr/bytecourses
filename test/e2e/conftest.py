import pytest
import subprocess
import time
from http import HTTPStatus
import requests
import os
import signal

API_ROOT: str = "http://localhost:8080/api"


@pytest.fixture(scope="session")
def go_server():
    proc = subprocess.Popen(
        ["go", "run", "cmd/server/main.go", "--seed-admin=true"],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        start_new_session=True,
    )

    for _ in range(30):
        try:
            if requests.get(f"{API_ROOT}/health").status_code == HTTPStatus.OK:
                break
        except requests.exceptions.ConnectionError:
            time.sleep(0.2)
    else:
        os.killpg(proc.pid, signal.SIGTERM)
        raise RuntimeError("Go server did not start")

    yield

    os.killpg(proc.pid, signal.SIGTERM)
    proc.wait()
