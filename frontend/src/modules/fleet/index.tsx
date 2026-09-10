import { ArrowRight, ChevronRight } from "lucide-react";
import { PageHeader } from "../../components/PageHeader";
import { PendingState } from "../../components/PendingState";

const measurements = ["Temperature", "Voltage", "Battery", "Pressure"];
const headers = ["Device", "Overall state", ...measurements, "Last processed"];

export function FleetPendingPage() {
  return <><PageHeader title="Fleet telemetry" detail="Live aggregate view · Awaiting first refresh" /><section className="status-strip" aria-label="Fleet status pending">{(["Critical", "Warning", "Stale", "Active"] as const).map((status) => <a className="status-metric" href="#fleet-table" key={status}><span>{status}</span><span className="skeleton-value" aria-hidden="true" /></a>)}</section><section className="panel fleet-panel"><div className="toolbar" aria-label="Fleet table controls"><label><span>Search fleet</span><input disabled placeholder="Device ID or name" type="search" /></label><label><span>State</span><select disabled defaultValue="all"><option value="all">All states</option></select></label><label><span>Measurement</span><select disabled defaultValue="all"><option value="all">All measurements</option></select></label><span className="toolbar__result">Awaiting fleet data</span></div><p className="sr-only" id="fleet-table-scroll-help">Scroll horizontally to review all fleet table columns.</p><div aria-describedby="fleet-table-scroll-help" aria-label="Fleet telemetry table" className="table-scroll" id="fleet-table" role="region" tabIndex={0}><table><caption className="sr-only">Fleet telemetry pending table</caption><thead><tr>{headers.map((header) => <th key={header} scope="col">{header}</th>)}</tr></thead><tbody>{Array.from({ length: 7 }, (_, index) => <tr key={index} aria-hidden="true">{headers.map((header) => <td key={header}><span className="skeleton-cell" /></td>)}</tr>)}</tbody></table></div><p className="panel-note">Fleet rows will appear when Redis live aggregates are connected.</p></section></>;
}

export function DevicePendingPage({ deviceID }: { deviceID: string }) {
  const breadcrumb = <><a href="/">Fleet</a><ChevronRight aria-hidden="true" size={16} /><span>{deviceID}</span></>;
  const action = <a className="button button--primary" href={`/devices/${encodeURIComponent(deviceID)}/history`}>View history <ArrowRight aria-hidden="true" size={16} /></a>;
  return <><PageHeader breadcrumb={breadcrumb} detail="Live aggregate view · Awaiting first device snapshot" title={deviceID} action={action} /><section className="fact-grid" aria-label="Device measurement facts pending">{measurements.map((measurement) => <PendingState key={measurement} label={`${measurement} pending`} rows={3} />)}</section><section className="chart-grid" aria-label="Device aggregate charts pending">{measurements.map((measurement) => <PendingState key={measurement} label={`${measurement} chart pending`} rows={5} />)}</section><p className="source-note">Data source: Redis live aggregate · Awaiting first device snapshot</p></>;
}
