import { EntityDetailPage } from './EntityDetailPage';
import { campaignFields } from '../config/entityFields';
import type { Campaign, Lead, Opportunity } from '../types/entities';
import type { FieldDef } from '../components/EntityForm';
import type { RelatedEntitySection } from '../components/RelatedEntities';

const detailFields = [
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status' },
  { key: 'budget', label: 'Budget', render: (v: unknown) => v != null ? `$${Number(v).toLocaleString()}` : '' },
  { key: 'target_leads', label: 'Target Leads', render: (v: unknown) => v != null ? String(v) : '' },
  { key: 'target_conversion', label: 'Target Conversion', render: (v: unknown) => v != null ? `${v}%` : '' },
  { key: 'target_revenue', label: 'Target Revenue', render: (v: unknown) => v != null ? `$${Number(v).toLocaleString()}` : '' },
  { key: 'starts_on', label: 'Starts On', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
  { key: 'ends_on', label: 'Ends On', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
  { key: 'leads_count', label: 'Leads', render: (v: unknown) => v != null ? String(v) : '0' },
  { key: 'opportunities_count', label: 'Opportunities', render: (v: unknown) => v != null ? String(v) : '0' },
  { key: 'access', label: 'Access' },
  { key: 'objectives', label: 'Objectives' },
  { key: 'background_info', label: 'Background Info' },
  { key: 'created_at', label: 'Created', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
  { key: 'updated_at', label: 'Updated', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
];

function relatedSections(id: string): RelatedEntitySection[] {
  return [
    {
      title: 'Leads',
      endpoint: `/campaigns/${id}/leads`,
      columns: [
        { key: 'name', label: 'Name', render: (l: Lead) => `${l.first_name} ${l.last_name}` },
        { key: 'status', label: 'Status' },
        { key: 'email', label: 'Email' },
      ],
      linkPath: (l: Lead) => `/leads/${l.id}`,
    } as RelatedEntitySection,
    {
      title: 'Opportunities',
      endpoint: `/campaigns/${id}/opportunities`,
      columns: [
        { key: 'name', label: 'Name' },
        { key: 'stage', label: 'Stage' },
        { key: 'amount', label: 'Amount', render: (o: Opportunity) => o.amount != null ? `$${Number(o.amount).toLocaleString()}` : '' },
      ],
      linkPath: (o: Opportunity) => `/opportunities/${o.id}`,
    } as RelatedEntitySection,
  ];
}

export function CampaignDetailPage() {
  const id = window.location.pathname.split('/').pop() ?? '';
  return (
    <EntityDetailPage<Campaign>
      entityName="Campaign"
      entitySlug="campaigns"
      entityType="Campaign"
      endpoint="/campaigns"
      fields={detailFields}
      formFields={campaignFields as FieldDef[]}
      getTitle={(c) => c.name}
      relatedEntities={relatedSections(id)}
    />
  );
}
