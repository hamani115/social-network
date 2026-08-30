// Date and time
export function formatDateTime(value) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}
export function formatDate(value) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}
// date of birth
export function formatDateOfBirth(value) {
  if (!value) {
    return "";
  }
  const [year, month, day] = value.split("-").map(Number);
  const date = new Date(year, month - 1, day);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

// Group events
function eventDateObject(value) {
  if (!value) {
    return null;
  }
  return new Date(value.replace(" ", "T"));
}

export function formatEventDate(value) {
  const date = eventDateObject(value);
  if (!date || Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleDateString([], {
    weekday: "short",
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}
export function formatEventClock(value) {
  const date = eventDateObject(value);
  if (!date || Number.isNaN(date.getTime())) {
    return "";
  }
  return date.toLocaleTimeString([], {
    hour: "numeric",
    minute: "2-digit",
  });
}
export function formatEventTimeForBackend(value) {
  if (!value) {
    return "";
  }
  const formatted = value.replace("T", " ");
  if (formatted.length === 16) {
    return formatted + ":00";
  }
  return formatted;
}
export function currentLocalDateTimeForBackend() {
  const now = new Date();
  const localTime = new Date(now.getTime() - now.getTimezoneOffset() * 60_000);
  return localTime.toISOString().slice(0, 19).replace("T", " ");
}

// Notifications
function notificationDate(value) {
  if (!value) {
    return null;
  }
  let normalized = String(value).trim().replace(" ", "T");
  const hasTimezone = /(?:Z|[+-]\d{2}:\d{2})$/i.test(normalized);
  if (!hasTimezone) {
    normalized += "Z";
  }
  const date = new Date(normalized);
  if (Number.isNaN(date.getTime())) {
    return null;
  }
  return date;
}

export function formatNotificationTime(value) {
  const date = notificationDate(value);
  if (!date) {
    return value || "";
  }
  const now = new Date();
  const differenceSeconds = Math.max(
    0,
    Math.floor((now.getTime() - date.getTime()) / 1000),
  );
  if (differenceSeconds < 60) {
    return "Just now";
  }
  if (differenceSeconds < 3600) {
    return Math.floor(differenceSeconds / 60) + "m ago";
  }
  if (differenceSeconds < 86400) {
    return Math.floor(differenceSeconds / 3600) + "h ago";
  }
  if (differenceSeconds < 172800) {
    return "Yesterday";
  }
  return date.toLocaleDateString([], {
    month: "short",
    day: "numeric",
    year: date.getFullYear() !== now.getFullYear() ? "numeric" : undefined,
  });
}
// date inputs
export function todayDateInput() {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}
