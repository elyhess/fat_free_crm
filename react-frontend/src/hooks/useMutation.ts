import { useState, useCallback } from 'react';
import { api } from '../api/client';

interface UseMutationState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
}

interface UseMutationReturn<T> extends UseMutationState<T> {
  post: (path: string, body: unknown) => Promise<T>;
  put: (path: string, body: unknown) => Promise<T>;
  del: (path: string) => Promise<T>;
  reset: () => void;
}

export function useMutation<T = unknown>(): UseMutationReturn<T> {
  const [state, setState] = useState<UseMutationState<T>>({
    data: null,
    loading: false,
    error: null,
  });

  const execute = useCallback(async (fn: () => Promise<T>): Promise<T> => {
    setState({ data: null, loading: true, error: null });
    try {
      const result = await fn();
      setState({ data: result, loading: false, error: null });
      return result;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'An error occurred';
      // Try to parse JSON error from API
      let errorText = message;
      try {
        const parsed = JSON.parse(message);
        errorText = parsed.error || parsed.message || message;
      } catch {
        // not JSON, use as-is
      }
      setState({ data: null, loading: false, error: errorText });
      throw err;
    }
  }, []);

  return {
    ...state,
    post: useCallback((path: string, body: unknown) => execute(() => api.post<T>(path, body)), [execute]),
    put: useCallback((path: string, body: unknown) => execute(() => api.put<T>(path, body)), [execute]),
    del: useCallback((path: string) => execute(() => api.delete<T>(path)), [execute]),
    reset: useCallback(() => setState({ data: null, loading: false, error: null }), []),
  };
}
