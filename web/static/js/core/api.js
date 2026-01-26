function handleResponse(response) {
    if (response.status === 401) {
        const next = encodeURIComponent(
            window.location.pathname + window.location.search,
        );
        window.location.href = `/login?next=${next}`;
        return null;
    }
    return response;
}

async function handleError(response) {
    if (response.status === 404) {
        throw new Error("Not found");
    }
    if (response.status === 403) {
        const text = await response.text();
        if (text.includes("CSRF")) {
            throw new Error("CSRF token validation failed - please refresh the page");
        }
        throw new Error("Permission denied");
    }
    if (response.status === 409) {
        throw new Error("Conflict - please refresh the page");
    }
    
    const contentType = response.headers.get("Content-Type") || "";
    let message = "Request failed";
    const text = await response.text();
    
    if (contentType.includes("application/json") && text) {
        try {
            const data = JSON.parse(text);
            if (data.error) {
                message = data.error;
            } else if (data.errors && Array.isArray(data.errors) && data.errors.length > 0) {
                const parts = data.errors.map((e) => {
                    const m = e.Message || e.message || String(e);
                    return e.Field ? `${e.Field}: ${m}` : m;
                });
                message = parts.join("; ");
            }
        } catch (_) {
            message = text || message;
        }
    } else {
        message = text || message;
    }
    
    throw new Error(message);
}

function getCSRFToken() {
    const name = "csrf-token=";
    const decodedCookie = decodeURIComponent(document.cookie);
    const cookieArray = decodedCookie.split(";");
    for (let i = 0; i < cookieArray.length; i++) {
        let cookie = cookieArray[i];
        while (cookie.charAt(0) === " ") {
            cookie = cookie.substring(1);
        }
        if (cookie.indexOf(name) === 0) {
            return cookie.substring(name.length, cookie.length);
        }
    }
    return "";
}

const api = {
    async get(path) {
        const response = await fetch(path);
        const handled = handleResponse(response);
        if (!handled) return null;
        if (!response.ok) {
            await handleError(response);
        }
        return response;
    },

    async post(path, data) {
        const options = {
            method: "POST",
        };
        const headers = {};
        if (data !== undefined) {
            headers["Content-Type"] = "application/json";
            options.body = JSON.stringify(data);
        }
        const csrfToken = getCSRFToken();
        if (csrfToken) {
            headers["X-CSRF-Token"] = csrfToken;
        }
        options.headers = headers;
        const response = await fetch(path, options);
        const handled = handleResponse(response);
        if (!handled) return null;
        if (!response.ok) {
            await handleError(response);
        }
        return response;
    },

    async patch(path, data) {
        const headers = { "Content-Type": "application/json" };
        const csrfToken = getCSRFToken();
        if (csrfToken) {
            headers["X-CSRF-Token"] = csrfToken;
        }
        const response = await fetch(path, {
            method: "PATCH",
            headers: headers,
            body: JSON.stringify(data),
        });
        const handled = handleResponse(response);
        if (!handled) return null;
        if (!response.ok) {
            await handleError(response);
        }
        return response;
    },

    async delete(path) {
        const headers = {};
        const csrfToken = getCSRFToken();
        if (csrfToken) {
            headers["X-CSRF-Token"] = csrfToken;
        }
        const options = {
            method: "DELETE",
        };
        if (Object.keys(headers).length > 0) {
            options.headers = headers;
        }
        const response = await fetch(path, options);
        const handled = handleResponse(response);
        if (!handled) return null;
        if (!response.ok) {
            await handleError(response);
        }
        return response;
    },
};

export default api;
