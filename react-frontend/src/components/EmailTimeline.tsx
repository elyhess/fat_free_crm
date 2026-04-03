import { useApi } from '../hooks/useApi';
import { useMutation } from '../hooks/useMutation';

interface Email {
  id: number;
  sent_from: string;
  sent_to: string;
  cc?: string;
  subject: string;
  body: string;
  state: string;
  sent_at?: string;
  created_at?: string;
}

interface EmailTimelineProps {
  entitySlug: string;
  entityId: number;
}

export function EmailTimeline({ entitySlug, entityId }: EmailTimelineProps) {
  const { data: emails, loading, refetch } = useApi<Email[]>(`/${entitySlug}/${entityId}/emails`);
  const deleteMutation = useMutation();

  async function handleDelete(emailId: number) {
    try {
      await deleteMutation.del(`/emails/${emailId}`);
      refetch();
    } catch { /* error in mutation */ }
  }

  if (loading) return <div className="text-sm text-gray-500">Loading emails...</div>;
  if (!emails || emails.length === 0) return null;

  return (
    <div className="bg-white shadow rounded-lg p-6">
      <h2 className="text-lg font-medium text-gray-900 mb-4">Emails ({emails.length})</h2>
      <div className="space-y-4">
        {emails.map((email) => (
          <div key={email.id} className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-start justify-between">
              <div className="flex-1 min-w-0">
                <h3 className="text-sm font-medium text-gray-900 truncate">{email.subject || '(no subject)'}</h3>
                <p className="text-xs text-gray-500 mt-1">
                  From: {email.sent_from} &rarr; To: {email.sent_to}
                  {email.cc && <span> (CC: {email.cc})</span>}
                </p>
                {email.sent_at && (
                  <p className="text-xs text-gray-400 mt-0.5">
                    {new Date(email.sent_at).toLocaleString()}
                  </p>
                )}
              </div>
              <button
                onClick={() => handleDelete(email.id)}
                disabled={deleteMutation.loading}
                className="ml-2 text-xs text-red-500 hover:text-red-700"
                title="Delete email"
              >
                &times;
              </button>
            </div>
            <div className="mt-2 text-sm text-gray-700 whitespace-pre-wrap max-h-40 overflow-y-auto">
              {email.body?.slice(0, 500)}
              {email.body && email.body.length > 500 && '...'}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
