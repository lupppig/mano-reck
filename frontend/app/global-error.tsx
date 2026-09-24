"use client";

export default function GlobalError({
  reset,
}: Readonly<{ error: Error & { digest?: string }; reset: () => void }>) {
  return (
    <html lang="en">
      <body>
        <main className="global-fallback">
          <section
            className="state-panel"
            aria-labelledby="global-error-title"
            role="alert"
          >
            <p className="eyebrow">Application unavailable</p>
            <h1 id="global-error-title">The console needs a fresh start.</h1>
            <p>
              No action was completed. Retry now, and contact support if the
              problem continues.
            </p>
            <button className="primary-action" type="button" onClick={reset}>
              Reload console
            </button>
          </section>
        </main>
      </body>
    </html>
  );
}
