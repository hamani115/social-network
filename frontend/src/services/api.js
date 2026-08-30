const API_BASE = "/api";

export class ApiError extends Error {
  constructor(message, { status = 0, type = "api", path = "" } = {}) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.type = type;
    this.path = path;
  }
}

export function isServerUnavailableError(error) {
  return (
    error instanceof ApiError &&
    (error.type === "network" ||
      error.status === 502 ||
      error.status === 503 ||
      error.status === 504)
  );
}

function dispatchGlobalApiError(detail) {
  if (typeof window === "undefined") {
    return;
  }

  window.dispatchEvent(
    new CustomEvent("api-global-error", {
      detail,
    }),
  );
}

function defaultErrorMessage(status) {
  switch (status) {
    case 400:
      return "The request was invalid";

    case 401:
      return "Your session has expired";

    case 403:
      return "You do not have permission to do that";

    case 404:
      return "The requested resource was not found";

    case 409:
      return "The request conflicts with the current state";

    case 500:
      return "Something went wrong on the server";

    case 502:
    case 503:
    case 504:
      return "The server is currently unavailable";

    default:
      return "Something went wrong";
  }
}

export async function apiRequest(path, options = {}) {
  const isFormData = options.body instanceof FormData;

  let response;

  try {
    response = await fetch(`${API_BASE}${path}`, {
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
  } catch {
    const error = new ApiError("Could not connect to the server", {
      status: 0,
      type: "network",
      path,
    });

    dispatchGlobalApiError({
      type: "server-unavailable",
      status: 0,
      path,
      message: error.message,
    });

    throw error;
  }

  let data = null;

  const contentType = response.headers.get("content-type") || "";

  if (response.status !== 204 && contentType.includes("application/json")) {
    try {
      data = await response.json();
    } catch {
      data = null;
    }
  }

  if (!response.ok) {
    const message = data?.error || defaultErrorMessage(response.status);

    const error = new ApiError(message, {
      status: response.status,
      type: "http",
      path,
    });

    if (response.status === 401 && path !== "/login") {
      dispatchGlobalApiError({
        type: "unauthorized",
        status: response.status,
        path,
        message,
      });
    }

    if (
      response.status === 502 ||
      response.status === 503 ||
      response.status === 504
    ) {
      dispatchGlobalApiError({
        type: "server-unavailable",
        status: response.status,
        path,
        message,
      });
    }

    else if (response.status >= 500) {
      dispatchGlobalApiError({
        type: "generic",
        status: response.status,
        path,
        message,
      });
    }

    throw error;
  }

  return data;
}
