import { EntityDetailPage } from './EntityDetailPage';
import { opportunityFields } from '../config/entityFields';
import type { Opportunity } from '../types/entities';
import type { FieldDef } from '../components/EntityForm';

const detailFields = [
  { key: 'name', label: 'Name' },
  { key: 'stage', label: 'Stage' },
  { key: 'amount', label: 'Amount', render: (v: unknown) => v != null ? `$${Number(v).toLocaleString()}` : '' },
  { key: 'probability', label: 'Probability', render: (v: unknown) => v != null ? `${v}%` : '' },
  { key: 'discount', label: 'Discount', render: (v: unknown) => v != null ? `$${Number(v).toLocaleString()}` : '' },
  { key: 'closes_on', label: 'Closes On', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
  { key: 'source', label: 'Source' },
  { key: 'access', label: 'Access' },
  { key: 'background_info', label: 'Background Info' },
  { key: 'created_at', label: 'Created', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
  { key: 'updated_at', label: 'Updated', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
];

export function OpportunityDetailPage() {
  return (
    <EntityDetailPage<Opportunity>
      entityName="Opportunity"
      entitySlug="opportunities"
      endpoint="/opportunities"
      fields={detailFields}
      formFields={opportunityFields as FieldDef[]}
      getTitle={(o) => o.name}
    />
  );
}
