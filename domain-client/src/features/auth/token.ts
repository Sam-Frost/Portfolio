const TOKEN_KEY = "domain_auth_token";

// Set right after a successful login, consumed once by DomainLayout to
// decide whether to show the one-shot TimeLeftClockModal for this session.
const JUST_LOGGED_IN_KEY = "domain_just_logged_in";

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY);
  sessionStorage.removeItem(JUST_LOGGED_IN_KEY);
}

export function markJustLoggedIn(): void {
  sessionStorage.setItem(JUST_LOGGED_IN_KEY, "1");
}

// Returns whether this session just logged in, clearing the flag so it only
// fires once.
export function consumeJustLoggedIn(): boolean {
  const flagged = sessionStorage.getItem(JUST_LOGGED_IN_KEY) === "1";
  if (flagged) sessionStorage.removeItem(JUST_LOGGED_IN_KEY);
  return flagged;
}
