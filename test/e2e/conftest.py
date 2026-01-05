import pytest
import subprocess
import time
from http import HTTPStatus
import requests
import os
import signal
import socket


def get_free_port():
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


@pytest.fixture(scope="function")
def go_server():
    port = get_free_port()
    api_root = f"http://127.0.0.1:{port}/api"

    proc = subprocess.Popen(
        [
            "go",
            "run",
            "cmd/server/main.go",
            "--seed-admin=true",
            "--bcrypt-cost=5",
            "--addr",
            f":{port}",
        ],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        start_new_session=True,
    )

    for _ in range(30):
        try:
            if requests.get(f"{api_root}/health").status_code == HTTPStatus.OK:
                break
        except requests.exceptions.ConnectionError:
            time.sleep(0.2)
    else:
        os.killpg(proc.pid, signal.SIGTERM)
        raise RuntimeError("Go server did not start")

    yield api_root

    os.killpg(proc.pid, signal.SIGTERM)
    proc.wait()
