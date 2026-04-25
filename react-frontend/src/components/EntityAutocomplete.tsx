import { useState, useEffect, useRef } from 'react';
import { api } from '../api/client';

interface AutocompleteItem {
  id: number;
  name: string;
}

interface EntityAutocompleteProps {
  entity: string;
  value: number | string | null;
  onChange: (id: number | null) => void;
  placeholder?: string;
  required?: boolean;
}

export function EntityAutocomplete({
  entity,
  value,
  onChange,
  placeholder = 'Search...',
  required = false,
}: EntityAutocompleteProps) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<AutocompleteItem[]>([]);
  const [open, setOpen] = useState(false);
  const [selectedLabel, setSelectedLabel] = useState('');
  const [loading, setLoading] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  // Resolve initial value to a label
  useEffect(() => {
    if (value && !selectedLabel) {
      api
        .get<AutocompleteItem[]>(`/${entity}/autocomplete?q=`)
        .then((items) => {
          const match = items.find((i) => i.id === Number(value));
          if (match) setSelectedLabel(match.name);
        })
        .catch(() => {});
    }
  }, [value, entity, selectedLabel]);

  // Close dropdown on outside click
  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  function handleInputChange(q: string) {
    setQuery(q);
    setSelectedLabel('');
    onChange(null);

    if (debounceRef.current) clearTimeout(debounceRef.current);

    if (q.length < 1) {
      setResults([]);
      setOpen(false);
      return;
    }

    debounceRef.current = setTimeout(() => {
      setLoading(true);
      api
        .get<AutocompleteItem[]>(`/${entity}/autocomplete?q=${encodeURIComponent(q)}`)
        .then((items) => {
          setResults(items);
          setOpen(true);
        })
        .catch(() => setResults([]))
        .finally(() => setLoading(false));
    }, 200);
  }

  function handleSelect(item: AutocompleteItem) {
    setSelectedLabel(item.name);
    setQuery('');
    setResults([]);
    setOpen(false);
    onChange(item.id);
  }

  function handleClear() {
    setSelectedLabel('');
    setQuery('');
    setResults([]);
    onChange(null);
  }

  return (
    <div ref={containerRef} className="relative">
      {selectedLabel ? (
        <div className="flex items-center w-full border border-gray-300 rounded-md px-3 py-2 text-sm bg-white">
          <span className="flex-1 truncate">{selectedLabel}</span>
          <button
            type="button"
            onClick={handleClear}
            className="ml-2 text-gray-400 hover:text-gray-600 text-xs"
            aria-label="Clear selection"
          >
            &#x2715;
          </button>
        </div>
      ) : (
        <input
          type="text"
          value={query}
          onChange={(e) => handleInputChange(e.target.value)}
          onFocus={() => query.length >= 1 && results.length > 0 && setOpen(true)}
          placeholder={placeholder}
          required={required && !value}
          className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      )}

      {open && results.length > 0 && (
        <ul className="absolute z-50 mt-1 w-full bg-white border border-gray-200 rounded-md shadow-lg max-h-48 overflow-y-auto">
          {results.map((item) => (
            <li key={item.id}>
              <button
                type="button"
                onClick={() => handleSelect(item)}
                className="w-full text-left px-3 py-2 text-sm hover:bg-blue-50 focus:bg-blue-50"
              >
                {item.name}
              </button>
            </li>
          ))}
        </ul>
      )}

      {open && !loading && results.length === 0 && query.length >= 1 && (
        <div className="absolute z-50 mt-1 w-full bg-white border border-gray-200 rounded-md shadow-lg px-3 py-2 text-sm text-gray-500">
          No results found
        </div>
      )}
    </div>
  );
}
