import { useEffect, useMemo, useState } from 'react';
import { createResource, deleteResource, fetchResource, fetchResources, updateResource } from '../lib/api.js';
import DynamicRecordForm from './DynamicRecordForm.jsx';
import { findDocTypeImageField, formatFieldValue, resolveFileURL } from '../lib/metadata.js';
import RecordFieldValue from './RecordFieldValue.jsx';
import SingleResourceWorkbench from './SingleResourceWorkbench.jsx';

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
  if (docType?.is_single) {
    return <SingleResourceWorkbench docType={docType} loading={loading} />;
  }

  return <CollectionResourceWorkbench docType={docType} loading={loading} />;
}

function CollectionResourceWorkbench({ docType, loading }) {
  const [records, setRecords] = useState([]);
  const [meta, setMeta] = useState(null);
  const [listBusy, setListBusy] = useState(false);
  const [detailBusy, setDetailBusy] = useState(false);
  const [mutationBusy, setMutationBusy] = useState(false);
  const [error, setError] = useState('');
  const [refreshKey, setRefreshKey] = useState(0);
  const [search, setSearch] = useState('');
  const [sortBy, setSortBy] = useState('updated_at');
  const [sortOrder, setSortOrder] = useState('desc');
  const [filterField, setFilterField] = useState('');
  const [filterOperator, setFilterOperator] = useState('eq');
  const [filterValue, setFilterValue] = useState('');
  const [selectedRecordName, setSelectedRecordName] = useState('');
  const [selectedRecord, setSelectedRecord] = useState(null);
  const [deskMode, setDeskMode] = useState('create');

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
    setListBusy(true);
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
        const nextRecords = payload.data || [];
        setRecords(nextRecords);
        setMeta(payload.meta || null);
      setSelectedRecordName((current) => {
        if (deskMode === 'create' && current === '') {
          return current;
        }
        const stillExists = nextRecords.some((record) => record.name === current);
        if (stillExists) {
          return current;
        }
        if (nextRecords[0]?.name) {
          setDeskMode('view');
          return nextRecords[0].name;
        }
        setDeskMode('create');
        return '';
      });
      })
      .catch((requestError) => {
        if (active) {
          setError(requestError.message);
        }
      })
      .finally(() => {
        if (active) {
          setListBusy(false);
        }
      });

    return () => {
      active = false;
    };
  }, [docType, refreshKey, search, sortBy, sortOrder, filterField, filterOperator, filterValue]);

  useEffect(() => {
    if (!docType?.doctype || !selectedRecordName || deskMode === 'create') {
      setSelectedRecord(null);
      return;
    }

    let active = true;
    setDetailBusy(true);
    fetchResource(docType.doctype, selectedRecordName)
      .then((record) => {
        if (active) {
          setSelectedRecord(record);
        }
      })
      .catch((requestError) => {
        if (active) {
          setError(requestError.message);
        }
      })
      .finally(() => {
        if (active) {
          setDetailBusy(false);
        }
      });

    return () => {
      active = false;
    };
  }, [docType, selectedRecordName, deskMode]);

  const listFields = useMemo(() => {
    if (!docType) {
      return [];
    }
    const visibleFields = (docType.fields || []).filter((field) => !field.hidden);
    const preferred = visibleFields.filter((field) => field.in_list_view);
    const fallback = visibleFields.filter((field) => !field.in_list_view);
    return [...preferred, ...fallback].slice(0, 5);
  }, [docType]);

  const listPreviewFields = useMemo(() => listFields.slice(0, 3), [listFields]);

  const detailFields = useMemo(() => (docType?.fields || []).filter((field) => !field.hidden), [docType]);

  const selectedRecordSummary = useMemo(
    () => records.find((record) => record.name === selectedRecordName) || null,
    [records, selectedRecordName],
  );

  const listGridTemplate = useMemo(
    () => ({ gridTemplateColumns: `minmax(0, 1.4fr) ${listPreviewFields.map(() => 'minmax(0, 1fr)').join(' ')}` }),
    [listPreviewFields],
  );

  const profileImageField = useMemo(() => findDocTypeImageField(docType), [docType]);
  const profileImageURL = useMemo(() => {
    if (!profileImageField || !selectedRecord) {
      return '';
    }
    return resolveFileURL(selectedRecord[profileImageField.fieldname]);
  }, [profileImageField, selectedRecord]);

  const handleCreate = async (payload) => {
    if (!docType) {
      return;
    }
    setMutationBusy(true);
    try {
      const created = await createResource(docType.doctype, payload);
      setRecords((current) => [created, ...current.filter((record) => record.name !== created.name)]);
      setSelectedRecordName(created.name);
      setSelectedRecord(created);
      setDeskMode('view');
    } finally {
      setMutationBusy(false);
    }
  };

  const handleUpdate = async (payload) => {
    if (!docType || !selectedRecordName) {
      return;
    }

    setMutationBusy(true);
    try {
      const updated = await updateResource(docType.doctype, selectedRecordName, payload);
      setRecords((current) => current.map((record) => (record.name === selectedRecordName ? { ...record, ...updated } : record)));
      setSelectedRecord(updated);
      setDeskMode('view');
    } finally {
      setMutationBusy(false);
    }
  };

  const handleDelete = async (name) => {
    if (!docType || !window.confirm(`Delete ${name}?`)) {
      return;
    }

    try {
      setMutationBusy(true);
      await deleteResource(docType.doctype, name);
      setRecords((current) => {
        const next = current.filter((record) => record.name !== name);
        if (selectedRecordName === name) {
          setSelectedRecordName(next[0]?.name || '');
          setSelectedRecord(next[0] || null);
          setDeskMode(next[0] ? 'view' : 'create');
        }
        return next;
      });
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setMutationBusy(false);
    }
  };

  const handleSelectRecord = (name) => {
    setSelectedRecordName(name);
    setDeskMode('view');
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
            <div className="flex flex-wrap items-center gap-2">
              <span className="badge">{docType.module || 'Core'} module</span>
              <span className="badge">Frappe-style desk</span>
              {docType.is_single ? <span className="badge">single</span> : null}
            </div>
            <h2 className="mt-3 text-xl font-semibold text-white">{docType.label} desk</h2>
            <p className="mt-1 text-sm text-slate-400">Compact list view on the left, generated form/detail desk on the right, aligned with Frappe’s record-first navigation model.</p>
          </div>
          <div className="flex flex-wrap gap-2 text-xs text-slate-300">
            <span className="badge">{records.length} visible rows</span>
            {meta ? <span className="badge">total {meta.total}</span> : null}
            <span className="badge">{selectedRecordName || 'new record'}</span>
          </div>
        </div>

        <div className="mt-5 grid gap-3 xl:grid-cols-[minmax(0,1.4fr)_repeat(4,minmax(0,0.9fr))]">
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

      <div className="grid gap-4 xl:grid-cols-[minmax(320px,0.76fr)_minmax(0,1.24fr)]">
      <div className="panel overflow-hidden">
        <div className="border-b border-white/10 px-5 py-4">
          <div className="flex items-center justify-between gap-3">
            <div>
              <h3 className="text-lg font-semibold text-white">List view</h3>
              <p className="mt-1 text-sm text-slate-400">Compact rows, fast scan, and direct jump into form/detail mode.</p>
            </div>
            <button
              type="button"
              onClick={() => {
                setDeskMode('create');
                setSelectedRecordName('');
                setSelectedRecord(null);
              }}
              className="rounded-xl bg-cyan-500 px-3 py-2 text-sm font-semibold text-slate-950 transition hover:bg-cyan-400"
            >
              New record
            </button>
          </div>

          <div style={listGridTemplate} className="mt-4 hidden items-center gap-3 border-t border-white/10 pt-3 text-[11px] font-semibold uppercase tracking-[0.22em] text-slate-500 md:grid">
            <span>Name</span>
            {listPreviewFields.map((field) => (
              <span key={`header-${field.fieldname}`} className="truncate">{field.label}</span>
            ))}
          </div>
        </div>

        <div className="max-h-[820px] overflow-y-auto">
          {listBusy ? (
            <div className="space-y-3 p-4">
              {Array.from({ length: 5 }).map((_, index) => <div key={index} className="skeleton-line h-18 rounded-2xl" />)}
            </div>
          ) : records.length > 0 ? (
            records.map((record) => (
              <button
                key={record.name}
                type="button"
                onClick={() => handleSelectRecord(record.name)}
                style={listGridTemplate}
                className={`grid w-full items-center gap-3 border-b border-white/6 px-4 py-3 text-left transition hover:bg-white/[0.05] ${
                  selectedRecordName === record.name && deskMode !== 'create' ? 'bg-cyan-500/10' : 'bg-transparent'
                }`}
              >
                <div className="min-w-0">
                  <div className="truncate text-sm font-semibold text-white">{record.name}</div>
                  <div className="mt-1 text-[11px] uppercase tracking-[0.2em] text-slate-500">{docType.label}</div>
                </div>
                {listPreviewFields.map((field) => (
                  <div key={`${record.name}-${field.fieldname}`} className="min-w-0 text-sm text-slate-300">
                    <div className="truncate">{formatFieldValue(field, record[field.fieldname])}</div>
                  </div>
                ))}
              </button>
            ))
          ) : (
            <div className="p-4">
              <div className="rounded-3xl border border-dashed border-white/10 bg-white/[0.03] px-5 py-8 text-center text-sm leading-6 text-slate-400">
                No records match the current search/filter state. Create the first record to populate this desk.
              </div>
            </div>
          )}
        </div>
      </div>

      <div className="space-y-4">
        {deskMode === 'create' ? (
        <DynamicRecordForm
          docType={docType}
          onSubmit={handleCreate}
          busy={mutationBusy}
          mode="create"
          heading={`Create ${docType.label}`}
          description="Generate a new document using a form rendered entirely from live metadata."
          submitLabel="Create record"
        />
        ) : detailBusy ? (
        <div className="panel p-6">
          <div className="space-y-3">
          <div className="skeleton-line h-16 rounded-3xl" />
          <div className="skeleton-line h-28 rounded-3xl" />
          <div className="skeleton-line h-28 rounded-3xl" />
          </div>
        </div>
        ) : deskMode === 'edit' && selectedRecord ? (
        <DynamicRecordForm
          docType={docType}
          initialValues={selectedRecord}
          onSubmit={handleUpdate}
          onCancel={() => setDeskMode('view')}
          busy={mutationBusy}
          mode="edit"
          heading={`Edit ${selectedRecord.name}`}
          description="Change values with the same metadata-driven form renderer the desk uses for create flows."
          submitLabel="Save changes"
        />
        ) : selectedRecord ? (
        <>
          <div className="panel p-5">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div>
            <div className="flex flex-wrap items-center gap-2">
              <span className="badge">Form view</span>
              <span className="badge">Metadata rendered</span>
            </div>
            <h3 className="mt-3 text-2xl font-semibold text-white">{selectedRecord.name}</h3>
            <p className="mt-2 text-sm leading-6 text-slate-400">
              Inspect the current document, review generated fields, and jump into edit mode without leaving the Studio shell.
            </p>
            </div>
            <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={() => setDeskMode('edit')}
              className="rounded-xl border border-cyan-400/20 bg-cyan-500/10 px-3 py-2 text-sm font-semibold text-cyan-100 transition hover:bg-cyan-500/20"
            >
              Edit record
            </button>
            <button
              type="button"
              onClick={() => handleDelete(selectedRecord.name)}
              className="rounded-xl border border-rose-400/20 bg-rose-500/10 px-3 py-2 text-sm font-semibold text-rose-100 transition hover:bg-rose-500/20"
            >
              Delete
            </button>
            </div>
          </div>

          <div className="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {detailFields.map((field) => (
            <div key={field.fieldname} className="rounded-3xl border border-white/10 bg-white/[0.03] p-4">
              <div className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">{field.label}</div>
              <div className="mt-1 text-[11px] uppercase tracking-[0.2em] text-slate-600">{field.fieldtype}</div>
						  <div className={`mt-4 text-sm ${field.fieldtype === 'JSON' ? 'text-cyan-200' : 'text-slate-100'}`}>
							<RecordFieldValue field={field} value={selectedRecord[field.fieldname]} />
						  </div>
            </div>
            ))}
          </div>
          </div>

          <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_320px]">
          <div className="panel p-5">
            <div className="text-xs font-semibold uppercase tracking-[0.24em] text-slate-500">Quick facts</div>
            <div className="mt-4 space-y-3 text-sm text-slate-300">
            <div className="flex items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
              <span>Name</span>
              <span className="font-mono text-cyan-200">{selectedRecord.name}</span>
            </div>
            <div className="flex items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
              <span>Created at</span>
              <span>{selectedRecord.created_at ? new Date(selectedRecord.created_at).toLocaleString() : '—'}</span>
            </div>
            <div className="flex items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
              <span>Updated at</span>
              <span>{selectedRecord.updated_at ? new Date(selectedRecord.updated_at).toLocaleString() : '—'}</span>
            </div>
            <div className="flex items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
              <span>Visible fields</span>
              <span>{detailFields.length}</span>
            </div>
            </div>
          </div>

          <div className="panel p-5">
      {profileImageURL ? (
        <div className="mb-4 rounded-3xl border border-cyan-400/15 bg-cyan-500/10 p-4">
        <div className="text-xs font-semibold uppercase tracking-[0.24em] text-cyan-100/80">Profile preview</div>
        <div className="mt-3 flex items-center gap-4">
          <img src={profileImageURL} alt={profileImageField?.label || docType.label} className="h-24 w-24 rounded-3xl object-cover ring-1 ring-white/10" />
          <div>
          <div className="text-sm font-semibold text-white">{profileImageField?.label || 'Image field'}</div>
          <div className="mt-1 text-xs text-cyan-100/75">Driven from DocType `image_field` metadata, Frappe-style.</div>
          </div>
        </div>
        </div>
      ) : null}
            <div className="text-xs font-semibold uppercase tracking-[0.24em] text-slate-500">Version history & audit trail</div>
            <div className="mt-4 space-y-3">
            <div className="rounded-2xl border border-emerald-400/20 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-100">
              <div className="font-semibold">Record loaded</div>
              <div className="mt-1 text-xs text-emerald-100/80">Studio desk hydrated {selectedRecordSummary?.updated_at ? 'from the latest API payload' : 'from the live list state'}.</div>
            </div>
            <div className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3 text-sm text-slate-300">
              <div className="font-semibold text-white">Created</div>
              <div className="mt-1 text-xs text-slate-500">{selectedRecord.created_at ? new Date(selectedRecord.created_at).toLocaleString() : 'Timestamp pending'}</div>
            </div>
            <div className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3 text-sm text-slate-300">
              <div className="font-semibold text-white">Last updated</div>
              <div className="mt-1 text-xs text-slate-500">{selectedRecord.updated_at ? new Date(selectedRecord.updated_at).toLocaleString() : 'Timestamp pending'}</div>
            </div>
            <div className="rounded-2xl border border-dashed border-white/10 bg-white/[0.03] px-4 py-3 text-sm text-slate-400">
              Full diff history will slot here once the backend exposes version/audit APIs. The desk layout is ready for it already.
            </div>
            </div>
          </div>
          </div>
        </>
        ) : (
        <div className="panel p-6 text-sm text-slate-300">Choose a record from the list or create a new one to open the desk.</div>
        )}
      </div>
      </div>
    </section>
  );
}
