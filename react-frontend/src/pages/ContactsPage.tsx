import { EntityList } from '../components/EntityList';
import { contactFields } from '../config/entityFields';
import type { Contact } from '../types/entities';

const columns = [
  {
    key: 'last_name',
    label: 'Name',
    render: (c: Contact) => `${c.first_name} ${c.last_name}`,
  },
  { key: 'title', label: 'Title' },
  { key: 'email', label: 'Email' },
  { key: 'phone', label: 'Phone' },
  { key: 'department', label: 'Department' },
];

export function ContactsPage() {
  return (
    <EntityList<Contact>
      title="Contacts"
      endpoint="/contacts"
      columns={columns}
      getRowKey={(c) => c.id}
      formFields={contactFields}
    />
  );
}
