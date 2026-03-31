import type { FieldDef } from '../components/EntityForm';

const accessOptions = [
  { value: 'Public', label: 'Public' },
  { value: 'Private', label: 'Private' },
  { value: 'Shared', label: 'Shared' },
];

export const accountFields: FieldDef[] = [
  { key: 'name', label: 'Name', required: true },
  { key: 'email', label: 'Email', type: 'email' },
  { key: 'phone', label: 'Phone', type: 'tel' },
  { key: 'website', label: 'Website' },
  { key: 'fax', label: 'Fax', type: 'tel' },
  { key: 'category', label: 'Category', type: 'select', options: [
    { value: 'affiliate', label: 'Affiliate' },
    { value: 'competitor', label: 'Competitor' },
    { value: 'customer', label: 'Customer' },
    { value: 'partner', label: 'Partner' },
    { value: 'reseller', label: 'Reseller' },
    { value: 'vendor', label: 'Vendor' },
  ]},
  { key: 'rating', label: 'Rating', type: 'number' },
  { key: 'access', label: 'Access', type: 'select', options: accessOptions },
  { key: 'background_info', label: 'Background Info', type: 'textarea' },
];

export const contactFields: FieldDef[] = [
  { key: 'first_name', label: 'First Name', required: true },
  { key: 'last_name', label: 'Last Name', required: true },
  { key: 'account_id', label: 'Account', type: 'autocomplete', entity: 'accounts' },
  { key: 'title', label: 'Title' },
  { key: 'department', label: 'Department' },
  { key: 'email', label: 'Email', type: 'email' },
  { key: 'alt_email', label: 'Alt Email', type: 'email' },
  { key: 'phone', label: 'Phone', type: 'tel' },
  { key: 'mobile', label: 'Mobile', type: 'tel' },
  { key: 'fax', label: 'Fax', type: 'tel' },
  { key: 'access', label: 'Access', type: 'select', options: accessOptions },
  { key: 'background_info', label: 'Background Info', type: 'textarea' },
];

export const leadFields: FieldDef[] = [
  { key: 'first_name', label: 'First Name', required: true },
  { key: 'last_name', label: 'Last Name', required: true },
  { key: 'company', label: 'Company' },
  { key: 'title', label: 'Title' },
  { key: 'email', label: 'Email', type: 'email' },
  { key: 'phone', label: 'Phone', type: 'tel' },
  { key: 'mobile', label: 'Mobile', type: 'tel' },
  { key: 'source', label: 'Source', type: 'select', options: [
    { value: 'campaign', label: 'Campaign' },
    { value: 'cold_call', label: 'Cold Call' },
    { value: 'conference', label: 'Conference' },
    { value: 'online', label: 'Online' },
    { value: 'referral', label: 'Referral' },
    { value: 'self', label: 'Self' },
    { value: 'web', label: 'Web' },
    { value: 'word_of_mouth', label: 'Word of Mouth' },
    { value: 'other', label: 'Other' },
  ]},
  { key: 'status', label: 'Status', type: 'select', options: [
    { value: 'new', label: 'New' },
    { value: 'contacted', label: 'Contacted' },
    { value: 'converted', label: 'Converted' },
    { value: 'rejected', label: 'Rejected' },
  ]},
  { key: 'rating', label: 'Rating', type: 'number' },
  { key: 'access', label: 'Access', type: 'select', options: accessOptions },
  { key: 'background_info', label: 'Background Info', type: 'textarea' },
];

export const opportunityFields: FieldDef[] = [
  { key: 'name', label: 'Name', required: true },
  { key: 'account_id', label: 'Account', type: 'autocomplete', entity: 'accounts' },
  { key: 'campaign_id', label: 'Campaign', type: 'autocomplete', entity: 'campaigns' },
  { key: 'stage', label: 'Stage', type: 'select', options: [
    { value: 'prospecting', label: 'Prospecting' },
    { value: 'analysis', label: 'Analysis' },
    { value: 'presentation', label: 'Presentation' },
    { value: 'proposal', label: 'Proposal' },
    { value: 'negotiation', label: 'Negotiation' },
    { value: 'final_review', label: 'Final Review' },
    { value: 'won', label: 'Won' },
    { value: 'lost', label: 'Lost' },
  ]},
  { key: 'amount', label: 'Amount', type: 'number' },
  { key: 'probability', label: 'Probability (%)', type: 'number' },
  { key: 'discount', label: 'Discount', type: 'number' },
  { key: 'closes_on', label: 'Closes On', type: 'date' },
  { key: 'source', label: 'Source', type: 'select', options: [
    { value: 'campaign', label: 'Campaign' },
    { value: 'cold_call', label: 'Cold Call' },
    { value: 'conference', label: 'Conference' },
    { value: 'online', label: 'Online' },
    { value: 'referral', label: 'Referral' },
    { value: 'self', label: 'Self' },
    { value: 'web', label: 'Web' },
    { value: 'word_of_mouth', label: 'Word of Mouth' },
    { value: 'other', label: 'Other' },
  ]},
  { key: 'access', label: 'Access', type: 'select', options: accessOptions },
  { key: 'background_info', label: 'Background Info', type: 'textarea' },
];

export const campaignFields: FieldDef[] = [
  { key: 'name', label: 'Name', required: true },
  { key: 'status', label: 'Status', type: 'select', options: [
    { value: 'planned', label: 'Planned' },
    { value: 'started', label: 'Started' },
    { value: 'completed', label: 'Completed' },
    { value: 'on_hold', label: 'On Hold' },
    { value: 'called_off', label: 'Called Off' },
  ]},
  { key: 'budget', label: 'Budget', type: 'number' },
  { key: 'target_leads', label: 'Target Leads', type: 'number' },
  { key: 'target_conversion', label: 'Target Conversion (%)', type: 'number' },
  { key: 'target_revenue', label: 'Target Revenue', type: 'number' },
  { key: 'starts_on', label: 'Starts On', type: 'date' },
  { key: 'ends_on', label: 'Ends On', type: 'date' },
  { key: 'access', label: 'Access', type: 'select', options: accessOptions },
  { key: 'objectives', label: 'Objectives', type: 'textarea' },
  { key: 'background_info', label: 'Background Info', type: 'textarea' },
];

export const taskFields: FieldDef[] = [
  { key: 'name', label: 'Name', required: true },
  { key: 'priority', label: 'Priority', type: 'select', options: [
    { value: 'high', label: 'High' },
    { value: 'medium', label: 'Medium' },
    { value: 'low', label: 'Low' },
  ]},
  { key: 'category', label: 'Category', type: 'select', options: [
    { value: 'call', label: 'Call' },
    { value: 'email', label: 'Email' },
    { value: 'follow_up', label: 'Follow Up' },
    { value: 'lunch', label: 'Lunch' },
    { value: 'meeting', label: 'Meeting' },
    { value: 'money', label: 'Money' },
    { value: 'presentation', label: 'Presentation' },
    { value: 'trip', label: 'Trip' },
  ]},
  { key: 'bucket', label: 'Due', type: 'select', options: [
    { value: 'due_asap', label: 'ASAP' },
    { value: 'due_today', label: 'Today' },
    { value: 'due_tomorrow', label: 'Tomorrow' },
    { value: 'due_this_week', label: 'This Week' },
    { value: 'due_next_week', label: 'Next Week' },
    { value: 'due_later', label: 'Later' },
  ]},
  { key: 'background_info', label: 'Notes', type: 'textarea' },
];
