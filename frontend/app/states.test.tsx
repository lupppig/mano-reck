import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";

import ErrorPage from "./error";
import Loading from "./loading";
import NotFound from "./not-found";

describe("global route states", () => {
  it("announces loading without presenting an empty screen", () => {
    const markup = renderToStaticMarkup(<Loading />);
    expect(markup).toContain('aria-live="polite"');
    expect(markup).toContain("Preparing the console.");
  });

  it("offers a safe retry without rendering raw error details", () => {
    const markup = renderToStaticMarkup(
      <ErrorPage
        error={new Error("database password leaked")}
        reset={() => undefined}
      />,
    );
    expect(markup).toContain("Try again");
    expect(markup).not.toContain("database password leaked");
  });

  it("provides a route back from unknown pages", () => {
    const markup = renderToStaticMarkup(<NotFound />);
    expect(markup).toContain("This console route does not exist.");
    expect(markup).toContain('href="/"');
  });
});
