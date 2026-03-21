import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useApi } from '../hooks/useApi';
import { useMutation } from '../hooks/useMutation';
import { Modal } from '../components/Modal';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { EntityForm } from '../components/EntityForm';
import type { FieldDef } from '../components/EntityForm';

interface Comment {
  id: number;
  user_id: number;
  comment: string;
  created_at: string;
}

interface Tag {
  id: number;
  name: string;
}

interface Address {
  id: number;
  street1?: string;
  city?: string;
  state?: string;
  country?: string;
  address_type?: string;
}

interface Version {
  id: number;
  event: string;
  whodunnit?: string;
  created_at: string;
}

interface DetailField {
  key: string;
  label: string;
  render?: (value: unknown) => string;
}

interface EntityDetailPageProps<T> {
  entityName: string;
  entitySlug: string;
  endpoint: string;
  fields: DetailField[];
  formFields: FieldDef[];
  getTitle: (item: T) => string;
}

export function EntityDetailPage<T extends { id: number }>({
  entityName,
  entitySlug,
  endpoint,
  fields,
  formFields,
  getTitle,
}: EntityDetailPageProps<T>) {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data, loading, error, refetch } = useApi<T>(`${endpoint}/${id}`);
  const { data: comments, refetch: refetchComments } = useApi<Comment[]>(`/${entitySlug}/${id}/comments`);
  const { data: tags, refetch: refetchTags } = useApi<Tag[]>(`/${entitySlug}/${id}/tags`);
  const { data: addresses } = useApi<Address[]>(`/${entitySlug}/${id}/addresses`);
  const { data: versions } = useApi<Version[]>(`/${entitySlug}/${id}/versions`);

  const [showEdit, setShowEdit] = useState(false);
  const [showDelete, setShowDelete] = useState(false);
  const [newComment, setNewComment] = useState('');
  const [newTag, setNewTag] = useState('');

  const editMutation = useMutation();
  const deleteMutation = useMutation();
  const commentMutation = useMutation();
  const tagMutation = useMutation();

  async function handleUpdate(values: Record<string, unknown>) {
    try {
      await editMutation.put(`${endpoint}/${id}`, values);
      setShowEdit(false);
      refetch();
    } catch { /* error in mutation.error */ }
  }

  async function handleDelete() {
    try {
      await deleteMutation.del(`${endpoint}/${id}`);
      navigate(`/${entitySlug}`);
    } catch { /* error in mutation.error */ }
  }

  async function handleAddComment(e: React.FormEvent) {
    e.preventDefault();
    if (!newComment.trim()) return;
    try {
      await commentMutation.post(`/${entitySlug}/${id}/comments`, { comment: newComment });
      setNewComment('');
      refetchComments();
    } catch { /* */ }
  }

  async function handleAddTag(e: React.FormEvent) {
    e.preventDefault();
    if (!newTag.trim()) return;
    try {
      await tagMutation.post(`/${entitySlug}/${id}/tags`, { name: newTag });
      setNewTag('');
      refetchTags();
    } catch { /* */ }
  }

  async function handleRemoveTag(tagId: number) {
    try {
      await tagMutation.del(`/${entitySlug}/${id}/tags/${tagId}`);
      refetchTags();
    } catch { /* */ }
  }

  if (loading) return <div className="text-gray-500">Loading...</div>;
  if (error) return <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">{error}</div>;
  if (!data) return null;

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <button onClick={() => navigate(`/${entitySlug}`)} className="text-sm text-blue-600 hover:text-blue-800 mb-1">
            &larr; Back to {entityName}s
          </button>
          <h1 className="text-2xl font-semibold text-gray-900">{getTitle(data)}</h1>
        </div>
        <div className="flex gap-2">
          <button onClick={() => { editMutation.reset(); setShowEdit(true); }} className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700">
            Edit
          </button>
          <button onClick={() => { deleteMutation.reset(); setShowDelete(true); }} className="px-4 py-2 text-sm bg-red-600 text-white rounded-md hover:bg-red-700">
            Delete
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Info */}
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-white shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Details</h2>
            <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-3">
              {fields.map((f) => {
                const val = (data as Record<string, unknown>)[f.key];
                const display = f.render ? f.render(val) : String(val ?? '');
                if (!display) return null;
                return (
                  <div key={f.key}>
                    <dt className="text-sm font-medium text-gray-500">{f.label}</dt>
                    <dd className="text-sm text-gray-900">{display}</dd>
                  </div>
                );
              })}
            </dl>
          </div>

          {/* Comments */}
          <div className="bg-white shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Comments</h2>
            {comments && comments.length > 0 ? (
              <div className="space-y-3 mb-4">
                {comments.map((c) => (
                  <div key={c.id} className="border-l-2 border-gray-200 pl-3">
                    <p className="text-sm text-gray-900">{c.comment}</p>
                    <p className="text-xs text-gray-500 mt-1">{new Date(c.created_at).toLocaleString()}</p>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-gray-500 mb-4">No comments yet.</p>
            )}
            <form onSubmit={handleAddComment} className="flex gap-2">
              <input
                type="text"
                value={newComment}
                onChange={(e) => setNewComment(e.target.value)}
                placeholder="Add a comment..."
                className="flex-1 border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <button type="submit" disabled={commentMutation.loading} className="px-4 py-2 text-sm bg-gray-800 text-white rounded-md hover:bg-gray-900 disabled:opacity-50">
                Add
              </button>
            </form>
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Tags */}
          <div className="bg-white shadow rounded-lg p-6">
            <h2 className="text-sm font-medium text-gray-900 mb-3">Tags</h2>
            <div className="flex flex-wrap gap-2 mb-3">
              {tags && tags.length > 0 ? (
                tags.map((tag) => (
                  <span key={tag.id} className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                    {tag.name}
                    <button onClick={() => handleRemoveTag(tag.id)} className="text-blue-600 hover:text-blue-900">&times;</button>
                  </span>
                ))
              ) : (
                <span className="text-xs text-gray-500">No tags</span>
              )}
            </div>
            <form onSubmit={handleAddTag} className="flex gap-1">
              <input
                type="text"
                value={newTag}
                onChange={(e) => setNewTag(e.target.value)}
                placeholder="Add tag..."
                className="flex-1 border border-gray-300 rounded px-2 py-1 text-xs focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
              <button type="submit" className="px-2 py-1 text-xs bg-gray-800 text-white rounded hover:bg-gray-900">+</button>
            </form>
          </div>

          {/* Addresses */}
          {addresses && addresses.length > 0 && (
            <div className="bg-white shadow rounded-lg p-6">
              <h2 className="text-sm font-medium text-gray-900 mb-3">Addresses</h2>
              {addresses.map((a) => (
                <div key={a.id} className="text-sm text-gray-700 mb-2">
                  {a.address_type && <span className="text-xs text-gray-500 uppercase">{a.address_type}</span>}
                  <p>{[a.street1, a.city, a.state, a.country].filter(Boolean).join(', ')}</p>
                </div>
              ))}
            </div>
          )}

          {/* History */}
          {versions && versions.length > 0 && (
            <div className="bg-white shadow rounded-lg p-6">
              <h2 className="text-sm font-medium text-gray-900 mb-3">History</h2>
              <div className="space-y-2">
                {versions.slice(0, 10).map((v) => (
                  <div key={v.id} className="text-xs text-gray-600">
                    <span className="font-medium capitalize">{v.event}</span>
                    {v.whodunnit && <span> by user {v.whodunnit}</span>}
                    <span className="text-gray-400 ml-1">{new Date(v.created_at).toLocaleString()}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Edit Modal */}
      <Modal open={showEdit} onClose={() => setShowEdit(false)} title={`Edit ${entityName}`}>
        <EntityForm
          fields={formFields}
          initialValues={data as unknown as Record<string, unknown>}
          onSubmit={handleUpdate}
          onCancel={() => setShowEdit(false)}
          loading={editMutation.loading}
          error={editMutation.error}
          submitLabel="Update"
        />
      </Modal>

      {/* Delete Confirmation */}
      <ConfirmDialog
        open={showDelete}
        onClose={() => setShowDelete(false)}
        onConfirm={handleDelete}
        title={`Delete ${entityName}`}
        message={`Are you sure you want to delete "${getTitle(data)}"? This action cannot be undone.`}
        loading={deleteMutation.loading}
      />
    </div>
  );
}
