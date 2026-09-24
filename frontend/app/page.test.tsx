import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";

import Home from "./page";

describe("enterprise shell home", () => {
  it("explains the platform without claiming unavailable product behavior", () => {
    const markup = renderToStaticMarkup(<Home />);

    expect(markup).toContain(
      "Operational foundations, ready for product work.",
    );
    expect(markup).toContain(
      "Authentication, tenant selection, permissions, and product navigation",
    );
    expect(markup).not.toContain("Sign in");
  });
});
