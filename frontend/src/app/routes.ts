export type Route =
  | { kind: "fleet" }
  | { kind: "device"; deviceID: string }
  | { kind: "history"; deviceID: string }
  | { kind: "not-found" };

export function resolveRoute(pathname: string): Route {
  const segments = pathname.split("/").filter(Boolean);
  if (segments.length === 0) return { kind: "fleet" };
  if (segments[0] !== "devices" || !segments[1] || segments.length > 3) return { kind: "not-found" };
  const deviceID = decodeURIComponent(segments[1]);
  if (segments.length === 2) return { kind: "device", deviceID };
  return segments[2] === "history" ? { kind: "history", deviceID } : { kind: "not-found" };
}
