function handleResponse(response) {
    if (response.status === 401) {
        window.location.href = "/login";
        return null;
    }
    return response;
}

async function handleError(response) {
    if (response.status === 404) {
        throw new Error("Not found");
    }
    if (response.status === 403) {
        throw new Error("Permission denied");
    }
    if (response.status === 409) {
        throw new Error("Conflict - please refresh the page");
    }
    const text = await response.text();
    throw new Error(text || "Request failed");
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
        if (data !== undefined) {
            options.headers = { "Content-Type": "application/json" };
            options.body = JSON.stringify(data);
        }
        const response = await fetch(path, options);
        const handled = handleResponse(response);
        if (!handled) return null;
        if (!response.ok) {
            await handleError(response);
        }
        return response;
    },

    async patch(path, data) {
        const response = await fetch(path, {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
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
        const response = await fetch(path, {
            method: "DELETE",
        });
        const handled = handleResponse(response);
        if (!handled) return null;
        if (!response.ok) {
            await handleError(response);
        }
        return response;
    },
};

export default api;
