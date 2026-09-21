import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";

import Home from "./page";

describe("foundation shell", () => {
  it("explains the platform without claiming unavailable product behavior", () => {
    const markup = renderToStaticMarkup(<Home />);

    expect(markup).toContain("Credit infrastructure, built deliberately.");
    expect(markup).toContain(
      "The local runtime foundation is being assembled now.",
    );
  });
});
