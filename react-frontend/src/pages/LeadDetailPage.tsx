import { EntityDetailPage } from './EntityDetailPage';
import { leadFields } from '../config/entityFields';
import type { Lead } from '../types/entities';
import type { FieldDef } from '../components/EntityForm';

const detailFields = [
  { key: 'first_name', label: 'First Name' },
  { key: 'last_name', label: 'Last Name' },
  { key: 'company', label: 'Company' },
  { key: 'title', label: 'Title' },
  { key: 'email', label: 'Email' },
  { key: 'phone', label: 'Phone' },
  { key: 'mobile', label: 'Mobile' },
  { key: 'source', label: 'Source' },
  { key: 'status', label: 'Status' },
  { key: 'rating', label: 'Rating', render: (v: unknown) => v ? String(v) : '' },
  { key: 'access', label: 'Access' },
  { key: 'background_info', label: 'Background Info' },
  { key: 'created_at', label: 'Created', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
  { key: 'updated_at', label: 'Updated', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
];

export function LeadDetailPage() {
  return (
    <EntityDetailPage<Lead>
      entityName="Lead"
      entitySlug="leads"
      endpoint="/leads"
      fields={detailFields}
      formFields={leadFields as FieldDef[]}
      getTitle={(l) => `${l.first_name} ${l.last_name}`}
    />
  );
}
