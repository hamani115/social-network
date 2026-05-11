const API_ENDPOINT = "http://localhost:8080/api";

export async function apiRequest(path, options = {}) {
  const response = await fetch(`${API_ENDPOINT}${path}`, {
    ...options,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.error || "error...");
  }

  return data;
}