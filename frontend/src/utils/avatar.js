export function nameInitials(name) {
  if (!name) {
    return "?";
  }
  return (
    name
      .trim()
      .split(/\s+/)
      .slice(0, 2)
      .map((part) => part.charAt(0))
      .join("")
      .toUpperCase() || "?"
  );
}
export function userInitials(user) {
  if (!user) {
    return "?";
  }
  return nameInitials(`${user.first_name || ""} ${user.last_name || ""}`);
}
