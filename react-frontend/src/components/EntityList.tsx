import { useState } from 'react';
import { useApi } from '../hooks/useApi';
import type { PaginatedResult } from '../types/entities';

interface Column<T> {
  key: string;
  label: string;
  render?: (item: T) => React.ReactNode;
}

interface EntityListProps<T> {
  title: string;
  endpoint: string;
  columns: Column<T>[];
  getRowKey: (item: T) => string | number;
}

export function EntityList<T>({
  title,
  endpoint,
  columns,
  getRowKey,
}: EntityListProps<T>) {
  const [page, setPage] = useState(1);
  const [perPage] = useState(20);
  const [sort, setSort] = useState('id');
  const [order, setOrder] = useState('desc');

  const path = `${endpoint}?page=${page}&per_page=${perPage}&sort=${sort}&order=${order}`;
  const { data, loading, error } = useApi<PaginatedResult<T>>(path);

  function handleSort(key: string) {
    if (sort === key) {
      setOrder(order === 'asc' ? 'desc' : 'asc');
    } else {
      setSort(key);
      setOrder('asc');
    }
    setPage(1);
  }

  return (
    <div>
      <h1 className="text-2xl font-semibold text-gray-900 mb-6">{title}</h1>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      {loading && !data ? (
        <div className="text-gray-500">Loading...</div>
      ) : (
        <>
          <div className="bg-white shadow rounded-lg overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  {columns.map((col) => (
                    <th
                      key={col.key}
                      onClick={() => handleSort(col.key)}
                      className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:text-gray-700"
                    >
                      {col.label}
                      {sort === col.key && (
                        <span className="ml-1">{order === 'asc' ? '\u2191' : '\u2193'}</span>
                      )}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {data?.data.map((item) => (
                  <tr key={getRowKey(item)} className="hover:bg-gray-50">
                    {columns.map((col) => (
                      <td key={col.key} className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {col.render
                          ? col.render(item)
                          : String((item as Record<string, unknown>)[col.key] ?? '')}
                      </td>
                    ))}
                  </tr>
                ))}
                {data?.data.length === 0 && (
                  <tr>
                    <td
                      colSpan={columns.length}
                      className="px-6 py-8 text-center text-sm text-gray-500"
                    >
                      No records found.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          {data && data.total_pages > 1 && (
            <div className="flex items-center justify-between mt-4">
              <span className="text-sm text-gray-700">
                {data.total} total &middot; Page {data.page} of {data.total_pages}
              </span>
              <div className="space-x-2">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page <= 1}
                  className="px-3 py-1 text-sm border rounded disabled:opacity-50"
                >
                  Previous
                </button>
                <button
                  onClick={() => setPage((p) => Math.min(data.total_pages, p + 1))}
                  disabled={page >= data.total_pages}
                  className="px-3 py-1 text-sm border rounded disabled:opacity-50"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
