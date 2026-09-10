import { Activity, CircleAlert, Radio } from "lucide-react";
import type { PropsWithChildren } from "react";

export function AppShell({ children }: PropsWithChildren) {
  return <div className="app-shell"><a className="skip-link" href="#main-content">Skip to page content</a><header className="top-bar"><div className="content-frame top-bar__content"><a className="brand" href="/" aria-label="Telemetry Operations Fleet"><Activity aria-hidden="true" size={20} /><span>Telemetry Operations</span></a><nav aria-label="Primary navigation"><a aria-current="page" href="/">Fleet</a></nav><div className="top-bar__status" aria-label="System status awaiting live data"><Radio aria-hidden="true" size={16} /><span>Awaiting live data</span><span className="operator-indicator"><CircleAlert aria-hidden="true" size={16} /> Operator access</span></div></div></header><main id="main-content" className="content-frame page-content" tabIndex={-1}>{children}</main></div>;
}
