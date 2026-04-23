import { useMemo, useState } from 'react';
import { createDocType } from '../lib/api.js';

const fieldTypeOptions = [
  'Data',
  'Text',
  'Small Text',
  'Long Text',
  'Check',
  'Int',
  'Float',
  'Currency',
  'Percent',
  'Date',
  'Datetime',
  'Time',
  'JSON',
  'Select',
  'Link',
  'DynamicLink',
];

function slugifyFieldName(value) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '_')
    .replace(/^_+|_+$/g, '');
}

function createEmptyField(index = 0) {
  return {
    id: `field-${Date.now()}-${index}`,
    label: '',
    fieldname: '',
    fieldtype: 'Data',
    reqd: false,
    unique: false,
  };
}

export default function DocTypeBuilder({ open, onClose, onCreated }) {
  const [doctype, setDocType] = useState('');
  const [label, setLabel] = useState('');
  const [moduleName, setModuleName] = useState('Core');
  const [description, setDescription] = useState('');
  const [fields, setFields] = useState([createEmptyField(1), createEmptyField(2)]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  const payloadPreview = useMemo(
    () => ({
      doctype,
      label,
      module: moduleName,
      description,
      fields: fields.map(({ id, ...rest }) => rest),
    }),
    [description, doctype, fields, label, moduleName],
  );

  if (!open) {
    return null;
  }

  const updateField = (fieldId, updater) => {
    setFields((current) =>
      current.map((field) => (field.id === fieldId ? { ...field, ...updater(field) } : field)),
    );
  };

  const addField = () => {
    setFields((current) => [...current, createEmptyField(current.length + 1)]);
  };

  const removeField = (fieldId) => {
    setFields((current) => (current.length > 1 ? current.filter((field) => field.id !== fieldId) : current));
  };

  const resetForm = () => {
    setDocType('');
    setLabel('');
    setModuleName('Core');
    setDescription('');
    setFields([createEmptyField(1), createEmptyField(2)]);
    setError('');
  };

  const handleClose = () => {
    if (busy) {
      return;
    }
    resetForm();
    onClose();
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');
    setBusy(true);

    try {
      const filteredFields = fields
        .map(({ id, ...field }) => field)
        .filter((field) => field.label.trim() || field.fieldname.trim());

      const created = await createDocType({
        doctype,
        label,
        module: moduleName,
        description,
        fields: filteredFields,
      });

      const createdName = created?.doctype || doctype;
      resetForm();
      onClose();
      onCreated(createdName);
    } catch (submitError) {
      setError(submitError.message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center bg-slate-950/80 px-4 py-8 backdrop-blur-sm">
      <div className="panel max-h-[92vh] w-full max-w-5xl overflow-hidden">
        <div className="flex items-start justify-between gap-4 border-b border-white/10 p-6">
          <div>
            <div className="mb-2 flex flex-wrap gap-2">
              <span className="badge">Low-code builder</span>
              <span className="badge">POST /api/doctypes</span>
            </div>
            <h2 className="text-2xl font-semibold text-white">Create a new DocType</h2>
            <p className="mt-2 text-sm text-slate-400">
              Define a model visually and let the Go backend generate metadata plus the PostgreSQL table.
            </p>
          </div>
          <button
            type="button"
            onClick={handleClose}
            className="rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-slate-300 transition hover:bg-white/10"
          >
            Close
          </button>
        </div>

        <div className="grid max-h-[calc(92vh-105px)] gap-0 overflow-hidden lg:grid-cols-[minmax(0,1.35fr)_420px]">
          <form className="overflow-y-auto p-6" onSubmit={handleSubmit}>
            {error ? <div className="mb-4 rounded-xl border border-rose-400/20 bg-rose-500/10 px-4 py-3 text-sm text-rose-100">{error}</div> : null}

            <div className="grid gap-4 md:grid-cols-2">
              <label className="grid gap-2">
                <span className="text-sm font-medium text-slate-200">DocType name</span>
                <input className="field" value={doctype} onChange={(event) => setDocType(event.target.value)} placeholder="CustomerInvoice" required />
              </label>
              <label className="grid gap-2">
                <span className="text-sm font-medium text-slate-200">Label</span>
                <input className="field" value={label} onChange={(event) => setLabel(event.target.value)} placeholder="Customer Invoice" required />
              </label>
              <label className="grid gap-2">
                <span className="text-sm font-medium text-slate-200">Module</span>
                <input className="field" value={moduleName} onChange={(event) => setModuleName(event.target.value)} placeholder="Core" />
              </label>
              <label className="grid gap-2 md:col-span-2">
                <span className="text-sm font-medium text-slate-200">Description</span>
                <textarea className="field min-h-24" value={description} onChange={(event) => setDescription(event.target.value)} placeholder="What is this DocType for?" />
              </label>
            </div>

            <div className="mt-6 flex items-center justify-between gap-4">
              <div>
                <h3 className="text-lg font-semibold text-white">Fields</h3>
                <p className="mt-1 text-sm text-slate-400">Define the data contract for this document type.</p>
              </div>
              <button
                type="button"
                onClick={addField}
                className="rounded-xl border border-cyan-400/30 bg-cyan-500/10 px-3 py-2 text-sm font-semibold text-cyan-200 transition hover:bg-cyan-500/20"
              >
                Add field
              </button>
            </div>

            <div className="mt-4 space-y-3">
              {fields.map((field, index) => (
                <div key={field.id} className="rounded-2xl border border-white/10 bg-white/5 p-4">
                  <div className="mb-3 flex items-center justify-between gap-3">
                    <div className="text-sm font-semibold text-white">Field {index + 1}</div>
                    <button
                      type="button"
                      onClick={() => removeField(field.id)}
                      className="rounded-lg border border-white/10 bg-white/5 px-2.5 py-1.5 text-xs text-slate-300 transition hover:bg-white/10"
                    >
                      Remove
                    </button>
                  </div>

                  <div className="grid gap-3 md:grid-cols-[1.2fr_1.2fr_1fr_auto_auto]">
                    <label className="grid gap-2">
                      <span className="text-xs font-semibold uppercase tracking-[0.15em] text-slate-400">Label</span>
                      <input
                        className="field"
                        value={field.label}
                        onChange={(event) =>
                          updateField(field.id, (current) => ({
                            label: event.target.value,
                            fieldname: current.fieldname || slugifyFieldName(event.target.value),
                          }))
                        }
                        placeholder="Customer Name"
                      />
                    </label>
                    <label className="grid gap-2">
                      <span className="text-xs font-semibold uppercase tracking-[0.15em] text-slate-400">Fieldname</span>
                      <input
                        className="field font-mono"
                        value={field.fieldname}
                        onChange={(event) => updateField(field.id, () => ({ fieldname: slugifyFieldName(event.target.value) }))}
                        placeholder="customer_name"
                      />
                    </label>
                    <label className="grid gap-2">
                      <span className="text-xs font-semibold uppercase tracking-[0.15em] text-slate-400">Type</span>
                      <select className="field" value={field.fieldtype} onChange={(event) => updateField(field.id, () => ({ fieldtype: event.target.value }))}>
                        {fieldTypeOptions.map((option) => (
                          <option key={option} value={option}>{option}</option>
                        ))}
                      </select>
                    </label>
                    <label className="flex items-end gap-2 pb-2 text-sm text-slate-300">
                      <input type="checkbox" checked={field.reqd} onChange={(event) => updateField(field.id, () => ({ reqd: event.target.checked }))} />
                      Required
                    </label>
                    <label className="flex items-end gap-2 pb-2 text-sm text-slate-300">
                      <input type="checkbox" checked={field.unique} onChange={(event) => updateField(field.id, () => ({ unique: event.target.checked }))} />
                      Unique
                    </label>
                  </div>
                </div>
              ))}
            </div>

            <div className="mt-6 flex justify-end gap-3">
              <button
                type="button"
                onClick={handleClose}
                className="rounded-xl border border-white/10 bg-white/5 px-4 py-2.5 text-sm font-medium text-slate-300 transition hover:bg-white/10"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={busy}
                className="rounded-xl bg-cyan-500 px-4 py-2.5 text-sm font-semibold text-slate-950 transition hover:bg-cyan-400 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {busy ? 'Creating…' : 'Create DocType'}
              </button>
            </div>
          </form>

          <div className="border-l border-white/10 bg-slate-950/70 p-6">
            <h3 className="text-lg font-semibold text-white">Payload preview</h3>
            <p className="mt-2 text-sm text-slate-400">This is the JSON sent to the Go API. Handy when you want low-code and pro-code to shake hands politely.</p>
            <pre className="mt-4 overflow-auto rounded-2xl border border-white/10 bg-slate-950 p-4 text-xs leading-6 text-cyan-200">
{JSON.stringify(payloadPreview, null, 2)}
            </pre>
          </div>
        </div>
      </div>
    </div>
  );
}