import { EntityList } from '../components/EntityList';
import type { Account } from '../types/entities';

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'email', label: 'Email' },
  { key: 'phone', label: 'Phone' },
  { key: 'category', label: 'Category' },
  {
    key: 'access',
    label: 'Access',
    render: (a: Account) => (
      <span
        className={`inline-flex px-2 py-0.5 text-xs font-medium rounded ${
          a.access === 'Public'
            ? 'bg-green-100 text-green-700'
            : a.access === 'Private'
              ? 'bg-red-100 text-red-700'
              : 'bg-yellow-100 text-yellow-700'
        }`}
      >
        {a.access}
      </span>
    ),
  },
];

export function AccountsPage() {
  return (
    <EntityList<Account>
      title="Accounts"
      endpoint="/accounts"
      columns={columns}
      getRowKey={(a) => a.id}
    />
  );
}
