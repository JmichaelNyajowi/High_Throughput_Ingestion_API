export const bootstrapTitle = "Application foundation ready";

export function App() {
  return (
    <main className="bootstrap-shell" aria-labelledby="bootstrap-title">
      <p className="eyebrow">Telemetry Operations</p>
      <h1 id="bootstrap-title">{bootstrapTitle}</h1>
      <p>
        Fleet monitoring screens are delivered through the application-shell ticket.
      </p>
    </main>
  );
}
