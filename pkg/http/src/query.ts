import { QueryClient } from "@tanstack/solid-query";

/**
 * Shared TanStack Query defaults. 5-minute stale time + no refetch-on-focus
 * matches the reference admin: a moderation surface shouldn't re-run every
 * query on alt-tab. Mutations invalidate explicitly.
 */
export function createQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 5 * 60 * 1000,
        refetchOnWindowFocus: false,
        retry: 1,
      },
    },
  });
}
