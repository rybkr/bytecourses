export async function extractErrorMessage(response) {
    const contentType = response.headers.get("Content-Type") || "";
    let message = "";
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
            message = text || "";
        }
    } else {
        message = text || "";
    }

    return message;
}

export function parseErrorFromText(text, contentType) {
    let message = "";

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
            message = text || "";
        }
    } else {
        message = text || "";
    }

    return message;
}
