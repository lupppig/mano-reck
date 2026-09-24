import { describe, expect, it, vi } from "vitest";

import {
  BackendClient,
  BackendRequestError,
  createCorrelationId,
} from "./backend-client";

const correlationId = "01991a4b-fa00-7000-8000-000000000003";

describe("backend client", () => {
  it("propagates one UUIDv7 correlation ID", async () => {
    const fetchStub = vi.fn<typeof fetch>(async (_input, init) => {
      const headers = new Headers(init?.headers);
      expect(headers.get("X-Correlation-ID")).toBe(correlationId);
      expect(headers.get("Accept")).toBe("application/json");
      return Response.json({ status: "ok" });
    });
    const client = new BackendClient({
      baseUrl: "http://backend:8080",
      fetch: fetchStub,
      newCorrelationId: () => correlationId,
    });

    await expect(
      client.request<{ status: string }>("/healthz"),
    ).resolves.toEqual({ status: "ok" });
    expect(fetchStub).toHaveBeenCalledOnce();
    expect(String(fetchStub.mock.calls[0]?.[0])).toBe(
      "http://backend:8080/healthz",
    );
  });

  it("maps problem responses and response identifiers", async () => {
    const responseCorrelationId = "01991a4b-fa00-7000-8000-000000000004";
    const fetchStub = vi.fn<typeof fetch>(async () =>
      Response.json(
        { code: "dependency_unavailable", detail: "Try again later." },
        {
          status: 503,
          headers: {
            "Content-Type": "application/problem+json",
            "X-Correlation-ID": responseCorrelationId,
            "X-Request-ID": "01991a4b-fa00-7000-8000-000000000005",
          },
        },
      ),
    );
    const client = new BackendClient({
      baseUrl: "http://backend:8080",
      fetch: fetchStub,
      newCorrelationId: () => correlationId,
    });

    const failure = await client
      .request("/readyz")
      .catch((error: unknown) => error);
    expect(failure).toBeInstanceOf(BackendRequestError);
    expect(failure).toMatchObject({
      code: "dependency_unavailable",
      correlationId: responseCorrelationId,
      message: "Try again later.",
      requestId: "01991a4b-fa00-7000-8000-000000000005",
      status: 503,
    });
  });

  it("generates canonical UUIDv7 identifiers", () => {
    expect(createCorrelationId(1_797_504_000_000)).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/,
    );
  });
});
