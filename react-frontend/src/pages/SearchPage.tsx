import { useSearchParams, Link } from 'react-router-dom';
import { useApi } from '../hooks/useApi';

interface SearchResult {
  query: string;
  accounts: { id: number; name: string; email?: string }[] | null;
  contacts: { id: number; first_name: string; last_name: string; email?: string }[] | null;
  leads: { id: number; first_name: string; last_name: string; company?: string }[] | null;
  opportunities: { id: number; name: string; stage?: string; amount?: number }[] | null;
  campaigns: { id: number; name: string; status?: string }[] | null;
  tasks: { id: number; name: string; bucket?: string }[] | null;
  total_count: number;
}

const entityTypes = [
  { value: '', label: 'All' },
  { value: 'accounts', label: 'Accounts' },
  { value: 'contacts', label: 'Contacts' },
  { value: 'leads', label: 'Leads' },
  { value: 'opportunities', label: 'Opportunities' },
  { value: 'campaigns', label: 'Campaigns' },
  { value: 'tasks', label: 'Tasks' },
];

function ResultSection({ title, children, count }: { title: string; children: React.ReactNode; count: number }) {
  if (count === 0) return null;
  return (
    <div className="mb-6">
      <h2 className="text-lg font-semibold text-gray-800 mb-2">
        {title} <span className="text-sm font-normal text-gray-500">({count})</span>
      </h2>
      <div className="bg-white shadow rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <tbody className="divide-y divide-gray-200">{children}</tbody>
        </table>
      </div>
    </div>
  );
}

function Row({ cells, link }: { cells: string[]; link?: string }) {
  return (
    <tr className="hover:bg-gray-50">
      {cells.map((cell, i) => (
        <td key={i} className="px-6 py-3 text-sm text-gray-900">
          {i === 0 && link ? (
            <Link to={link} className="text-blue-600 hover:text-blue-800 hover:underline">{cell}</Link>
          ) : (
            cell
          )}
        </td>
      ))}
    </tr>
  );
}

export function SearchPage() {
  const [params, setParams] = useSearchParams();
  const q = params.get('q') || '';
  const entity = params.get('entity') || '';

  const apiPath = q
    ? `/search?q=${encodeURIComponent(q)}${entity ? `&entity=${entity}` : ''}`
    : '';
  const { data, loading, error } = useApi<SearchResult>(apiPath);

  function handleEntityChange(newEntity: string) {
    const next = new URLSearchParams(params);
    if (newEntity) {
      next.set('entity', newEntity);
    } else {
      next.delete('entity');
    }
    setParams(next);
  }

  if (!q) {
    return <div className="text-gray-500">Enter a search term above.</div>;
  }

  return (
    <div>
      <div className="flex items-center gap-4 mb-6">
        <h1 className="text-2xl font-semibold text-gray-900">Search Results</h1>
        <select
          value={entity}
          onChange={(e) => handleEntityChange(e.target.value)}
          className="px-3 py-1.5 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          {entityTypes.map((et) => (
            <option key={et.value} value={et.value}>{et.label}</option>
          ))}
        </select>
      </div>
      <p className="text-sm text-gray-500 mb-6">
        {loading ? 'Searching...' : data ? `${data.total_count} results for "${q}"` : ''}
      </p>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      {data && (
        <>
          <ResultSection title="Accounts" count={data.accounts?.length ?? 0}>
            {data.accounts?.map((a) => (
              <Row key={a.id} cells={[a.name, a.email ?? '']} link={`/accounts/${a.id}`} />
            ))}
          </ResultSection>

          <ResultSection title="Contacts" count={data.contacts?.length ?? 0}>
            {data.contacts?.map((c) => (
              <Row key={c.id} cells={[`${c.first_name} ${c.last_name}`, c.email ?? '']} link={`/contacts/${c.id}`} />
            ))}
          </ResultSection>

          <ResultSection title="Leads" count={data.leads?.length ?? 0}>
            {data.leads?.map((l) => (
              <Row key={l.id} cells={[`${l.first_name} ${l.last_name}`, l.company ?? '']} link={`/leads/${l.id}`} />
            ))}
          </ResultSection>

          <ResultSection title="Opportunities" count={data.opportunities?.length ?? 0}>
            {data.opportunities?.map((o) => (
              <Row key={o.id} cells={[o.name, o.stage ?? '', o.amount != null ? `$${o.amount}` : '']} link={`/opportunities/${o.id}`} />
            ))}
          </ResultSection>

          <ResultSection title="Campaigns" count={data.campaigns?.length ?? 0}>
            {data.campaigns?.map((c) => (
              <Row key={c.id} cells={[c.name, c.status ?? '']} link={`/campaigns/${c.id}`} />
            ))}
          </ResultSection>

          <ResultSection title="Tasks" count={data.tasks?.length ?? 0}>
            {data.tasks?.map((t) => (
              <Row key={t.id} cells={[t.name, t.bucket ?? '']} link={`/tasks/${t.id}`} />
            ))}
          </ResultSection>

          {data.total_count === 0 && (
            <div className="text-center text-gray-500 py-8">
              No results found for "{q}".
            </div>
          )}
        </>
      )}
    </div>
  );
}
