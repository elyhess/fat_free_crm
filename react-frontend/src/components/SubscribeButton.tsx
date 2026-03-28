import { useState, useEffect, useCallback } from 'react';
import { api } from '../api/client';

interface SubscriptionState {
  subscribed: boolean;
  subscribed_users: number[];
}

interface SubscribeButtonProps {
  entitySlug: string;
  entityId: number;
}

export function SubscribeButton({ entitySlug, entityId }: SubscribeButtonProps) {
  const [state, setState] = useState<SubscriptionState | null>(null);
  const [loading, setLoading] = useState(false);

  const fetchState = useCallback(() => {
    api.get<SubscriptionState>(`/${entitySlug}/${entityId}/subscription`).then(setState).catch(() => {});
  }, [entitySlug, entityId]);

  useEffect(() => {
    fetchState();
  }, [fetchState]);

  async function toggle() {
    if (!state) return;
    setLoading(true);
    try {
      const action = state.subscribed ? 'unsubscribe' : 'subscribe';
      const result = await api.post<SubscriptionState>(`/${entitySlug}/${entityId}/${action}`, {});
      setState(result);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }

  if (!state) return null;

  return (
    <button
      onClick={toggle}
      disabled={loading}
      className={`w-full text-left px-3 py-2 text-sm rounded-md border transition-colors ${
        state.subscribed
          ? 'border-blue-200 bg-blue-50 text-blue-700 hover:bg-blue-100'
          : 'border-gray-200 bg-white text-gray-600 hover:bg-gray-50'
      } disabled:opacity-50`}
    >
      {state.subscribed ? 'Subscribed to notifications' : 'Subscribe to notifications'}
    </button>
  );
}
