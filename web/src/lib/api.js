// web/src/lib/api.js
function getCookie(name) {
    return (
        document.cookie
            .split("; ")
            .find((row) => row.startsWith(name + "="))
            ?.split("=")[1] ?? ""
    );
}

async function request(path, { method = "GET", headers = {}, body } = {}) {
    const isMutation = !["GET", "HEAD"].includes(method.toUpperCase());
    const finalHeaders = new Headers(headers);
    finalHeaders.set("Accept", "application/json");

    // JSON body for mutations unless caller sets something else
    let payload = body;
    if (isMutation && body && !(body instanceof FormData)) {
        finalHeaders.set("Content-Type", "application/json");
        payload = JSON.stringify(body);
    }

    // CSRF for mutations if present
    if (isMutation) {
        const csrf = getCookie("csrf");
        if (csrf) finalHeaders.set("X-CSRF-Token", csrf);
    }

    const res = await fetch(path, {
        method,
        headers: finalHeaders,
        body: payload,
        credentials: "include", // send cookies (sid/csrf)
    });

    // Try to parse JSON if server says it's JSON
    const ctype = res.headers.get("content-type") || "";
    const looksJson = ctype.toLowerCase().startsWith("application/json");
    const data = looksJson ? await res.json().catch(() => null) : await res.text();

    if (!res.ok) {
        const msg = (data && data.error) || (typeof data === "string" ? data : "") || `HTTP ${res.status}`;
        const err = new Error(msg);
        err.status = res.status;
        err.data = data;
        throw err;
    }
    return data;
}

export const api = {
    get: (path, opts = {}) => request(path, { ...opts, method: "GET" }),
    post: (path, body, opts = {}) => request(path, { ...opts, method: "POST", body }),
    put: (path, body, opts = {}) => request(path, { ...opts, method: "PUT", body }),
    patch: (path, body, opts = {}) => request(path, { ...opts, method: "PATCH", body }),
    del: (path, opts = {}) => request(path, { ...opts, method: "DELETE" }),
};
