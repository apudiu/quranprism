import axios, {
  type AxiosError,
  type AxiosInstance,
  type InternalAxiosRequestConfig,
} from "axios";

/**
 * Wiring points each app supplies so the shared client stays auth-strategy
 * agnostic. The admin app plugs in its in-memory access token + single-flight
 * refresh; the web app can plug in its own session model later.
 */
export interface HttpClientOptions {
  baseURL: string;
  /** Returns the current access token to attach as a Bearer header, or null. */
  getAccessToken?: () => string | null;
  /**
   * Called once on a 401. Should attempt a token refresh and resolve with a
   * fresh access token, or null to give up. Callers are expected to coalesce
   * concurrent refreshes themselves (single in-flight promise).
   */
  onUnauthorized?: () => Promise<string | null>;
  /** Called on a 403 — e.g. redirect to a "forbidden" page. */
  onForbidden?: () => void;
  timeoutMs?: number;
  withCredentials?: boolean;
}

type RetryableConfig = InternalAxiosRequestConfig & { _retry?: boolean };

/**
 * Builds an axios instance mirroring the reference admin's interceptor seam:
 * attach Bearer token on request, refresh-and-retry once on 401, hook 403.
 */
export function createHttpClient(opts: HttpClientOptions): AxiosInstance {
  const instance = axios.create({
    baseURL: opts.baseURL,
    timeout: opts.timeoutMs ?? 10_000,
    withCredentials: opts.withCredentials ?? true,
  });

  instance.interceptors.request.use((config) => {
    const token = opts.getAccessToken?.();
    if (token) config.headers.set("Authorization", `Bearer ${token}`);
    return config;
  });

  instance.interceptors.response.use(
    (res) => res,
    async (error: AxiosError) => {
      const status = error.response?.status;
      const original = error.config as RetryableConfig | undefined;

      // Refresh-and-retry once. `_retry` guards against an infinite loop when
      // the retried request also 401s.
      if (status === 401 && original && !original._retry && opts.onUnauthorized) {
        original._retry = true;
        const newToken = await opts.onUnauthorized();
        if (newToken) {
          original.headers.set("Authorization", `Bearer ${newToken}`);
          return instance(original);
        }
      }

      if (status === 403) opts.onForbidden?.();

      return Promise.reject(error);
    },
  );

  return instance;
}
