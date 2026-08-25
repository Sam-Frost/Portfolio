import { API_BASE_URL } from "../config";
import { clearToken, getToken } from "../features/auth/token";

// Thrown by apiRequest for every failure case (bad status, malformed body,
// or the fetch itself throwing) so callers get one type to check against
// and, where the backend sent {"error": "..."}, its actual message instead
// of a generic status string.
export class ApiError extends Error {
  status: number | null;

  constructor(message: string, status: number | null) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

async function messageFromResponse(res: Response, method: string, path: string): Promise<string> {
  try {
    const body = (await res.json()) as { error?: string };
    if (body.error) return body.error;
  } catch {
    // Body wasn't JSON (or was empty) — fall back to a generic message below.
  }
  return `${method} ${path} failed: ${res.status}`;
}

// Shared by every domain feature's api.ts (todos, settings, ...): attaches
// the bearer token from login and, on a 401, clears it and bounces back to
// the password gate rather than leaving the caller to figure out why every
// request just started failing.
export async function apiRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getToken();
  const method = init?.method ?? "GET";

  let res: Response;
  try {
    res = await fetch(`${API_BASE_URL}${path}`, {
      ...init,
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...init?.headers,
      },
    });
  } catch {
    throw new ApiError("Couldn't reach the server. Check your connection and try again.", null);
  }

  if (res.status === 401) {
    clearToken();
    window.location.href = "/";
    throw new ApiError("Session expired. Please log in again.", 401);
  }

  if (!res.ok) {
    throw new ApiError(await messageFromResponse(res, method, path), res.status);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}
