import { useEffect, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';

export function ConfirmEmailPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token') || '';
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [message, setMessage] = useState('');

  useEffect(() => {
    if (!token) {
      setStatus('error');
      setMessage('Missing confirmation token.');
      return;
    }

    fetch('/api/v1/auth/confirm', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ token }),
    })
      .then(async (res) => {
        const body = await res.json();
        if (!res.ok) throw new Error(body.error || 'Confirmation failed');
        setStatus('success');
        setMessage(body.status || 'Email confirmed');
      })
      .catch((err) => {
        setStatus('error');
        setMessage(err instanceof Error ? err.message : 'Confirmation failed');
      });
  }, [token]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="w-full max-w-sm">
        <h1 className="text-2xl font-semibold text-center text-gray-900 mb-8">
          Fat Free CRM
        </h1>
        <div className="bg-white shadow rounded-lg px-8 py-8 space-y-4">
          {status === 'loading' && (
            <p className="text-center text-gray-600">Confirming your email...</p>
          )}
          {status === 'success' && (
            <>
              <div className="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded text-sm">
                {message}
              </div>
              <Link
                to="/login"
                className="block text-center text-sm text-blue-600 hover:text-blue-500"
              >
                Sign in to your account
              </Link>
            </>
          )}
          {status === 'error' && (
            <>
              <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded text-sm">
                {message}
              </div>
              <Link
                to="/login"
                className="block text-center text-sm text-blue-600 hover:text-blue-500"
              >
                Back to sign in
              </Link>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
