import { EntityDetailPage } from './EntityDetailPage';
import { accountFields } from '../config/entityFields';
import type { Account } from '../types/entities';
import type { FieldDef } from '../components/EntityForm';

const detailFields = [
  { key: 'name', label: 'Name' },
  { key: 'email', label: 'Email' },
  { key: 'phone', label: 'Phone' },
  { key: 'website', label: 'Website' },
  { key: 'fax', label: 'Fax' },
  { key: 'category', label: 'Category' },
  { key: 'rating', label: 'Rating', render: (v: unknown) => v ? String(v) : '' },
  { key: 'access', label: 'Access' },
  { key: 'background_info', label: 'Background Info' },
  { key: 'created_at', label: 'Created', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
  { key: 'updated_at', label: 'Updated', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
];

export function AccountDetailPage() {
  return (
    <EntityDetailPage<Account>
      entityName="Account"
      entitySlug="accounts"
      endpoint="/accounts"
      fields={detailFields}
      formFields={accountFields as FieldDef[]}
      getTitle={(a) => a.name}
    />
  );
}
