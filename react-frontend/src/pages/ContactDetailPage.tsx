import { EntityDetailPage } from './EntityDetailPage';
import { contactFields } from '../config/entityFields';
import type { Contact, Opportunity } from '../types/entities';
import type { FieldDef } from '../components/EntityForm';
import type { RelatedEntitySection } from '../components/RelatedEntities';

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

function relatedSections(id: string): RelatedEntitySection[] {
  return [
    {
      title: 'Opportunities',
      endpoint: `/contacts/${id}/opportunities`,
      columns: [
        { key: 'name', label: 'Name' },
        { key: 'stage', label: 'Stage' },
        { key: 'amount', label: 'Amount', render: (o: Opportunity) => o.amount != null ? `$${Number(o.amount).toLocaleString()}` : '' },
      ],
      linkPath: (o: Opportunity) => `/opportunities/${o.id}`,
    },
  ];
}

export function ContactDetailPage() {
  const id = window.location.pathname.split('/').pop() ?? '';
  return (
    <EntityDetailPage<Contact>
      entityName="Contact"
      entitySlug="contacts"
      entityType="Contact"
      endpoint="/contacts"
      fields={detailFields}
      formFields={contactFields as FieldDef[]}
      getTitle={(c) => `${c.first_name} ${c.last_name}`}
      relatedEntities={relatedSections(id)}
    />
  );
}
