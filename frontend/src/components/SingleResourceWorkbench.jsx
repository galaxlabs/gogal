import { useEffect, useMemo, useState } from 'react';
import { deleteSingleResource, fetchSingleResource, updateSingleResource } from '../lib/api.js';
import DynamicRecordForm from './DynamicRecordForm.jsx';
import RecordFieldValue from './RecordFieldValue.jsx';

export default function SingleResourceWorkbench({ docType, loading }) {
  const [record, setRecord] = useState(null);
  const [busy, setBusy] = useState(false);
  const [loadBusy, setLoadBusy] = useState(false);
  const [error, setError] = useState('');
  const [exists, setExists] = useState(false);

  const detailFields = useMemo(() => (docType?.fields || []).filter((field) => !field.hidden), [docType]);

  useEffect(() => {
    if (!docType?.doctype) {
      setRecord(null);
      return;
    }

    let active = true;
    setLoadBusy(true);
    setError('');

    fetchSingleResource(docType.doctype)
      .then((data) => {
        if (!active) {
          return;
        }
        setRecord({ name: docType.doctype, ...(data || {}) });
        setExists(Boolean(data?.created_at || data?.updated_at));
      })
      .catch((requestError) => {
        if (active) {
          setError(requestError.message);
        }
      })
      .finally(() => {
        if (active) {
          setLoadBusy(false);
        }
      });

    return () => {
      active = false;
    };
  }, [docType]);

  const handleSave = async (payload) => {
    if (!docType) {
      return;
    }

    setBusy(true);
    try {
      const updated = await updateSingleResource(docType.doctype, payload);
      setRecord({ name: docType.doctype, ...(updated || {}) });
      setExists(true);
      setError('');
    } catch (requestError) {
      setError(requestError.message);
      throw requestError;
    } finally {
      setBusy(false);
    }
  };

  const handleReset = async () => {
    if (!docType || !window.confirm(`Reset all saved values for ${docType.label}? Defaults will remain available.`)) {
      return;
    }

    try {
      setBusy(true);
      const resetRecord = await deleteSingleResource(docType.doctype);
      setRecord({ name: docType.doctype, ...(resetRecord || {}) });
      setExists(false);
      setError('');
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setBusy(false);
    }
  };

  if (loading) {
    return <section className="panel p-6 text-sm text-slate-300">Loading single DocType metadata…</section>;
  }

  if (!docType) {
    return <section className="panel p-6 text-sm text-slate-300">Choose a DocType to view its singleton settings desk.</section>;
  }

  return (
    <section className="space-y-4">
      <div className="panel p-5">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <span className="badge">Single DocType</span>
              <span className="badge">Settings desk</span>
            </div>
            <h2 className="mt-3 text-2xl font-semibold text-white">{docType.label} settings</h2>
            <p className="mt-2 max-w-3xl text-sm leading-6 text-slate-400">
              This DocType behaves like a singleton configuration document: one logical record, metadata-driven form rendering, and clean Go-backed persistence in the singles store.
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2 text-xs text-slate-300">
            <span className="badge">{exists ? 'Saved singleton' : 'Defaults only'}</span>
            <span className="badge">{detailFields.length} fields</span>
          </div>
        </div>
        {error ? <div className="mt-4 rounded-xl border border-rose-400/20 bg-rose-500/10 px-3 py-2 text-sm text-rose-100">{error}</div> : null}
      </div>

      <div className="grid gap-4 xl:grid-cols-[minmax(0,1.15fr)_360px]">
        <DynamicRecordForm
          docType={docType}
          initialValues={record || { name: docType.doctype }}
          onSubmit={handleSave}
          busy={busy || loadBusy}
          mode="edit"
          heading={`Edit ${docType.label} settings`}
          description="A singleton settings form rendered from metadata. Save updates without creating duplicate records."
          submitLabel={busy ? 'Saving…' : 'Save settings'}
        />

        <div className="space-y-4">
          <div className="panel p-5">
            <div className="text-xs font-semibold uppercase tracking-[0.24em] text-slate-500">Singleton state</div>
            <div className="mt-4 space-y-3 text-sm text-slate-300">
              <div className="flex items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
                <span>Logical name</span>
                <span className="font-mono text-cyan-200">{docType.doctype}</span>
              </div>
              <div className="flex items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
                <span>Saved state</span>
                <span>{exists ? 'Persisted' : 'Default-only'}</span>
              </div>
              <div className="flex items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
                <span>Created at</span>
                <span>{record?.created_at ? new Date(record.created_at).toLocaleString() : 'Not saved yet'}</span>
              </div>
              <div className="flex items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
                <span>Updated at</span>
                <span>{record?.updated_at ? new Date(record.updated_at).toLocaleString() : 'Not saved yet'}</span>
              </div>
            </div>
            <button
              type="button"
              onClick={handleReset}
              disabled={busy || loadBusy}
              className="mt-4 w-full rounded-xl border border-rose-400/20 bg-rose-500/10 px-4 py-2.5 text-sm font-semibold text-rose-100 transition hover:bg-rose-500/20 disabled:cursor-not-allowed disabled:opacity-60"
            >
              Reset saved values
            </button>
          </div>

          <div className="panel p-5">
            <div className="text-xs font-semibold uppercase tracking-[0.24em] text-slate-500">Current preview</div>
            <div className="mt-4 space-y-3">
              {detailFields.map((field) => (
                <div key={field.fieldname} className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
                  <div className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">{field.label}</div>
                  <div className="mt-1 text-[11px] uppercase tracking-[0.18em] text-slate-600">{field.fieldtype}</div>
                  <div className="mt-3 text-sm text-slate-100">
                    <RecordFieldValue field={field} value={record?.[field.fieldname]} />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
