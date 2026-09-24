"use client";

import { useEffect } from "react";

export default function ErrorPage({
  error,
  reset,
}: Readonly<{ error: Error & { digest?: string }; reset: () => void }>) {
  useEffect(() => {
    console.error("route rendering failed", { digest: error.digest });
  }, [error]);

  return (
    <section className="state-panel" aria-labelledby="error-title" role="alert">
      <p className="eyebrow">Console unavailable</p>
      <h1 id="error-title">This view could not be loaded.</h1>
      <p>
        Your request is safe. Retry the view, or use the correlation reference
        from the response when contacting support.
      </p>
      <button className="primary-action" type="button" onClick={reset}>
        Try again
      </button>
    </section>
  );
}
