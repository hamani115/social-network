const API_BASE = "/api";

export async function apiRequest(path, options = {}) {
  const isFormData = options.body instanceof FormData;

  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    credentials: "include",
    headers: isFormData
      ? {
        ...(options.headers || {}),
      }
      : {
        "Content-Type": "application/json",
        ...(options.headers || {}),
      },
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.error || "Something went wrong");
  }

  return data;
}