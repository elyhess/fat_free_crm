import { EntityDetailPage } from './EntityDetailPage';
import { accountFields } from '../config/entityFields';
import type { Account, Contact, Opportunity } from '../types/entities';
import type { FieldDef } from '../components/EntityForm';
import type { RelatedEntitySection } from '../components/RelatedEntities';

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

function relatedSections(id: string): RelatedEntitySection[] {
  return [
    {
      title: 'Contacts',
      endpoint: `/accounts/${id}/contacts`,
      columns: [
        { key: 'name', label: 'Name', render: (c: Contact) => `${c.first_name} ${c.last_name}` },
        { key: 'email', label: 'Email' },
        { key: 'phone', label: 'Phone' },
      ],
      linkPath: (c: Contact) => `/contacts/${c.id}`,
    } as RelatedEntitySection,
    {
      title: 'Opportunities',
      endpoint: `/accounts/${id}/opportunities`,
      columns: [
        { key: 'name', label: 'Name' },
        { key: 'stage', label: 'Stage' },
        { key: 'amount', label: 'Amount', render: (o: Opportunity) => o.amount != null ? `$${Number(o.amount).toLocaleString()}` : '' },
      ],
      linkPath: (o: Opportunity) => `/opportunities/${o.id}`,
    } as RelatedEntitySection,
  ];
}

export function AccountDetailPage() {
  const id = window.location.pathname.split('/').pop() ?? '';
  return (
    <EntityDetailPage<Account>
      entityName="Account"
      entitySlug="accounts"
      endpoint="/accounts"
      fields={detailFields}
      formFields={accountFields as FieldDef[]}
      getTitle={(a) => a.name}
      relatedEntities={relatedSections(id)}
    />
  );
}
