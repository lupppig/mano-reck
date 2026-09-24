export default function Loading() {
  return (
    <section
      className="state-panel"
      aria-labelledby="loading-title"
      aria-live="polite"
    >
      <span className="state-indicator" aria-hidden="true" />
      <p className="eyebrow">Loading workspace</p>
      <h1 id="loading-title">Preparing the console.</h1>
      <p>We are assembling the latest platform view for this request.</p>
    </section>
  );
}
