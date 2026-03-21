import { EntityDetailPage } from './EntityDetailPage';
import { contactFields } from '../config/entityFields';
import type { Contact } from '../types/entities';
import type { FieldDef } from '../components/EntityForm';

const detailFields = [
  { key: 'first_name', label: 'First Name' },
  { key: 'last_name', label: 'Last Name' },
  { key: 'title', label: 'Title' },
  { key: 'department', label: 'Department' },
  { key: 'email', label: 'Email' },
  { key: 'phone', label: 'Phone' },
  { key: 'mobile', label: 'Mobile' },
  { key: 'access', label: 'Access' },
  { key: 'background_info', label: 'Background Info' },
  { key: 'created_at', label: 'Created', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
  { key: 'updated_at', label: 'Updated', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
];

export function ContactDetailPage() {
  return (
    <EntityDetailPage<Contact>
      entityName="Contact"
      entitySlug="contacts"
      endpoint="/contacts"
      fields={detailFields}
      formFields={contactFields as FieldDef[]}
      getTitle={(c) => `${c.first_name} ${c.last_name}`}
    />
  );
}
