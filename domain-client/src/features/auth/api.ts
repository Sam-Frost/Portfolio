import { API_BASE_URL } from "../../config";

export async function login(password: string): Promise<string> {
  const res = await fetch(`${API_BASE_URL}/api/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ password }),
  });

  if (!res.ok) {
    throw new Error(res.status === 401 ? "Incorrect password." : "Login failed, try again.");
  }

  const { token } = (await res.json()) as { token: string };
  return token;
}
