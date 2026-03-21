import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useApi } from '../hooks/useApi';
import { useMutation } from '../hooks/useMutation';
import { Modal } from './Modal';
import { ConfirmDialog } from './ConfirmDialog';
import { EntityForm } from './EntityForm';
import type { FieldDef } from './EntityForm';
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
  formFields?: FieldDef[];
  detailPath?: (item: T) => string;
}

export function EntityList<T extends { id: number }>({
  title,
  endpoint,
  columns,
  getRowKey,
  formFields,
  detailPath,
}: EntityListProps<T>) {
  const [page, setPage] = useState(1);
  const [perPage] = useState(20);
  const [sort, setSort] = useState('id');
  const [order, setOrder] = useState('desc');

  // Form state
  const [showForm, setShowForm] = useState(false);
  const [editItem, setEditItem] = useState<T | null>(null);
  const [deleteItem, setDeleteItem] = useState<T | null>(null);

  const path = `${endpoint}?page=${page}&per_page=${perPage}&sort=${sort}&order=${order}`;
  const { data, loading, error, refetch } = useApi<PaginatedResult<T>>(path);
  const mutation = useMutation();
  const deleteMutation = useMutation();

  function handleSort(key: string) {
    if (sort === key) {
      setOrder(order === 'asc' ? 'desc' : 'asc');
    } else {
      setSort(key);
      setOrder('asc');
    }
    setPage(1);
  }

  function openCreate() {
    setEditItem(null);
    mutation.reset();
    setShowForm(true);
  }

  function openEdit(item: T) {
    setEditItem(item);
    mutation.reset();
    setShowForm(true);
  }

  function closeForm() {
    setShowForm(false);
    setEditItem(null);
  }

  async function handleSubmit(values: Record<string, unknown>) {
    try {
      if (editItem) {
        await mutation.put(`${endpoint}/${editItem.id}`, values);
      } else {
        await mutation.post(endpoint, values);
      }
      closeForm();
      refetch();
    } catch {
      // error is captured in mutation.error
    }
  }

  async function handleDelete() {
    if (!deleteItem) return;
    try {
      await deleteMutation.del(`${endpoint}/${deleteItem.id}`);
      setDeleteItem(null);
      refetch();
    } catch {
      // error is captured in deleteMutation.error
    }
  }

  const hasActions = !!formFields;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold text-gray-900">{title}</h1>
        {formFields && (
          <button
            onClick={openCreate}
            className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            + New
          </button>
        )}
      </div>

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
                  {hasActions && (
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Actions
                    </th>
                  )}
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {data?.data.map((item) => (
                  <tr key={getRowKey(item)} className="hover:bg-gray-50">
                    {columns.map((col, colIdx) => {
                      const content = col.render
                        ? col.render(item)
                        : String((item as Record<string, unknown>)[col.key] ?? '');
                      return (
                        <td key={col.key} className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {colIdx === 0 && detailPath ? (
                            <Link to={detailPath(item)} className="text-blue-600 hover:text-blue-800 hover:underline">
                              {content}
                            </Link>
                          ) : (
                            content
                          )}
                        </td>
                      );
                    })}
                    {hasActions && (
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm">
                        <button
                          onClick={() => openEdit(item)}
                          className="text-blue-600 hover:text-blue-800 mr-3"
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => { deleteMutation.reset(); setDeleteItem(item); }}
                          className="text-red-600 hover:text-red-800"
                        >
                          Delete
                        </button>
                      </td>
                    )}
                  </tr>
                ))}
                {data?.data.length === 0 && (
                  <tr>
                    <td
                      colSpan={columns.length + (hasActions ? 1 : 0)}
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

      {/* Create/Edit Modal */}
      {formFields && (
        <Modal
          open={showForm}
          onClose={closeForm}
          title={editItem ? `Edit ${title.replace(/s$/, '')}` : `New ${title.replace(/s$/, '')}`}
        >
          <EntityForm
            fields={formFields}
            initialValues={editItem ? (editItem as unknown as Record<string, unknown>) : {}}
            onSubmit={handleSubmit}
            onCancel={closeForm}
            loading={mutation.loading}
            error={mutation.error}
            submitLabel={editItem ? 'Update' : 'Create'}
          />
        </Modal>
      )}

      {/* Delete Confirmation */}
      <ConfirmDialog
        open={!!deleteItem}
        onClose={() => setDeleteItem(null)}
        onConfirm={handleDelete}
        title={`Delete ${title.replace(/s$/, '')}`}
        message="Are you sure you want to delete this record? This action cannot be undone."
        loading={deleteMutation.loading}
      />
    </div>
  );
}
