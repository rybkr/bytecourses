/**
 * Extracts a user-friendly error message from an API response.
 * Handles JSON error responses with `error` or `errors` fields.
 *
 * @param {Response} response - The fetch Response object
 * @returns {Promise<string>} The extracted error message
 */
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

/**
 * Extracts error message from already-parsed response text.
 * Use this when you've already consumed the response body.
 *
 * @param {string} text - The response body text
 * @param {string} contentType - The Content-Type header value
 * @returns {string} The extracted error message
 */
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
