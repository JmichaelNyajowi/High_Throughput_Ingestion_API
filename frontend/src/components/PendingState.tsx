export function PendingState({ rows = 6, label }: { rows?: number; label: string }) {
  return <section className="pending-panel" aria-label={label} aria-busy="true"><span className="sr-only">{label}</span>{Array.from({ length: rows }, (_, index) => <span className="skeleton-line" key={index} />)}</section>;
}
