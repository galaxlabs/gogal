import { useEffect, useMemo, useState } from 'react';
import { createResource, deleteResource, fetchResources } from '../lib/api.js';
import DynamicRecordForm from './DynamicRecordForm.jsx';

const operatorOptions = [
  { value: 'eq', label: '=' },
  { value: 'ne', label: '!=' },
  { value: 'gt', label: '>' },
  { value: 'gte', label: '>=' },
  { value: 'lt', label: '<' },
  { value: 'lte', label: '<=' },
  { value: 'like', label: 'like' },
  { value: 'ilike', label: 'ilike' },
  { value: 'in', label: 'in' },
  { value: 'isnull', label: 'is null' },
];

export default function ResourceWorkbench({ docType, loading }) {
  const [records, setRecords] = useState([]);
  const [meta, setMeta] = useState(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [refreshKey, setRefreshKey] = useState(0);
  const [search, setSearch] = useState('');
  const [sortBy, setSortBy] = useState('updated_at');
  const [sortOrder, setSortOrder] = useState('desc');
  const [filterField, setFilterField] = useState('');
  const [filterOperator, setFilterOperator] = useState('eq');
  const [filterValue, setFilterValue] = useState('');

  const queryableFields = useMemo(() => {
    if (!docType) {
      return [];
    }
    return [
      { fieldname: 'name', label: 'Name', fieldtype: 'Data' },
      { fieldname: 'created_at', label: 'Created At', fieldtype: 'Datetime' },
      { fieldname: 'updated_at', label: 'Updated At', fieldtype: 'Datetime' },
      ...(docType.fields || []),
    ];
  }, [docType]);

  useEffect(() => {
    if (!docType?.doctype) {
      return;
    }

    let active = true;
    setBusy(true);
    setError('');

    const query = {
      search,
      sort_by: sortBy,
      sort_order: sortOrder,
    };

    if (filterField) {
      const key = filterOperator === 'eq' ? `filter_${filterField}` : `filter_${filterField}__${filterOperator}`;
      query[key] = filterValue || (filterOperator === 'isnull' ? 'true' : '');
    }

    fetchResources(docType.doctype, query)
      .then((payload) => {
        if (!active) {
          return;
        }
        setRecords(payload.data || []);
        setMeta(payload.meta || null);
      })
      .catch((requestError) => {
        if (active) {
          setError(requestError.message);
        }
      })
      .finally(() => {
        if (active) {
          setBusy(false);
        }
      });

    return () => {
      active = false;
    };
  }, [docType, refreshKey, search, sortBy, sortOrder, filterField, filterOperator, filterValue]);

  const columns = useMemo(() => {
    if (!docType) {
      return [];
    }
    return ['name', ...(docType.fields || []).slice(0, 6).map((field) => field.fieldname)];
  }, [docType]);

  const handleCreate = async (payload) => {
    if (!docType) {
      return;
    }
    await createResource(docType.doctype, payload);
    setRefreshKey((current) => current + 1);
  };

  const handleDelete = async (name) => {
    if (!docType || !window.confirm(`Delete ${name}?`)) {
      return;
    }

    try {
      setBusy(true);
      await deleteResource(docType.doctype, name);
      setRefreshKey((current) => current + 1);
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setBusy(false);
    }
  };

  if (loading) {
    return <section className="panel p-6 text-sm text-slate-300">Loading live records…</section>;
  }

  if (!docType) {
    return <section className="panel p-6 text-sm text-slate-300">Choose a DocType to view records and generated forms.</section>;
  }

  return (
    <section className="flex min-h-[70vh] flex-col gap-4">
      <div className="panel p-5">
        <div className="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
          <div>
            <h2 className="text-xl font-semibold text-white">{docType.label} workspace</h2>
            <p className="mt-1 text-sm text-slate-400">Search, filter, sort, and create records against the live metadata-driven API.</p>
          </div>
          <div className="flex flex-wrap gap-2 text-xs text-slate-300">
            <span className="badge">{records.length} visible rows</span>
            {meta ? <span className="badge">total {meta.total}</span> : null}
          </div>
        </div>

        <div className="mt-5 grid gap-3 xl:grid-cols-[minmax(0,1.3fr)_repeat(4,minmax(0,0.9fr))]">
          <label className="grid gap-2">
            <span className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">Search</span>
            <input className="field" value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search name or text-like fields" />
          </label>

          <label className="grid gap-2">
            <span className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">Sort by</span>
            <select className="field" value={sortBy} onChange={(event) => setSortBy(event.target.value)}>
              {queryableFields.map((field) => (
                <option key={field.fieldname} value={field.fieldname}>{field.label}</option>
              ))}
            </select>
          </label>

          <label className="grid gap-2">
            <span className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">Order</span>
            <select className="field" value={sortOrder} onChange={(event) => setSortOrder(event.target.value)}>
              <option value="asc">Ascending</option>
              <option value="desc">Descending</option>
            </select>
          </label>

          <label className="grid gap-2">
            <span className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">Filter field</span>
            <select className="field" value={filterField} onChange={(event) => setFilterField(event.target.value)}>
              <option value="">No filter</option>
              {queryableFields.map((field) => (
                <option key={field.fieldname} value={field.fieldname}>{field.label}</option>
              ))}
            </select>
          </label>

          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-2">
            <label className="grid gap-2">
              <span className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">Operator</span>
              <select className="field" value={filterOperator} onChange={(event) => setFilterOperator(event.target.value)}>
                {operatorOptions.map((option) => (
                  <option key={option.value} value={option.value}>{option.label}</option>
                ))}
              </select>
            </label>
            <label className="grid gap-2">
              <span className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">Value</span>
              <input
                className="field"
                value={filterValue}
                onChange={(event) => setFilterValue(event.target.value)}
                placeholder={filterOperator === 'in' ? '1,2,3' : filterOperator === 'isnull' ? 'true / false' : 'Filter value'}
                disabled={!filterField}
              />
            </label>
          </div>
        </div>

        {error ? <div className="mt-4 rounded-xl border border-rose-400/20 bg-rose-500/10 px-3 py-2 text-sm text-rose-100">{error}</div> : null}
      </div>

      <div className="grid gap-4 xl:grid-cols-[minmax(0,1.2fr)_minmax(360px,0.9fr)]">
        <div className="panel overflow-hidden">
          <div className="border-b border-white/10 px-5 py-4">
            <h3 className="text-lg font-semibold text-white">Live records</h3>
          </div>

          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-white/10 text-sm">
              <thead className="bg-white/5 text-left text-slate-400">
                <tr>
                  {columns.map((column) => (
                    <th key={column} className="px-4 py-3 font-medium uppercase tracking-[0.12em]">{column}</th>
                  ))}
                  <th className="px-4 py-3 font-medium uppercase tracking-[0.12em]">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/5">
                {records.map((record) => (
                  <tr key={record.name} className="hover:bg-white/[0.03]">
                    {columns.map((column) => (
                      <td key={`${record.name}-${column}`} className="max-w-[220px] px-4 py-3 align-top text-slate-200">
                        <div className="truncate">{String(record[column] ?? '')}</div>
                      </td>
                    ))}
                    <td className="px-4 py-3">
                      <button
                        type="button"
                        onClick={() => handleDelete(record.name)}
                        className="rounded-lg border border-rose-400/20 bg-rose-500/10 px-3 py-1.5 text-xs font-semibold text-rose-100 transition hover:bg-rose-500/20"
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
                {!busy && records.length === 0 ? (
                  <tr>
                    <td colSpan={columns.length + 1} className="px-4 py-8 text-center text-slate-400">
                      No records match the current search/filter state.
                    </td>
                  </tr>
                ) : null}
              </tbody>
            </table>
          </div>
        </div>

        <DynamicRecordForm docType={docType} onSubmit={handleCreate} busy={busy} />
      </div>
    </section>
  );
}
