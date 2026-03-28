import { useState } from 'react';
import { EntityDetailPage } from './EntityDetailPage';
import { leadFields } from '../config/entityFields';
import { ConvertLeadForm } from '../components/ConvertLeadForm';
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
  const [showConvert, setShowConvert] = useState(false);

  return (
    <>
      <EntityDetailPage<Lead>
        entityName="Lead"
        entitySlug="leads"
        endpoint="/leads"
        fields={detailFields}
        formFields={leadFields as FieldDef[]}
        getTitle={(l) => `${l.first_name} ${l.last_name}`}
        customActions={(lead) =>
          lead.status !== 'converted' ? (
            <>
              <button
                onClick={() => setShowConvert(true)}
                className="px-4 py-2 text-sm bg-green-600 text-white rounded-md hover:bg-green-700"
              >
                Convert
              </button>
              <ConvertLeadForm
                lead={lead}
                open={showConvert}
                onClose={() => setShowConvert(false)}
              />
            </>
          ) : (
            <span className="px-3 py-2 text-sm text-green-700 bg-green-100 rounded-md font-medium">
              Converted
            </span>
          )
        }
      />
    </>
  );
}
