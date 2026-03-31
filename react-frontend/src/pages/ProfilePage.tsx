import { useState, useEffect, useRef } from 'react';
import { useApi } from '../hooks/useApi';
import { useMutation } from '../hooks/useMutation';

const API_BASE = '/api/v1';

interface Profile {
  id: number;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  title?: string;
  company?: string;
  alt_email?: string;
  phone?: string;
  mobile?: string;
  admin: boolean;
  avatar_url?: string;
}

export function ProfilePage() {
  const { data: profile, loading, error, refetch } = useApi<Profile>('/profile');
  const profileMutation = useMutation();
  const passwordMutation = useMutation();

  const [form, setForm] = useState({
    email: '',
    first_name: '',
    last_name: '',
    title: '',
    company: '',
    alt_email: '',
    phone: '',
    mobile: '',
  });

  const [passwordForm, setPasswordForm] = useState({
    current_password: '',
    new_password: '',
    confirm_password: '',
  });

  const [profileSaved, setProfileSaved] = useState(false);
  const [passwordChanged, setPasswordChanged] = useState(false);
  const [confirmError, setConfirmError] = useState('');
  const [avatarKey, setAvatarKey] = useState(0);
  const [avatarUploading, setAvatarUploading] = useState(false);
  const [avatarError, setAvatarError] = useState('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (profile) {
      setForm({
        email: profile.email || '',
        first_name: profile.first_name || '',
        last_name: profile.last_name || '',
        title: profile.title || '',
        company: profile.company || '',
        alt_email: profile.alt_email || '',
        phone: profile.phone || '',
        mobile: profile.mobile || '',
      });
    }
  }, [profile]);

  async function handleProfileSubmit(e: React.FormEvent) {
    e.preventDefault();
    setProfileSaved(false);
    try {
      await profileMutation.put('/profile', form);
      setProfileSaved(true);
      refetch();
    } catch { /* error in mutation */ }
  }

  async function handlePasswordSubmit(e: React.FormEvent) {
    e.preventDefault();
    setPasswordChanged(false);
    setConfirmError('');

    if (passwordForm.new_password !== passwordForm.confirm_password) {
      setConfirmError('New passwords do not match');
      return;
    }

    try {
      await passwordMutation.put('/profile/password', {
        current_password: passwordForm.current_password,
        new_password: passwordForm.new_password,
      });
      setPasswordChanged(true);
      setPasswordForm({ current_password: '', new_password: '', confirm_password: '' });
    } catch { /* error in mutation */ }
  }

  async function handleAvatarUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setAvatarError('');
    setAvatarUploading(true);
    try {
      const formData = new FormData();
      formData.append('avatar', file);
      const token = localStorage.getItem('token');
      const res = await fetch(`${API_BASE}/profile/avatar`, {
        method: 'POST',
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        body: formData,
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({ error: 'Upload failed' }));
        setAvatarError(body.error || 'Upload failed');
      } else {
        setAvatarKey((k) => k + 1);
      }
    } catch {
      setAvatarError('Upload failed');
    } finally {
      setAvatarUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  }

  async function handleAvatarDelete() {
    setAvatarError('');
    try {
      const token = localStorage.getItem('token');
      await fetch(`${API_BASE}/profile/avatar`, {
        method: 'DELETE',
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      });
      setAvatarKey((k) => k + 1);
    } catch {
      setAvatarError('Failed to remove avatar');
    }
  }

  if (loading) return <div className="text-gray-500">Loading...</div>;
  if (error) return <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">{error}</div>;
  if (!profile) return null;

  return (
    <div className="max-w-2xl">
      <h1 className="text-2xl font-semibold text-gray-900 mb-6">Profile</h1>

      {/* Avatar */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h2 className="text-lg font-medium text-gray-900 mb-4">Avatar</h2>
        <div className="flex items-center gap-6">
          <img
            key={avatarKey}
            src={`${API_BASE}/avatars/${profile.id}?v=${avatarKey}`}
            alt="Avatar"
            className="w-20 h-20 rounded-full object-cover bg-gray-200"
            onError={(e) => { (e.target as HTMLImageElement).src = ''; }}
          />
          <div className="space-y-2">
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                disabled={avatarUploading}
                className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
              >
                {avatarUploading ? 'Uploading...' : 'Upload Photo'}
              </button>
              <button
                type="button"
                onClick={handleAvatarDelete}
                className="px-3 py-1.5 text-sm border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
              >
                Remove
              </button>
            </div>
            <p className="text-xs text-gray-500">PNG, JPEG, or GIF. Max 5MB.</p>
            {avatarError && (
              <p className="text-xs text-red-600">{avatarError}</p>
            )}
            <input
              ref={fileInputRef}
              type="file"
              accept="image/png,image/jpeg,image/gif"
              onChange={handleAvatarUpload}
              className="hidden"
            />
          </div>
        </div>
      </div>

      {/* Profile Info */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h2 className="text-lg font-medium text-gray-900 mb-4">Personal Information</h2>
        <p className="text-sm text-gray-500 mb-4">Username: <span className="font-medium text-gray-900">{profile.username}</span></p>

        {profileSaved && (
          <div className="bg-green-50 border border-green-200 text-green-700 px-4 py-2 rounded mb-4 text-sm">
            Profile updated successfully.
          </div>
        )}
        {profileMutation.error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-2 rounded mb-4 text-sm">
            {profileMutation.error}
          </div>
        )}

        <form onSubmit={handleProfileSubmit} className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">First Name</label>
              <input
                type="text"
                value={form.first_name}
                onChange={(e) => setForm({ ...form, first_name: e.target.value })}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Last Name</label>
              <input
                type="text"
                value={form.last_name}
                onChange={(e) => setForm({ ...form, last_name: e.target.value })}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
            <input
              type="email"
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Title</label>
              <input
                type="text"
                value={form.title}
                onChange={(e) => setForm({ ...form, title: e.target.value })}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Company</label>
              <input
                type="text"
                value={form.company}
                onChange={(e) => setForm({ ...form, company: e.target.value })}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Phone</label>
              <input
                type="tel"
                value={form.phone}
                onChange={(e) => setForm({ ...form, phone: e.target.value })}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Mobile</label>
              <input
                type="tel"
                value={form.mobile}
                onChange={(e) => setForm({ ...form, mobile: e.target.value })}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Alt Email</label>
            <input
              type="email"
              value={form.alt_email}
              onChange={(e) => setForm({ ...form, alt_email: e.target.value })}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div className="flex justify-end">
            <button
              type="submit"
              disabled={profileMutation.loading}
              className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
            >
              {profileMutation.loading ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>

      {/* Change Password */}
      <div className="bg-white shadow rounded-lg p-6">
        <h2 className="text-lg font-medium text-gray-900 mb-4">Change Password</h2>

        {passwordChanged && (
          <div className="bg-green-50 border border-green-200 text-green-700 px-4 py-2 rounded mb-4 text-sm">
            Password changed successfully.
          </div>
        )}
        {(passwordMutation.error || confirmError) && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-2 rounded mb-4 text-sm">
            {confirmError || passwordMutation.error}
          </div>
        )}

        <form onSubmit={handlePasswordSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Current Password</label>
            <input
              type="password"
              value={passwordForm.current_password}
              onChange={(e) => setPasswordForm({ ...passwordForm, current_password: e.target.value })}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">New Password</label>
            <input
              type="password"
              value={passwordForm.new_password}
              onChange={(e) => setPasswordForm({ ...passwordForm, new_password: e.target.value })}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <p className="text-xs text-gray-500 mt-1">At least 14 characters with uppercase, lowercase, digit, and symbol.</p>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Confirm New Password</label>
            <input
              type="password"
              value={passwordForm.confirm_password}
              onChange={(e) => setPasswordForm({ ...passwordForm, confirm_password: e.target.value })}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div className="flex justify-end">
            <button
              type="submit"
              disabled={passwordMutation.loading}
              className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
            >
              {passwordMutation.loading ? 'Changing...' : 'Change Password'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
