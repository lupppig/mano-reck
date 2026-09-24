import { loadServerEnvironment } from "./environment";

const uuidV7Pattern =
  /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

type Fetch = typeof fetch;

export type BackendRequest = Omit<RequestInit, "headers"> & {
  correlationId?: string;
  headers?: HeadersInit;
};

type Problem = {
  code?: unknown;
  detail?: unknown;
  title?: unknown;
};

export class BackendRequestError extends Error {
  readonly code: string;
  readonly correlationId: string;
  readonly requestId: string | null;
  readonly status: number | null;

  constructor(options: {
    cause?: unknown;
    code: string;
    correlationId: string;
    message: string;
    requestId?: string | null;
    status?: number | null;
  }) {
    super(options.message, { cause: options.cause });
    this.name = "BackendRequestError";
    this.code = options.code;
    this.correlationId = options.correlationId;
    this.requestId = options.requestId ?? null;
    this.status = options.status ?? null;
  }
}

export class BackendClient {
  readonly #baseUrl: URL;
  readonly #fetch: Fetch;
  readonly #newCorrelationId: () => string;

  constructor(options?: {
    baseUrl?: string;
    fetch?: Fetch;
    newCorrelationId?: () => string;
  }) {
    const configuredBaseUrl =
      options?.baseUrl ?? loadServerEnvironment().backendInternalUrl;
    this.#baseUrl = new URL(`${configuredBaseUrl}/`);
    this.#fetch = options?.fetch ?? fetch;
    this.#newCorrelationId = options?.newCorrelationId ?? createCorrelationId;
  }

  async request<T>(path: string, request: BackendRequest = {}): Promise<T> {
    if (!path.startsWith("/") || path.startsWith("//")) {
      throw new Error("backend request path must start with one slash");
    }

    const correlationId = request.correlationId ?? this.#newCorrelationId();
    if (!uuidV7Pattern.test(correlationId)) {
      throw new Error("backend correlation ID must be a canonical UUIDv7");
    }

    const headers = new Headers(request.headers);
    headers.set("Accept", "application/json");
    headers.set("X-Correlation-ID", correlationId);
    const requestInit = { ...request };
    delete requestInit.correlationId;
    delete requestInit.headers;

    let response: Response;
    try {
      response = await this.#fetch(new URL(path.slice(1), this.#baseUrl), {
        ...requestInit,
        headers,
      });
    } catch (cause) {
      throw new BackendRequestError({
        cause,
        code: "backend_unavailable",
        correlationId,
        message: "The platform service is temporarily unavailable.",
      });
    }

    if (!response.ok) {
      throw await responseError(response, correlationId);
    }
    if (response.status === 204) {
      return undefined as T;
    }
    return (await response.json()) as T;
  }
}

export function createCorrelationId(now = Date.now()): string {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  let timestamp = BigInt(now);
  for (let index = 5; index >= 0; index -= 1) {
    bytes[index] = Number(timestamp & 0xffn);
    timestamp >>= 8n;
  }
  bytes[6] = (bytes[6] & 0x0f) | 0x70;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;

  const hex = Array.from(bytes, (value) =>
    value.toString(16).padStart(2, "0"),
  ).join("");
  return [
    hex.slice(0, 8),
    hex.slice(8, 12),
    hex.slice(12, 16),
    hex.slice(16, 20),
    hex.slice(20),
  ].join("-");
}

async function responseError(
  response: Response,
  sentCorrelationId: string,
): Promise<BackendRequestError> {
  const requestId = response.headers.get("X-Request-ID");
  const correlationId =
    response.headers.get("X-Correlation-ID") ?? sentCorrelationId;
  const contentType = response.headers.get("Content-Type") ?? "";
  let problem: Problem = {};
  if (contentType.includes("json")) {
    try {
      problem = (await response.json()) as Problem;
    } catch {
      problem = {};
    }
  }

  return new BackendRequestError({
    code:
      typeof problem.code === "string"
        ? problem.code
        : "backend_request_failed",
    correlationId,
    message:
      typeof problem.detail === "string"
        ? problem.detail
        : typeof problem.title === "string"
          ? problem.title
          : "The platform service could not complete the request.",
    requestId,
    status: response.status,
  });
}
