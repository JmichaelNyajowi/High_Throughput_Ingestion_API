import { ArrowLeft } from "lucide-react";
import { useEffect, useState } from "react";
import { AppShell } from "../components/AppShell";
import { DeviceLivePage, FleetLivePage } from "../modules/fleet";
import { HistoryPreQueryPage } from "../modules/history";
import { resolveRoute } from "./routes";

export function App({ initialPath }: { initialPath?: string }) {
  const [pathname, setPathname] = useState(initialPath ?? window.location.pathname);
  useEffect(() => { const syncPath = () => setPathname(window.location.pathname); window.addEventListener("popstate", syncPath); return () => window.removeEventListener("popstate", syncPath); }, []);
  const route = resolveRoute(pathname);
  const content = route.kind === "fleet" ? <FleetLivePage /> : route.kind === "device" ? <DeviceLivePage deviceID={route.deviceID} /> : route.kind === "history" ? <HistoryPreQueryPage deviceID={route.deviceID} /> : <NotFoundPage />;
  return <AppShell isFleet={route.kind === "fleet"}>{content}</AppShell>;
}

function NotFoundPage() { return <section className="recovery-state" aria-labelledby="not-found-title"><p className="eyebrow">Navigation recovery</p><h1 id="not-found-title">Page not found</h1><p>The destination is unavailable or no longer exists.</p><a className="button button--primary" href="/"><ArrowLeft aria-hidden="true" size={16} /> Return to fleet</a></section>; }
