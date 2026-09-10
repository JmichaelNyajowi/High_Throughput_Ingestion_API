import { ChevronRight } from "lucide-react";
import { PageHeader } from "../../components/PageHeader";

export function HistoryPreQueryPage({ deviceID }: { deviceID: string }) {
  const breadcrumb = <><a href="/">Fleet</a><ChevronRight aria-hidden="true" size={16} /><a href={`/devices/${encodeURIComponent(deviceID)}`}>{deviceID}</a><ChevronRight aria-hidden="true" size={16} /><span>History</span></>;
  return <><PageHeader breadcrumb={breadcrumb} detail="Manual query — not live data" title={`History for ${deviceID}`} /><section className="panel history-query" aria-labelledby="history-query-title"><h2 id="history-query-title">Choose a time range</h2><p>History is queried only after you choose a valid range and run the query.</p><form><label><span>Start time</span><input type="datetime-local" /></label><label><span>End time</span><input type="datetime-local" /></label><label><span>Measurement</span><select defaultValue="all"><option value="all">All measurements</option></select></label><button className="button button--primary" disabled type="submit">Run query</button></form><p className="panel-note">Select a valid range to enable the manual query.</p></section></>;
}
