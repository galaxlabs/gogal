import { useMemo, useState } from 'react';

const supportedTypes = new Set(['Data', 'Text', 'Small Text', 'Long Text', 'Check', 'Int', 'Float', 'Currency', 'Percent', 'Date', 'Datetime', 'Time', 'JSON', 'Select', 'Link', 'DynamicLink']);

function defaultValueForField(field) {
  if (field.fieldtype === 'Check') {
    return false;
  }
  return '';
}

function fieldInputType(field) {
  switch (field.fieldtype) {
    case 'Date':
      return 'date';
    case 'Datetime':
      return 'datetime-local';
    case 'Time':
      return 'time';
    case 'Int':
    case 'Float':
    case 'Currency':
    case 'Percent':
      return 'number';
    default:
      return 'text';
  }
}

function normalizeSubmissionValue(field, value) {
  if (field.fieldtype === 'Check') {
    return Boolean(value);
  }

  if (value === '') {
    return '';
  }

  if (field.fieldtype === 'Int') {
    return Number.parseInt(value, 10);
  }

  if (field.fieldtype === 'Float' || field.fieldtype === 'Currency' || field.fieldtype === 'Percent') {
    return Number.parseFloat(value);
  }

  if (field.fieldtype === 'JSON') {
    return value.trim() ? JSON.parse(value) : {};
  }

  if (field.fieldtype === 'Datetime' && value) {
    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? value : date.toISOString();
  }

  return value;
}

export default function DynamicRecordForm({ docType, onSubmit, busy }) {
  const editableFields = useMemo(
    () => (docType?.fields || []).filter((field) => supportedTypes.has(field.fieldtype) && !field.read_only && !field.hidden),
    [docType],
  );

  const [values, setValues] = useState(() => Object.fromEntries(editableFields.map((field) => [field.fieldname, defaultValueForField(field)])));
  const [recordName, setRecordName] = useState('');
  const [error, setError] = useState('');

  if (!docType) {
    return null;
  }

  const handleChange = (fieldName, nextValue) => {
    setValues((current) => ({
      ...current,
      [fieldName]: nextValue,
    }));
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');

    try {
      const payload = {};
      if (recordName.trim()) {
        payload.name = recordName.trim();
      }

      editableFields.forEach((field) => {
        const rawValue = values[field.fieldname];
        if (rawValue === '' && !field.reqd && field.fieldtype !== 'Check') {
          return;
        }
        payload[field.fieldname] = normalizeSubmissionValue(field, rawValue);
      });

      await onSubmit(payload);
      setRecordName('');
      setValues(Object.fromEntries(editableFields.map((field) => [field.fieldname, defaultValueForField(field)])));
    } catch (submitError) {
      setError(submitError.message);
    }
  };

  return (
    <form className="panel p-5" onSubmit={handleSubmit}>
      <div className="mb-4 flex items-start justify-between gap-3">
        <div>
          <h3 className="text-lg font-semibold text-white">Create {docType.label}</h3>
          <p className="mt-1 text-sm text-slate-400">Metadata-driven form rendered from the selected DocType.</p>
        </div>
        <span className="badge">POST /api/resource/{docType.doctype}</span>
      </div>

      {error ? <div className="mb-4 rounded-xl border border-rose-400/20 bg-rose-500/10 px-3 py-2 text-sm text-rose-100">{error}</div> : null}

      <div className="grid gap-4">
        <label className="grid gap-2">
          <span className="text-sm font-medium text-slate-200">Name (optional)</span>
          <input className="field" value={recordName} onChange={(event) => setRecordName(event.target.value)} placeholder="Leave blank to auto-generate" />
        </label>

        {editableFields.map((field) => (
          <label key={field.fieldname} className="grid gap-2">
            <span className="flex items-center gap-2 text-sm font-medium text-slate-200">
              {field.label}
              <span className="text-xs text-slate-500">{field.fieldtype}</span>
              {field.reqd ? <span className="text-amber-300">*</span> : null}
            </span>

            {field.fieldtype === 'Check' ? (
              <button
                type="button"
                onClick={() => handleChange(field.fieldname, !values[field.fieldname])}
                className={`flex items-center justify-between rounded-xl border px-3 py-2.5 text-sm transition ${
                  values[field.fieldname]
                    ? 'border-emerald-400/30 bg-emerald-500/10 text-emerald-100'
                    : 'border-white/10 bg-slate-950/80 text-slate-300'
                }`}
              >
                <span>{values[field.fieldname] ? 'Enabled' : 'Disabled'}</span>
                <span>{values[field.fieldname] ? '✓' : '○'}</span>
              </button>
            ) : field.fieldtype === 'Text' || field.fieldtype === 'Small Text' || field.fieldtype === 'Long Text' || field.fieldtype === 'JSON' ? (
              <textarea
                className="field min-h-28"
                value={values[field.fieldname]}
                onChange={(event) => handleChange(field.fieldname, event.target.value)}
                placeholder={field.fieldtype === 'JSON' ? '{"hello":"world"}' : `Enter ${field.label.toLowerCase()}`}
              />
            ) : (
              <input
                className="field"
                type={fieldInputType(field)}
                step={field.fieldtype === 'Int' ? '1' : 'any'}
                value={values[field.fieldname]}
                onChange={(event) => handleChange(field.fieldname, event.target.value)}
                placeholder={`Enter ${field.label.toLowerCase()}`}
              />
            )}
          </label>
        ))}
      </div>

      <div className="mt-5 flex justify-end">
        <button
          type="submit"
          disabled={busy}
          className="rounded-xl bg-cyan-500 px-4 py-2.5 text-sm font-semibold text-slate-950 transition hover:bg-cyan-400 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {busy ? 'Saving…' : 'Create record'}
        </button>
      </div>
    </form>
  );
}
