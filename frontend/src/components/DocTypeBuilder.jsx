import { useEffect, useMemo, useState } from 'react';
import { childTableOptionsTemplate, isImageField } from '../lib/metadata.js';
import { createDocType } from '../lib/api.js';

const fieldTypeOptions = ['Data', 'Text', 'Check', 'Int', 'Float', 'Select', 'Link', 'Attach', 'Attach Image', 'Image', 'JSON'];

const fieldPalette = [
  {
    id: 'palette-data',
    title: 'Data',
    fieldtype: 'Data',
    label: 'New Data Field',
    description: 'Single-line text for names, titles, references, or search keys.',
    accent: 'cyan',
  },
  {
    id: 'palette-text',
    title: 'Text',
    fieldtype: 'Text',
    label: 'New Text Field',
    description: 'Longer notes, descriptions, and free-form content.',
    accent: 'violet',
  },
  {
    id: 'palette-check',
    title: 'Check',
    fieldtype: 'Check',
    label: 'New Check Field',
    description: 'Boolean toggle for flags like Active, Approved, or Is Paid.',
    accent: 'emerald',
  },
  {
    id: 'palette-int',
    title: 'Int',
    fieldtype: 'Int',
    label: 'New Integer Field',
    description: 'Whole-number values for quantities, priorities, or counters.',
    accent: 'violet',
  },
  {
    id: 'palette-link',
    title: 'Link',
    fieldtype: 'Link',
    label: 'Linked Record',
    description: 'Lookup another DocType and store a live relationship.',
    accent: 'cyan',
    options: 'doctype_name',
  },
  {
    id: 'palette-attach',
    title: 'Attach',
    fieldtype: 'Attach',
    label: 'Attachment',
    description: 'Store a file URL or uploaded file reference.',
    accent: 'violet',
  },
  {
    id: 'palette-attach-image',
    title: 'Attach Image',
    fieldtype: 'Attach Image',
    label: 'Profile Image',
    description: 'Store an image path for previews, avatars, or logos.',
    accent: 'emerald',
  },
  {
    id: 'palette-child-table',
    title: 'Child Table',
    fieldtype: 'JSON',
    label: 'Line Items',
    description: 'Repeated nested rows using the current child-table JSON contract.',
    accent: 'emerald',
    options: childTableOptionsTemplate(),
  },
];

function slugifyFieldName(value) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '_')
    .replace(/^_+|_+$/g, '');
}

function createFieldFromTemplate(template, index = 0) {
  return {
    id: `field-${Date.now()}-${index}`,
    label: template.label,
    fieldname: slugifyFieldName(`${template.label}_${index + 1}`),
    fieldtype: template.fieldtype,
    reqd: false,
    read_only: false,
    hidden: false,
    unique: false,
    default: '',
    options: template.options || '',
    in_list_view: true,
  };
}

function BaseInput({ label, hint, ...props }) {
  return (
    <label className="grid gap-2">
      {label ? <span className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">{label}</span> : null}
      <input className="field" {...props} />
      {hint ? <span className="text-xs text-slate-500">{hint}</span> : null}
    </label>
  );
}

function BaseTextArea({ label, hint, ...props }) {
  return (
    <label className="grid gap-2">
      {label ? <span className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">{label}</span> : null}
      <textarea className="field min-h-24" {...props} />
      {hint ? <span className="text-xs text-slate-500">{hint}</span> : null}
    </label>
  );
}

function BaseSelect({ label, children, ...props }) {
  return (
    <label className="grid gap-2">
      {label ? <span className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">{label}</span> : null}
      <select className="field" {...props}>
        {children}
      </select>
    </label>
  );
}

function ToggleCard({ label, description, checked, onChange, disabled = false }) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={() => onChange(!checked)}
      className={`flex items-start justify-between gap-4 rounded-2xl border px-4 py-3 text-left transition ${
        disabled
          ? 'cursor-not-allowed border-white/10 bg-white/[0.02] opacity-60'
          : checked
            ? 'border-cyan-400/30 bg-cyan-500/10'
            : 'border-white/10 bg-white/[0.03] hover:bg-white/[0.05]'
      }`}
    >
      <div>
        <div className="text-sm font-semibold text-white">{label}</div>
        <div className="mt-1 text-xs leading-5 text-slate-500">{description}</div>
      </div>
      <span className={`mt-0.5 inline-flex h-6 w-11 rounded-full p-1 transition ${checked ? 'bg-cyan-400/80' : 'bg-slate-700'}`}>
        <span className={`h-4 w-4 rounded-full bg-white transition ${checked ? 'translate-x-5' : ''}`} />
      </span>
    </button>
  );
}

function TemplateButton({ template, onAdd }) {
  const accentClasses = {
    cyan: 'border-cyan-400/20 bg-cyan-500/10 text-cyan-100',
    emerald: 'border-emerald-400/20 bg-emerald-500/10 text-emerald-100',
    violet: 'border-violet-400/20 bg-violet-500/10 text-violet-100',
  };

  return (
    <button
      type="button"
      onClick={onAdd}
      className="rounded-2xl border border-white/10 bg-white/[0.04] p-4 text-left transition hover:-translate-y-0.5 hover:border-white/15 hover:bg-white/[0.06]"
    >
      <div className="flex items-start justify-between gap-3">
        <div>
          <div className="text-sm font-semibold text-white">{template.title}</div>
          <p className="mt-2 text-xs leading-5 text-slate-400">{template.description}</p>
        </div>
        <span className={`rounded-full border px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.2em] ${accentClasses[template.accent]}`}>
          {template.fieldtype}
        </span>
      </div>
      <div className="mt-4 text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Add to canvas</div>
    </button>
  );
}

function FieldRow({ field, index, total, active, onSelect, onMoveUp, onMoveDown, onRemove }) {
  return (
    <div
      className={`rounded-3xl border p-4 transition ${
        active ? 'border-cyan-400/30 bg-cyan-500/10 shadow-lg shadow-cyan-500/10' : 'border-white/10 bg-white/[0.04]'
      }`}
    >
      <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <button type="button" onClick={onSelect} className="flex-1 text-left">
          <div className="flex flex-wrap items-center gap-2">
            <span className="rounded-full border border-white/10 px-2 py-1 text-[11px] font-semibold text-slate-300">#{index + 1}</span>
            <div className="text-sm font-semibold text-white">{field.label || 'Untitled field'}</div>
            <span className="text-xs text-slate-500">{field.fieldtype}</span>
          </div>
          <div className="mt-2 flex flex-wrap items-center gap-2 text-xs text-slate-500">
            <span>{field.fieldname || 'field_name'}</span>
            {field.reqd ? <span>• Mandatory</span> : null}
            {field.unique ? <span>• Unique</span> : null}
            {field.hidden ? <span>• Hidden</span> : null}
          </div>
        </button>
        <div className="flex flex-wrap items-center gap-2">
          <button type="button" onClick={onSelect} className="rounded-full border border-white/10 bg-white/[0.05] px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-white/[0.1]">Edit</button>
          <button type="button" onClick={onMoveUp} disabled={index === 0} className="rounded-full border border-white/10 bg-white/[0.05] px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-white/[0.1] disabled:cursor-not-allowed disabled:opacity-40">↑</button>
          <button type="button" onClick={onMoveDown} disabled={index === total - 1} className="rounded-full border border-white/10 bg-white/[0.05] px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-white/[0.1] disabled:cursor-not-allowed disabled:opacity-40">↓</button>
          <button type="button" onClick={onRemove} className="rounded-full border border-rose-400/20 bg-rose-500/10 px-3 py-1.5 text-xs font-semibold text-rose-100 transition hover:bg-rose-500/20">Remove</button>
        </div>
      </div>
    </div>
  );
}

function normalizeFieldPayload(field, index) {
  return {
    label: field.label.trim(),
    fieldname: slugifyFieldName(field.fieldname || field.label),
    fieldtype: field.fieldtype,
    reqd: Boolean(field.reqd),
    read_only: Boolean(field.read_only),
    unique: Boolean(field.unique),
    hidden: Boolean(field.hidden),
    default: String(field.default || '').trim(),
    options: String(field.options || '').trim(),
    in_list_view: Boolean(field.in_list_view),
    sort_order: index + 1,
  };
}

export default function DocTypeBuilder({ moduleOptions = [], existingDocTypes = [], onCreated }) {
  const [doctype, setDocType] = useState('');
  const [label, setLabel] = useState('');
  const [moduleName, setModuleName] = useState('Core');
  const [description, setDescription] = useState('');
  const [isSingle, setIsSingle] = useState(false);
  const [isChildTable, setIsChildTable] = useState(false);
  const [allowRename, setAllowRename] = useState(true);
  const [quickEntry, setQuickEntry] = useState(false);
  const [maxAttachments, setMaxAttachments] = useState('');
  const [imageField, setImageField] = useState('');
  const [fields, setFields] = useState([]);
  const [selectedFieldId, setSelectedFieldId] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [quickAddTemplateId, setQuickAddTemplateId] = useState(fieldPalette[0].id);

  const availableModules = useMemo(() => ['Core', ...moduleOptions.filter((item) => item && item !== 'Core')], [moduleOptions]);
  const availableImageFields = useMemo(() => fields.filter((field) => isImageField(field)), [fields]);

  const selectedField = useMemo(
    () => fields.find((field) => field.id === selectedFieldId) || fields[0] || null,
    [fields, selectedFieldId],
  );

  const payloadPreview = useMemo(
    () => ({
      doctype,
      label: label || doctype,
      module: moduleName,
      description,
      is_single: isSingle,
      is_child_table: isChildTable,
      allow_rename: allowRename,
      quick_entry: isSingle || isChildTable ? false : quickEntry,
      max_attachments: maxAttachments === '' ? 0 : Number(maxAttachments),
      image_field: imageField || undefined,
      fields: fields.map((field, index) => normalizeFieldPayload(field, index)),
    }),
    [allowRename, description, doctype, fields, imageField, isChildTable, isSingle, label, maxAttachments, moduleName, quickEntry],
  );

  useEffect(() => {
    if (!imageField) {
      return;
    }

    const stillExists = availableImageFields.some((field) => field.fieldname === imageField);
    if (!stillExists) {
      setImageField('');
    }
  }, [availableImageFields, imageField]);

  const updateField = (fieldId, updater) => {
    setFields((current) => current.map((field) => (field.id === fieldId ? { ...field, ...updater(field) } : field)));
  };

  const addFieldFromTemplate = (templateId) => {
    const template = fieldPalette.find((item) => item.id === templateId);
    if (!template) {
      return;
    }

    const created = createFieldFromTemplate(template, fields.length + 1);
    setFields((current) => [...current, created]);
    setSelectedFieldId(created.id);
  };

  const moveField = (fieldId, direction) => {
    setFields((current) => {
      const index = current.findIndex((field) => field.id === fieldId);
      if (index === -1) {
        return current;
      }

      const targetIndex = index + direction;
      if (targetIndex < 0 || targetIndex >= current.length) {
        return current;
      }

      const next = [...current];
      const [moved] = next.splice(index, 1);
      next.splice(targetIndex, 0, moved);
      return next;
    });
  };

  const removeField = (fieldId) => {
    setFields((current) => current.filter((field) => field.id !== fieldId));
    setSelectedFieldId((current) => (current === fieldId ? '' : current));
  };

  const resetForm = ({ keepSuccess = false } = {}) => {
    setDocType('');
    setLabel('');
    setModuleName('Core');
    setDescription('');
    setIsSingle(false);
    setIsChildTable(false);
    setAllowRename(true);
    setQuickEntry(false);
    setMaxAttachments('');
    setImageField('');
    setFields([]);
    setSelectedFieldId('');
    setError('');
    if (!keepSuccess) {
      setSuccess('');
    }
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');
    setSuccess('');

    const normalizedDocType = doctype.trim();
    const existingMatch = existingDocTypes.some((item) => item.toLowerCase() === normalizedDocType.toLowerCase());
    if (existingMatch) {
      setError(`DocType ${normalizedDocType} already exists.`);
      return;
    }

    if (!normalizedDocType) {
      setError('DocType name is required.');
      return;
    }

    if (fields.length === 0) {
      setError('Add at least one field from the field library or the canvas add control.');
      return;
    }

    const filteredFields = fields.map((field, index) => normalizeFieldPayload(field, index));
    const invalidField = filteredFields.find((field) => !field.label || !field.fieldname);
    if (invalidField) {
      setError('Every field needs both a label and a valid fieldname before saving.');
      return;
    }

    setBusy(true);

    try {
      const payload = {
        doctype: normalizedDocType,
        label: label.trim() || normalizedDocType,
        module: moduleName,
        description,
        is_single: isSingle,
        is_child_table: isChildTable,
        allow_rename: allowRename,
        quick_entry: isSingle || isChildTable ? false : quickEntry,
        max_attachments: maxAttachments === '' ? 0 : Number(maxAttachments),
        image_field: imageField || '',
        fields: filteredFields,
      };

      const created = await createDocType(payload);
      const createdName = created?.doctype || normalizedDocType;
      onCreated?.(created || payload);
      resetForm({ keepSuccess: true });
      setSuccess(`DocType ${createdName} created successfully.`);
    } catch (submitError) {
      setError(submitError.message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <section className="panel mx-auto w-full max-w-6xl overflow-hidden">
      <div className="border-b border-white/10 px-5 py-5 lg:px-6">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <div className="mb-2 flex flex-wrap gap-2">
              <span className="badge">UI Studio</span>
              <span className="badge">Plain canvas</span>
              <span className="badge">POST /api/doctypes</span>
            </div>
            <h2 className="text-2xl font-semibold text-white">Simple DocType Builder</h2>
            <p className="mt-2 max-w-3xl text-sm leading-6 text-slate-400">
              A simpler flow: choose a field from the library or use the canvas add control, then edit its properties below. No hunting across a giant layout.
            </p>
          </div>
          <div className="rounded-3xl border border-cyan-400/15 bg-cyan-500/10 px-4 py-3 text-sm text-cyan-100">
            <div className="font-semibold">Reliable field creation</div>
            <div className="mt-1 text-cyan-100/75">Works with add buttons even when drag-and-drop is not used.</div>
          </div>
        </div>
      </div>

      <form className="grid gap-6 p-5 lg:grid-cols-[280px_minmax(0,1fr)] lg:p-6" onSubmit={handleSubmit}>
        <div className="space-y-6">
          <section className="rounded-3xl border border-white/10 bg-white/[0.03] p-4">
            <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">DocType setup</div>
            <div className="mt-4 grid gap-4">
              <BaseInput label="DocType name" value={doctype} onChange={(event) => setDocType(event.target.value)} placeholder="CustomerInvoice" hint="Used as the API-facing DocType identifier." />
              <BaseInput label="Label" value={label} onChange={(event) => setLabel(event.target.value)} placeholder="Customer Invoice" hint="Human-readable label for desk and forms." />
              <BaseSelect label="Module" value={moduleName} onChange={(event) => setModuleName(event.target.value)}>
                {availableModules.map((moduleOption) => (
                  <option key={moduleOption} value={moduleOption}>{moduleOption}</option>
                ))}
              </BaseSelect>
              <BaseTextArea label="Description" value={description} onChange={(event) => setDescription(event.target.value)} placeholder="What workflow or business object does this DocType represent?" />
            </div>
          </section>

          <section className="rounded-3xl border border-white/10 bg-white/[0.03] p-4">
            <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">Field library</div>
            <div className="mt-4 grid gap-3">
              {fieldPalette.map((template) => (
                <TemplateButton key={template.id} template={template} onAdd={() => addFieldFromTemplate(template.id)} />
              ))}
            </div>
          </section>
        </div>

        <div className="space-y-6">
          {error ? <div className="rounded-2xl border border-rose-400/20 bg-rose-500/10 px-4 py-3 text-sm text-rose-100">{error}</div> : null}
          {success ? <div className="rounded-2xl border border-emerald-400/20 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-100">{success}</div> : null}

          <section className="rounded-3xl border border-white/10 bg-white/[0.03] p-4 lg:p-5">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
              <div>
                <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">Canvas</div>
                <h3 className="mt-2 text-xl font-semibold text-white">Field sequence</h3>
                <p className="mt-2 max-w-2xl text-sm leading-6 text-slate-400">
                  Use the add control here or the field library on the left. Reorder with the up/down buttons for a predictable, plain-canvas workflow.
                </p>
              </div>
              <div className="rounded-2xl border border-white/10 bg-white/[0.03] px-3 py-2 text-right">
                <div className="text-xs uppercase tracking-[0.24em] text-slate-500">Fields</div>
                <div className="mt-1 text-lg font-semibold text-white">{fields.length}</div>
              </div>
            </div>

            <div className="mt-5 flex flex-col gap-3 rounded-3xl border border-cyan-400/15 bg-cyan-500/10 p-4 lg:flex-row lg:items-end">
              <div className="min-w-0 flex-1">
                <BaseSelect label="Add field to canvas" value={quickAddTemplateId} onChange={(event) => setQuickAddTemplateId(event.target.value)}>
                  {fieldPalette.map((template) => (
                    <option key={template.id} value={template.id}>{template.title} ({template.fieldtype})</option>
                  ))}
                </BaseSelect>
              </div>
              <button
                type="button"
                onClick={() => addFieldFromTemplate(quickAddTemplateId)}
                className="rounded-2xl bg-cyan-400 px-4 py-3 text-sm font-semibold text-slate-950 transition hover:bg-cyan-300 active:scale-95"
              >
                Add field
              </button>
            </div>

            <div className="mt-5 space-y-3">
              {fields.length === 0 ? (
                <div className="flex min-h-[220px] flex-col items-center justify-center rounded-[24px] border border-dashed border-white/10 bg-slate-950/40 px-6 py-10 text-center">
                  <div className="rounded-full border border-cyan-400/20 bg-cyan-500/10 px-4 py-1.5 text-xs font-semibold uppercase tracking-[0.24em] text-cyan-100">
                    Canvas ready
                  </div>
                  <h4 className="mt-4 text-xl font-semibold text-white">Add your first field</h4>
                  <p className="mt-3 max-w-xl text-sm leading-6 text-slate-400">
                    Choose a field from the left library or pick one from the add control above. This path works even if drag-and-drop is ignored completely.
                  </p>
                </div>
              ) : (
                fields.map((field, index) => (
                  <FieldRow
                    key={field.id}
                    field={field}
                    index={index}
                    total={fields.length}
                    active={selectedField?.id === field.id}
                    onSelect={() => setSelectedFieldId(field.id)}
                    onMoveUp={() => moveField(field.id, -1)}
                    onMoveDown={() => moveField(field.id, 1)}
                    onRemove={() => removeField(field.id)}
                  />
                ))
              )}
            </div>
          </section>

          <section className="rounded-3xl border border-white/10 bg-slate-950/40 p-4 lg:p-5">
            <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">DocType behavior</div>
            <div className="mt-4 grid gap-3 xl:grid-cols-2">
              <ToggleCard
                label="Is Single"
                description="Create a singleton settings-style DocType."
                checked={isSingle}
                onChange={(value) => {
                  setIsSingle(value);
                  if (value) {
                    setIsChildTable(false);
                    setQuickEntry(false);
                  }
                }}
              />
              <ToggleCard
                label="Is Child Table"
                description="Use this DocType as nested rows inside a parent document."
                checked={isChildTable}
                onChange={(value) => {
                  setIsChildTable(value);
                  if (value) {
                    setIsSingle(false);
                    setQuickEntry(false);
                  }
                }}
                disabled={isSingle}
              />
              <ToggleCard label="Allow Rename" description="Prepare this DocType for future rename support." checked={allowRename} onChange={setAllowRename} />
              <ToggleCard
                label="Quick Entry"
                description="Reserve a compact create flow for normal doctypes."
                checked={isSingle || isChildTable ? false : quickEntry}
                onChange={setQuickEntry}
                disabled={isSingle || isChildTable}
              />
            </div>
            <div className="mt-4 grid gap-4 xl:grid-cols-2">
              <BaseInput label="Max attachments" type="number" min="0" value={maxAttachments} onChange={(event) => setMaxAttachments(event.target.value)} placeholder="0" hint="Use 0 for no explicit limit." />
              <BaseSelect label="Image field" value={imageField} onChange={(event) => setImageField(event.target.value)}>
                <option value="">No profile/sidebar image</option>
                {availableImageFields.map((fieldOption) => (
                  <option key={fieldOption.id} value={fieldOption.fieldname}>{fieldOption.label} ({fieldOption.fieldname})</option>
                ))}
              </BaseSelect>
            </div>
          </section>

          <section className="rounded-3xl border border-white/10 bg-slate-950/40 p-4 lg:p-5">
            <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">Field properties</div>
            <h3 className="mt-2 text-xl font-semibold text-white">{selectedField ? selectedField.label || 'Selected field' : 'Select a field'}</h3>
            <p className="mt-2 text-sm leading-6 text-slate-400">
              Edit one field at a time directly under the canvas, so your eyes stay in one working area.
            </p>

            {selectedField ? (
              <div className="mt-5 grid gap-4 xl:grid-cols-2">
                <BaseInput
                  label="Field label"
                  value={selectedField.label}
                  onChange={(event) =>
                    updateField(selectedField.id, (current) => ({
                      label: event.target.value,
                      fieldname: current.fieldname || slugifyFieldName(event.target.value),
                    }))
                  }
                  placeholder="Customer Name"
                />
                <BaseInput
                  label="Fieldname"
                  value={selectedField.fieldname}
                  onChange={(event) => updateField(selectedField.id, () => ({ fieldname: slugifyFieldName(event.target.value) }))}
                  placeholder="customer_name"
                />
                <BaseSelect label="Field type" value={selectedField.fieldtype} onChange={(event) => updateField(selectedField.id, () => ({ fieldtype: event.target.value }))}>
                  {fieldTypeOptions.map((option) => (
                    <option key={option} value={option}>{option}</option>
                  ))}
                </BaseSelect>
                <BaseInput label="Default value" value={selectedField.default} onChange={(event) => updateField(selectedField.id, () => ({ default: event.target.value }))} placeholder={selectedField.fieldtype === 'Check' ? '0 or 1' : 'Optional default'} />
                {selectedField.fieldtype === 'JSON' ? (
                  <BaseTextArea label="Options" value={selectedField.options} onChange={(event) => updateField(selectedField.id, () => ({ options: event.target.value }))} placeholder="JSON schema for child-table rendering" hint="For child tables, keep a JSON schema with mode=child-table and a columns array." />
                ) : (
                  <BaseInput
                    label="Options"
                    value={selectedField.options}
                    onChange={(event) => updateField(selectedField.id, () => ({ options: event.target.value }))}
                    placeholder={selectedField.fieldtype === 'Link' ? 'Target DocType, e.g. customer' : 'Options'}
                    hint={selectedField.fieldtype === 'Link' ? 'Link fields use options as the target DocType name for live lookups.' : undefined}
                  />
                )}
                <div className="grid gap-3 xl:col-span-2 xl:grid-cols-2">
                  <ToggleCard label="Mandatory" description="Make the field required during document creation." checked={selectedField.reqd} onChange={(value) => updateField(selectedField.id, () => ({ reqd: value }))} />
                  <ToggleCard label="Read-only" description="Render as a non-editable field in generated forms." checked={selectedField.read_only} onChange={(value) => updateField(selectedField.id, () => ({ read_only: value }))} />
                  <ToggleCard label="Unique" description="Hint that values should remain unique in this DocType." checked={selectedField.unique} onChange={(value) => updateField(selectedField.id, () => ({ unique: value }))} />
                  <ToggleCard label="Hidden" description="Keep the field in metadata while hiding it from normal entry screens." checked={selectedField.hidden} onChange={(value) => updateField(selectedField.id, () => ({ hidden: value }))} />
                  <ToggleCard label="List view" description="Prioritize this field for generated tables and desk list views." checked={selectedField.in_list_view} onChange={(value) => updateField(selectedField.id, () => ({ in_list_view: value }))} />
                </div>
              </div>
            ) : (
              <div className="mt-5 rounded-3xl border border-dashed border-white/10 bg-white/[0.03] px-4 py-5 text-sm leading-6 text-slate-400">
                Select a field from the canvas to edit it here.
              </div>
            )}
          </section>

          <details className="rounded-3xl border border-white/10 bg-slate-950/40 p-4 lg:p-5">
            <summary className="cursor-pointer list-none text-sm font-semibold text-white">Payload preview</summary>
            <pre className="mt-4 max-h-[320px] overflow-auto rounded-3xl border border-white/10 bg-slate-950 p-4 text-xs leading-6 text-cyan-200">
{JSON.stringify(payloadPreview, null, 2)}
            </pre>
          </details>

          <div className="flex flex-wrap items-center justify-between gap-3 rounded-3xl border border-white/10 bg-white/[0.03] px-4 py-4">
            <div>
              <div className="text-sm font-semibold text-white">Ready to save to Gogal?</div>
              <div className="mt-1 text-sm text-slate-400">This simpler builder sends clean JSON directly to your Go metadata endpoint.</div>
            </div>
            <div className="flex items-center gap-3">
              <button type="button" onClick={resetForm} className="rounded-2xl border border-white/10 bg-white/[0.05] px-4 py-2.5 text-sm font-medium text-slate-300 transition hover:bg-white/[0.1] active:scale-95">Reset</button>
              <button type="submit" disabled={busy} className="rounded-2xl bg-cyan-400 px-4 py-2.5 text-sm font-semibold text-slate-950 transition hover:bg-cyan-300 disabled:cursor-not-allowed disabled:opacity-60 active:scale-95">{busy ? 'Saving…' : 'Save DocType'}</button>
            </div>
          </div>
        </div>
      </form>
    </section>
  );
}
