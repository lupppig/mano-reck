import Link from "next/link";

export default function NotFound() {
  return (
    <section className="state-panel" aria-labelledby="not-found-title">
      <p className="eyebrow">Page not found</p>
      <h1 id="not-found-title">This console route does not exist.</h1>
      <p>
        The address may be outdated, or the workspace may not expose this view
        yet.
      </p>
      <Link className="primary-action" href="/">
        Return to foundation
      </Link>
    </section>
  );
}
