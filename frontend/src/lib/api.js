// Talks to the Go server's JSON API. Every page's data comes from GET /api/<page
// path>; forms POST to /api/<form path> and answer {"redirect": "/where/next"}.

export class ApiError extends Error {
  constructor(status, data, text) {
    super((data && data.error) || text || `Request failed (${status})`);
    this.status = status;
    this.data = data;
  }
}

async function request(path, { method = 'GET', body, signal } = {}) {
  const opts = { method, credentials: 'same-origin', headers: { Accept: 'application/json' }, signal };
  if (body instanceof FormData) opts.body = body;
  else if (body) opts.body = new URLSearchParams(body);
  const res = await fetch('/api' + path, opts);
  const text = await res.text();
  let data = null;
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    /* plain-text error from the server */
  }
  if (!res.ok) throw new ApiError(res.status, data, text.trim());
  return data;
}

export const get = (path, signal) => request(path, { signal });

// post sends a form (an HTMLFormElement, FormData or plain object).
export function post(path, form) {
  const body = form instanceof HTMLFormElement ? new FormData(form) : form || {};
  return request(path, { method: 'POST', body });
}
