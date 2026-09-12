import { ArrowRight, ChevronRight, CircleAlert, CircleCheck, Clock3 } from "lucide-react";
import { useMemo, useState } from "react";
import { PageHeader } from "../../components/PageHeader";
import { PendingState } from "../../components/PendingState";
import { useQuery } from "@tanstack/react-query";

const measurements = ["Temperature", "Voltage", "Battery", "Pressure"];
const headers = ["Device", "Overall state", ...measurements, "Last processed"];

export function FleetPendingPage() {
  return <><PageHeader title="Fleet telemetry" detail="Live aggregate view · Awaiting first refresh" /><section className="status-strip" aria-label="Fleet status pending">{(["Critical", "Warning", "Stale", "Active"] as const).map((status) => <a className="status-metric" href="#fleet-table" key={status}><span>{status}</span><span className="skeleton-value" aria-hidden="true" /></a>)}</section><section className="panel fleet-panel"><div className="toolbar" aria-label="Fleet table controls"><label><span>Search fleet</span><input disabled placeholder="Device ID or name" type="search" /></label><label><span>State</span><select disabled defaultValue="all"><option value="all">All states</option></select></label><label><span>Measurement</span><select disabled defaultValue="all"><option value="all">All measurements</option></select></label><span className="toolbar__result">Awaiting fleet data</span></div><p className="sr-only" id="fleet-table-scroll-help">Scroll horizontally to review all fleet table columns.</p><div aria-describedby="fleet-table-scroll-help" aria-label="Fleet telemetry table" className="table-scroll" id="fleet-table" role="region" tabIndex={0}><table><caption className="sr-only">Fleet telemetry pending table</caption><thead><tr>{headers.map((header) => <th key={header} scope="col">{header}</th>)}</tr></thead><tbody>{Array.from({ length: 7 }, (_, index) => <tr key={index} aria-hidden="true">{headers.map((header) => <td key={header}><span className="skeleton-cell" /></td>)}</tr>)}</tbody></table></div><p className="panel-note">Fleet rows will appear when Redis live aggregates are connected.</p></section></>;
}

type FleetResponse = { mode: string; generated_at: string; summary: Record<string, number>; devices: Array<{ device_id: string; freshness: string; overall_status: string; last_processed_at: string; measurements: Array<{ measurement_type: string; latest_value: number }> }> };
const fetchFleet = async (): Promise<FleetResponse> => { const response = await fetch("/v1/live/fleet"); if (!response.ok) throw new Error("Live aggregates are unavailable"); return response.json(); };
export function FleetLivePage() { const query=useQuery({queryKey:["fleet-live"],queryFn:fetchFleet,refetchInterval:3000}); const [search,setSearch]=useState(""); const [filter,setFilter]=useState("all"); const [sort,setSort]=useState("device"); const data=query.data; const devices=useMemo(()=> (data?.devices??[]).filter(d=>(filter==="all"||d.freshness===filter||d.overall_status===filter)&&d.device_id.toLowerCase().includes(search.toLowerCase())).sort((a,b)=>sort==="device"?a.device_id.localeCompare(b.device_id):a.last_processed_at.localeCompare(b.last_processed_at)),[data,filter,search,sort]); if(!data&&query.isLoading)return <FleetPendingPage/>; if(!data)return <section className="recovery-state"><h1>Live data unavailable</h1><p>Redis live aggregates could not be loaded. No historical data was requested.</p></section>; const icon=(d:string)=>d==="critical"?<CircleAlert aria-hidden size={16}/>:d==="stale"?<Clock3 aria-hidden size={16}/>:<CircleCheck aria-hidden size={16}/>; return <><PageHeader title="Fleet telemetry" detail={`Live aggregate view · Updated ${new Date(data.generated_at).toLocaleTimeString()}`}/><section className="status-strip" aria-label="Fleet status">{["critical","warning","stale","active"].map(s=><button className="status-metric" onClick={()=>setFilter(s)} key={s}>{s}<strong>{data.summary[s]??0}</strong></button>)}</section>{query.isError&&<p className="panel-note">Degraded: showing last successful live snapshot.</p>}<section className="panel fleet-panel"><div className="toolbar"><label>Search fleet<input value={search} onChange={e=>setSearch(e.target.value)} type="search"/></label><label>State<select value={filter} onChange={e=>setFilter(e.target.value)}><option value="all">All states</option>{["critical","warning","stale","active"].map(s=><option key={s}>{s}</option>)}</select></label><button onClick={()=>{setSearch("");setFilter("all")}}>Clear filters</button><span className="toolbar__result">{devices.length} devices</span></div><div className="table-scroll" role="region" aria-label="Fleet telemetry table" tabIndex={0}><table><thead><tr>{headers.map(h=>{const active=(h==="Device"&&sort==="device")||(h==="Last processed"&&sort==="time");return <th key={h} scope="col" aria-sort={active?"ascending":"none"}><button onClick={()=>h==="Last processed"?setSort("time"):setSort("device")}>{h}</button></th>})}</tr></thead><tbody>{devices.length===0?<tr><td colSpan={7}>No devices match the current filters.</td></tr>:devices.map(d=><tr key={d.device_id}><td><a href={`/devices/${encodeURIComponent(d.device_id)}`}>{d.device_id}</a></td><td><span>{icon(d.overall_status)} {d.overall_status} · {d.freshness}</span></td>{measurements.map(m=>{const x=d.measurements.find(v=>v.measurement_type===m.toLowerCase());return <td className="numeric" key={m}>{x?x.latest_value:"—"}</td>})}<td className="numeric">{d.last_processed_at}</td></tr>)}</tbody></table></div></section></>; }

export function DevicePendingPage({ deviceID }: { deviceID: string }) {
  const breadcrumb = <><a href="/">Fleet</a><ChevronRight aria-hidden="true" size={16} /><span>{deviceID}</span></>;
  const action = <a className="button button--primary" href={`/devices/${encodeURIComponent(deviceID)}/history`}>View history <ArrowRight aria-hidden="true" size={16} /></a>;
  return <><PageHeader breadcrumb={breadcrumb} detail="Live aggregate view · Awaiting first device snapshot" title={deviceID} action={action} /><section className="fact-grid" aria-label="Device measurement facts pending">{measurements.map((measurement) => <PendingState key={measurement} label={`${measurement} pending`} rows={3} />)}</section><section className="chart-grid" aria-label="Device aggregate charts pending">{measurements.map((measurement) => <PendingState key={measurement} label={`${measurement} chart pending`} rows={5} />)}</section><p className="source-note">Data source: Redis live aggregate · Awaiting first device snapshot</p></>;
}

type DeviceMeasurement = {
  measurement_type: string;
  unit: string;
  latest_value: number;
  average: number;
  minimum: number;
  maximum: number;
  window_start: string;
  window_end: string;
  threshold_status: string;
};

type DeviceResponse = {
  mode: string;
  generated_at: string;
  device: {
    device_id: string;
    freshness: string;
    overall_status: string;
    last_processed_at: string;
    measurements: DeviceMeasurement[];
  };
};

const deviceLivePath = (deviceID: string) => `/v1/live/devices/${encodeURIComponent(deviceID)}`;

async function fetchDevice(deviceID: string): Promise<DeviceResponse> {
  const response = await fetch(deviceLivePath(deviceID));
  if (response.status === 404) throw new Error("not-found");
  if (response.status === 503) throw new Error("degraded");
  if (!response.ok) throw new Error("unavailable");
  return response.json();
}

function DeviceRecovery({ title, children }: { title: string; children: React.ReactNode }) {
  return <section className="recovery-state" aria-labelledby="device-recovery-title">
    <p className="eyebrow">Navigation recovery</p>
    <h1 id="device-recovery-title">{title}</h1>
    <p>{children}</p>
    <a className="button button--primary" href="/">Return to fleet</a>
  </section>;
}

function statusIcon(status: string) {
  if (status === "critical") return <CircleAlert aria-hidden="true" size={16} />;
  if (status === "stale" || status === "degraded") return <Clock3 aria-hidden="true" size={16} />;
  return <CircleCheck aria-hidden="true" size={16} />;
}

function StatusChip({ status }: { status: string }) {
  return <span className={`status-chip status-chip--${status}`}>{statusIcon(status)}<span>{status}</span></span>;
}

type Range = { start: number; end: number; status: "normal" | "warning" | "critical" };
function rangesFor(measurement: DeviceMeasurement): Range[] {
  if (measurement.measurement_type === "temperature") return [{ start: -40, end: 70, status: "normal" }, { start: 70, end: 85, status: "warning" }, { start: 85, end: 125, status: "critical" }];
  if (measurement.measurement_type === "voltage") return [{ start: 0, end: 10.5, status: "critical" }, { start: 10.5, end: 11.5, status: "warning" }, { start: 11.5, end: 48, status: "normal" }];
  if (measurement.measurement_type === "battery") return [{ start: 0, end: 10, status: "critical" }, { start: 10, end: 20, status: "warning" }, { start: 20, end: 100, status: "normal" }];
  return [{ start: 0, end: 300, status: "normal" }, { start: 300, end: 700, status: "warning" }, { start: 700, end: 1000, status: "critical" }];
}

function StatusRangeChart({ measurement }: { measurement: DeviceMeasurement }) {
  const ranges = rangesFor(measurement);
  const domainStart = ranges[0].start;
  const domainEnd = ranges[ranges.length - 1].end;
  const x = (value: number) => Math.max(0, Math.min(100, ((value - domainStart) / (domainEnd - domainStart)) * 100));
  return <figure className="aggregate-chart" aria-labelledby={`${measurement.measurement_type}-chart-title`}>
    <svg viewBox="0 0 100 24" role="img" aria-labelledby={`${measurement.measurement_type}-chart-title ${measurement.measurement_type}-chart-description`}>
      <title id={`${measurement.measurement_type}-chart-title`}>{measurement.measurement_type} aggregate range chart</title>
      <desc id={`${measurement.measurement_type}-chart-description`}>Threshold bands with minimum, average, maximum, and latest values for the current five-minute window.</desc>
      {ranges.map((range) => <rect className={`aggregate-chart__band aggregate-chart__band--${range.status}`} height="10" key={range.status} width={x(range.end) - x(range.start)} x={x(range.start)} y="7" />)}
      <line className="aggregate-chart__range" x1={x(measurement.minimum)} x2={x(measurement.maximum)} y1="12" y2="12" />
      <line className="aggregate-chart__average" x1={x(measurement.average)} x2={x(measurement.average)} y1="3" y2="21" />
      <circle className="aggregate-chart__latest" cx={x(measurement.latest_value)} cy="12" r="3" />
    </svg>
    <figcaption>Threshold bands: Normal, Warning, Critical. Range: min—max; line: average; point: latest.</figcaption>
  </figure>;
}

function MeasurementFact({ measurement }: { measurement: DeviceMeasurement }) {
  return <section className="pending-panel" aria-label={`${measurement.measurement_type} live measurement`}>
    <strong>{measurement.measurement_type}</strong>
    <span className="numeric">{measurement.latest_value} {measurement.unit}</span>
    <StatusChip status={measurement.threshold_status} />
    <span>Freshness uses processing time</span>
  </section>;
}

function AggregateChart({ measurement }: { measurement: DeviceMeasurement }) {
  const summary = `Latest ${measurement.latest_value} ${measurement.unit}; ${measurement.threshold_status}; average ${measurement.average} ${measurement.unit}; minimum ${measurement.minimum} ${measurement.unit}; maximum ${measurement.maximum} ${measurement.unit}; window ${measurement.window_start} to ${measurement.window_end}.`;
  return <section className="pending-panel" aria-labelledby={`${measurement.measurement_type}-summary`}>
    <h2 id={`${measurement.measurement_type}-summary`}>{measurement.measurement_type} five-minute summary</h2>
    <p className="numeric">Latest {measurement.latest_value} {measurement.unit}</p>
    <StatusRangeChart measurement={measurement} />
    <p>Average {measurement.average} · Min {measurement.minimum} · Max {measurement.maximum} {measurement.unit}</p>
    <p>{measurement.threshold_status} · Window {measurement.window_start} to {measurement.window_end}</p>
    <p className="sr-only">{summary}</p>
  </section>;
}

export function DeviceLivePage({ deviceID }: { deviceID: string }) {
  const query = useQuery({ queryKey: ["device-live", deviceID], queryFn: () => fetchDevice(deviceID), refetchInterval: 3000 });

  if (!query.data && query.isLoading) return <DevicePendingPage deviceID={deviceID} />;
  if (!query.data) {
    const isNotFound = query.error instanceof Error && query.error.message === "not-found";
    const isDegraded = query.error instanceof Error && query.error.message === "degraded";
    return <DeviceRecovery title={isNotFound ? "Device not found" : isDegraded ? "Degraded mode: live aggregates paused" : "Device data could not be loaded"}>
      {isNotFound ? "This device has no current live aggregate." : isDegraded ? "Redis is offline. No historical data was requested." : "Live Redis data is unavailable. No historical data was requested."}
    </DeviceRecovery>;
  }

  const device = query.data.device;
  if (device.freshness === "unknown" || device.measurements.length === 0) {
    return <DeviceRecovery title="No live aggregate for this device">
      This device has not produced a current Redis-backed aggregate. No historical data was requested.
    </DeviceRecovery>;
  }

  const isStale = device.freshness === "stale";
  const isDegraded = query.isError || query.data.mode === "degraded";
  const detail = `Last processed ${device.last_processed_at}`;
  return <>
    <PageHeader
      breadcrumb={<><a href="/">Fleet</a><ChevronRight aria-hidden size={16} /><span>{device.device_id}</span></>}
      title={device.device_id}
      detail={detail}
      action={<span className="header-actions"><StatusChip status={device.overall_status} /><StatusChip status={isStale || isDegraded ? "stale" : device.freshness} /><a className="button button--primary" href={`/devices/${encodeURIComponent(device.device_id)}/history`}>View history</a></span>}
    />
    {isDegraded && <p className="degraded-banner" role="status">Degraded mode: live aggregates paused (Redis offline). Showing the last successful snapshot.</p>}
    {isStale && !isDegraded && <p className="panel-note" role="status">Stale: this device has not processed telemetry within 60 seconds.</p>}
    <section className="fact-grid" aria-label="Live measurements">{device.measurements.map((measurement) => <MeasurementFact key={measurement.measurement_type} measurement={measurement} />)}</section>
    <section className="chart-grid" aria-label="Five-minute aggregate summaries">{device.measurements.map((measurement) => <AggregateChart key={measurement.measurement_type} measurement={measurement} />)}</section>
    <p className="source-note">Data source: Redis live aggregate · Last live refresh {query.data.generated_at}</p>
  </>;
}
