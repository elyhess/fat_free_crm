import { Link } from 'react-router-dom';
import { useApi } from '../hooks/useApi';
import type { PaginatedResult } from '../types/entities';

interface RelatedColumn<T> {
  key: string;
  label: string;
  render?: (item: T) => React.ReactNode;
}

export interface RelatedEntitySection<T = Record<string, unknown>> {
  title: string;
  endpoint: string;
  columns: RelatedColumn<T>[];
  linkPath: (item: T) => string;
}

function RelatedSection<T extends { id: number }>({
  section,
}: {
  section: RelatedEntitySection<T>;
}) {
  const { data, loading } = useApi<PaginatedResult<T>>(section.endpoint);

  if (loading) {
    return (
      <div className="bg-white shadow rounded-lg p-6">
        <h2 className="text-lg font-medium text-gray-900 mb-4">{section.title}</h2>
        <p className="text-sm text-gray-500">Loading...</p>
      </div>
    );
  }

  const items = data?.data ?? [];

  return (
    <div className="bg-white shadow rounded-lg p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-medium text-gray-900">{section.title}</h2>
        {data && data.total > 0 && (
          <span className="text-sm text-gray-500">{data.total} total</span>
        )}
      </div>
      {items.length === 0 ? (
        <p className="text-sm text-gray-500">None found.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead>
              <tr className="border-b border-gray-200">
                {section.columns.map((col) => (
                  <th
                    key={col.key}
                    className="text-left py-2 pr-4 font-medium text-gray-500"
                  >
                    {col.label}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id} className="border-b border-gray-100 hover:bg-gray-50">
                  {section.columns.map((col, i) => {
                    const val = col.render
                      ? col.render(item)
                      : String((item as Record<string, unknown>)[col.key] ?? '');
                    return (
                      <td key={col.key} className="py-2 pr-4 text-gray-900">
                        {i === 0 ? (
                          <Link
                            to={section.linkPath(item)}
                            className="text-blue-600 hover:text-blue-800 hover:underline"
                          >
                            {val}
                          </Link>
                        ) : (
                          val
                        )}
                      </td>
                    );
                  })}
                </tr>
              ))}
            </tbody>
          </table>
          {data && data.total > data.per_page && (
            <p className="text-xs text-gray-500 mt-2">
              Showing {items.length} of {data.total}
            </p>
          )}
        </div>
      )}
    </div>
  );
}

export function RelatedEntities({
  sections,
}: {
  sections: RelatedEntitySection[];
}) {
  return (
    <>
      {sections.map((section) => (
        <RelatedSection key={section.title} section={section as RelatedEntitySection<{ id: number }>} />
      ))}
    </>
  );
}
