import { useEffect, useMemo, useState } from 'react';
import ChildTableField from './ChildTableField.jsx';
import LinkFieldInput from './LinkFieldInput.jsx';
import {
  buildInitialValues,
  fieldInputType,
  getSelectChoices,
  isAttachmentField,
  isChildTableField,
  isEditableField,
  isImageField,
  normalizeSubmissionValue,
  resolveFileURL,
} from '../lib/metadata.js';

export default function DynamicRecordForm({
	docType,
	onSubmit,
	busy,
	mode = 'create',
	initialValues = null,
	onCancel,
	submitLabel,
	heading,
	description,
}) {
  const editableFields = useMemo(
    () => (docType?.fields || []).filter((field) => isEditableField(field) && !field.read_only),
    [docType],
  );

  const [values, setValues] = useState(() => buildInitialValues(docType, initialValues || {}));
  const [recordName, setRecordName] = useState(initialValues?.name || '');
  const [error, setError] = useState('');

	useEffect(() => {
		setValues(buildInitialValues(docType, initialValues || {}));
		setRecordName(initialValues?.name || '');
		setError('');
	}, [docType, initialValues]);

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
      if (mode === 'create' && recordName.trim()) {
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
		if (mode === 'create') {
			setRecordName('');
			setValues(buildInitialValues(docType, {}));
		}
    } catch (submitError) {
      setError(submitError.message);
    }
  };

	const resolvedHeading = heading || (mode === 'edit' ? `Edit ${docType.label}` : `Create ${docType.label}`);
	const resolvedDescription = description || (mode === 'edit'
		? 'Update the selected record through the same metadata-driven form renderer.'
		: 'Metadata-driven form rendered from the selected DocType.');
	const resolvedSubmitLabel = submitLabel || (mode === 'edit' ? 'Save changes' : 'Create record');

  const spansFullWidth = (field) => (
    field.fieldtype === 'Text'
    || field.fieldtype === 'Small Text'
    || field.fieldtype === 'Long Text'
    || field.fieldtype === 'JSON'
    || field.fieldtype === 'Link'
    || field.fieldtype === 'DynamicLink'
    || isChildTableField(field)
    || isAttachmentField(field)
  );

  return (
    <form className="panel p-5" onSubmit={handleSubmit}>
      <div className="mb-4 flex items-start justify-between gap-3">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <span className="badge">{docType.module || 'Core'} module</span>
            <span className="badge">{mode === 'edit' ? 'Form view' : 'Create view'}</span>
          </div>
          <h3 className="text-lg font-semibold text-white">{resolvedHeading}</h3>
          <p className="mt-1 text-sm text-slate-400">{resolvedDescription}</p>
        </div>
        <span className="badge">{mode === 'edit' ? `PUT /api/resource/${docType.doctype}/:name` : `POST /api/resource/${docType.doctype}`}</span>
      </div>

      {error ? <div className="mb-4 rounded-xl border border-rose-400/20 bg-rose-500/10 px-3 py-2 text-sm text-rose-100">{error}</div> : null}

      <div className="mb-4 rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3 text-sm text-slate-300">
        <div className="grid gap-3 md:grid-cols-3">
          <div>
            <div className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500">DocType</div>
            <div className="mt-2 font-medium text-white">{docType.label}</div>
          </div>
          <div>
            <div className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500">Module</div>
            <div className="mt-2 font-medium text-white">{docType.module || 'Core'}</div>
          </div>
          <div>
            <div className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500">Editable fields</div>
            <div className="mt-2 font-medium text-white">{editableFields.length}</div>
          </div>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        {mode === 'create' ? (
      <label className="grid gap-2 md:col-span-2">
        <span className="text-sm font-medium text-slate-200">Name (optional)</span>
        <input className="field" value={recordName} onChange={(event) => setRecordName(event.target.value)} placeholder="Leave blank to auto-generate" />
      </label>
    ) : (
      <div className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3 text-sm text-slate-300 md:col-span-2">
        <div className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Record</div>
        <div className="mt-2 font-mono text-cyan-200">{recordName || initialValues?.name || 'Unknown record'}</div>
      </div>
    )}

        {editableFields.map((field) => (
          <label key={field.fieldname} className={`grid gap-2 ${spansFullWidth(field) ? 'md:col-span-2' : ''}`}>
            <span className="flex items-center gap-2 text-sm font-medium text-slate-200">
              <span>{field.label}</span>
              <span className="rounded-full border border-white/10 px-2 py-0.5 text-[10px] uppercase tracking-[0.18em] text-slate-400">{field.fieldtype}</span>
              {field.reqd ? <span className="text-amber-300">*</span> : null}
            </span>
            {field.description ? <span className="text-xs text-slate-500">{field.description}</span> : null}

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
            ) : field.fieldtype === 'Link' || field.fieldtype === 'DynamicLink' ? (
        <LinkFieldInput
          field={field}
          value={values[field.fieldname]}
          onChange={(nextValue) => handleChange(field.fieldname, nextValue)}
          placeholder={`Search ${field.label.toLowerCase()}`}
        />
      ) : isChildTableField(field) ? (
        <ChildTableField
          field={field}
          value={values[field.fieldname]}
          onChange={(nextValue) => handleChange(field.fieldname, nextValue)}
        />
      ) : isAttachmentField(field) ? (
      <div className="grid gap-3">
        <input
        className="field"
        value={values[field.fieldname]}
        onChange={(event) => handleChange(field.fieldname, event.target.value)}
        placeholder={isImageField(field) ? 'Paste an image URL or /files/... path' : 'Paste a file URL or /files/... path'}
        />
        {values[field.fieldname] ? (
        isImageField(field) ? (
          <div className="rounded-2xl border border-white/10 bg-white/[0.03] p-3">
          <img
            src={resolveFileURL(values[field.fieldname])}
            alt={field.label}
            className="h-32 w-32 rounded-2xl object-cover ring-1 ring-white/10"
          />
          </div>
        ) : (
          <a
          href={resolveFileURL(values[field.fieldname])}
          target="_blank"
          rel="noreferrer"
          className="text-sm text-cyan-200 underline decoration-cyan-400/40 underline-offset-4"
          >
          Open attached file
          </a>
        )
        ) : null}
      </div>
      ) : field.fieldtype === 'Select' && getSelectChoices(field).length > 0 ? (
        <select
          className="field"
          value={values[field.fieldname]}
          onChange={(event) => handleChange(field.fieldname, event.target.value)}
        >
          <option value="">Select {field.label}</option>
          {getSelectChoices(field).map((option) => (
            <option key={option} value={option}>{option}</option>
          ))}
        </select>
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

		<div className="mt-5 flex justify-end gap-3">
		  {onCancel ? (
				<button
				  type="button"
				  onClick={onCancel}
				  className="rounded-xl border border-white/10 bg-white/[0.05] px-4 py-2.5 text-sm font-medium text-slate-300 transition hover:bg-white/[0.1]"
				>
				  Cancel
				</button>
		  ) : null}
        <button
          type="submit"
          disabled={busy}
          className="rounded-xl bg-cyan-500 px-4 py-2.5 text-sm font-semibold text-slate-950 transition hover:bg-cyan-400 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {busy ? 'Saving…' : resolvedSubmitLabel}
        </button>
      </div>
    </form>
  );
}
