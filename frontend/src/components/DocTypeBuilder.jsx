import { useEffect, useMemo, useState } from 'react';
import { childTableOptionsTemplate, isImageField } from '../lib/metadata.js';
import { createDocType } from '../lib/api.js';

const fieldTypeOptions = ['Data', 'Text', 'Check', 'Int', 'Float', 'Select', 'Link', 'Attach', 'Attach Image', 'Image', 'JSON'];

const fieldPalette = [
  { id: 'data', title: 'Data', fieldtype: 'Data', label: 'New Data Field', description: 'Short text, code, or title.' },
  { id: 'text', title: 'Text', fieldtype: 'Text', label: 'New Text Field', description: 'Longer notes or descriptions.' },
  { id: 'check', title: 'Check', fieldtype: 'Check', label: 'New Check Field', description: 'Boolean yes/no toggle.' },
  { id: 'int', title: 'Int', fieldtype: 'Int', label: 'New Integer Field', description: 'Numbers without decimals.' },
  { id: 'float', title: 'Float', fieldtype: 'Float', label: 'New Float Field', description: 'Numbers with decimals.' },
  { id: 'select', title: 'Select', fieldtype: 'Select', label: 'New Select Field', description: 'Pick from options.' },
  { id: 'link', title: 'Link', fieldtype: 'Link', label: 'Linked Record', description: 'Reference another DocType.', options: 'doctype_name' },
  { id: 'attach', title: 'Attach', fieldtype: 'Attach', label: 'Attachment', description: 'File or document URL.' },
  { id: 'attach-image', title: 'Attach Image', fieldtype: 'Attach Image', label: 'Profile Image', description: 'Avatar, logo, or image field.' },
  { id: 'json', title: 'JSON', fieldtype: 'JSON', label: 'JSON Data', description: 'Structured JSON value.' },
  { id: 'child-table', title: 'Child Table', fieldtype: 'JSON', label: 'Line Items', description: 'Nested row structure.', options: childTableOptionsTemplate() },
];

const disallowedListViewTypes = new Set(['Attach Image', 'Table', 'Tab Break', 'Section Break', 'Column Break']);

function makeId(prefix) {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

function slugifyFieldName(value) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '_')
    .replace(/^_+|_+$/g, '');
}

function createFieldFromTemplate(template, index = 0) {
  return {
    id: makeId('field'),
    label: template.label,
    fieldname: slugifyFieldName(`${template.label}_${index + 1}`),
    fieldtype: template.fieldtype,
    description: '',
    options: template.options || '',
    default: '',
    reqd: false,
    read_only: false,
    hidden: false,
    unique: false,
    in_list_view: true,
  };
}

function createColumn(title = 'Column 1') {
  return {
    id: makeId('column'),
    title,
    fields: [],
  };
}

function createSection(title = 'Main') {
  return {
    id: makeId('section'),
    title,
    columns: [createColumn()],
  };
}

function flattenCanvasFields(sections) {
  return sections.flatMap((section) => section.columns.flatMap((column) => column.fields));
}

function BaseInput({ label, ...props }) {
  return (
    <label className="grid gap-2">
      {label ? <span className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500">{label}</span> : null}
      <input className="field" {...props} />
    </label>
  );
}

function BaseTextArea({ label, ...props }) {
  return (
    <label className="grid gap-2">
      {label ? <span className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500">{label}</span> : null}
      <textarea className="field min-h-24" {...props} />
    </label>
  );
}

function BaseSelect({ label, children, ...props }) {
  return (
    <label className="grid gap-2">
      {label ? <span className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500">{label}</span> : null}
      <select className="field" {...props}>{children}</select>
    </label>
  );
}

function ToggleRow({ label, checked, onChange, disabled = false }) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={() => onChange(!checked)}
      className={`flex items-center justify-between rounded-2xl border px-3 py-2.5 text-left text-sm transition ${
        disabled
          ? 'cursor-not-allowed border-white/10 bg-white/[0.02] text-slate-500 opacity-60'
          : checked
            ? 'border-cyan-400/25 bg-cyan-500/10 text-white'
            : 'border-white/10 bg-white/[0.03] text-slate-300 hover:bg-white/[0.05]'
      }`}
    >
      <span>{label}</span>
      <span className={`inline-flex h-5 w-10 rounded-full p-1 transition ${checked ? 'bg-cyan-400/80' : 'bg-slate-700'}`}>
        <span className={`h-3 w-3 rounded-full bg-white transition ${checked ? 'translate-x-5' : ''}`} />
      </span>
    </button>
  );
}

function PaletteButton({ template, onClick }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="w-full rounded-2xl border border-white/10 bg-white/[0.03] px-3 py-3 text-left transition hover:border-white/20 hover:bg-white/[0.06]"
    >
      <div className="flex items-start justify-between gap-3">
        <div>
          <div className="text-sm font-semibold text-white">{template.title}</div>
          <div className="mt-1 text-xs leading-5 text-slate-500">{template.description}</div>
        </div>
        <span className="rounded-full border border-white/10 px-2 py-1 text-[10px] uppercase tracking-[0.18em] text-slate-300">
          {template.fieldtype}
        </span>
      </div>
    </button>
  );
}

function normalizeFieldPayload(field, index) {
  return {
    label: field.label.trim(),
    fieldname: slugifyFieldName(field.fieldname || field.label),
    fieldtype: field.fieldtype,
    description: String(field.description || '').trim(),
    options: String(field.options || '').trim(),
    default: String(field.default || '').trim(),
    reqd: Boolean(field.reqd),
    read_only: Boolean(field.read_only),
    hidden: Boolean(field.hidden),
    unique: Boolean(field.unique),
    in_list_view: Boolean(field.in_list_view),
    sort_order: index + 1,
  };
}

function createLayoutField(fieldtype, index, label = '') {
  const baseName = fieldtype.toLowerCase().replace(/\s+/g, '_');
  return {
    id: makeId(baseName),
    label,
    fieldname: `${baseName}_${index + 1}`,
    fieldtype,
    description: '',
    options: '',
    default: '',
    reqd: false,
    read_only: false,
    hidden: false,
    unique: false,
    in_list_view: false,
  };
}

function serializeCanvasFields(sections) {
  const serialized = [];

  sections.forEach((section, sectionIndex) => {
    const sectionLabel = String(section.title || '').trim();
    const includeSectionBreak = sectionIndex > 0 || (sectionLabel && sectionLabel.toLowerCase() !== 'main');

    if (includeSectionBreak) {
      serialized.push(createLayoutField('Section Break', sectionIndex, sectionLabel));
    }

    section.columns.forEach((column, columnIndex) => {
      const columnLabel = String(column.title || '').trim();
      if (columnIndex > 0) {
        serialized.push(createLayoutField('Column Break', serialized.length, columnLabel));
      }

      column.fields.forEach((field) => {
        serialized.push(field);
      });
    });
  });

  return serialized;
}

function validateBuilderFields(fields) {
  const seen = new Set();

  for (const field of fields) {
    const label = field.label || field.fieldname || 'Field';
    const fieldname = slugifyFieldName(field.fieldname || field.label);

    if (!fieldname) {
      return 'Each field needs a label and valid field name.';
    }

    if (seen.has(fieldname)) {
      return `Fieldname ${label} (${fieldname}) appears multiple times`;
    }
    seen.add(fieldname);

    if ((field.fieldtype === 'Link' || field.fieldtype === 'Table') && !String(field.options || '').trim()) {
      return `Options is required for field ${label} (${fieldname}) of type ${field.fieldtype}`;
    }

    if (field.hidden && field.reqd && !String(field.default || '').trim()) {
      return `${label} (${fieldname}) cannot be hidden and mandatory without any default value`;
    }

    if (field.in_list_view && disallowedListViewTypes.has(field.fieldtype)) {
      return `'In List View' is not allowed for field ${label} (${fieldname}) of type ${field.fieldtype}`;
    }
  }

  return '';
}

function findTargetColumn(sections, sectionId, columnId) {
  const fallbackSection = sections[0];
  const section = sections.find((item) => item.id === sectionId) || fallbackSection;
  const column = section?.columns.find((item) => item.id === columnId) || section?.columns[0] || null;
  return { section, column };
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
  const [sections, setSections] = useState([createSection()]);
  const [activeSectionId, setActiveSectionId] = useState('');
  const [activeColumnId, setActiveColumnId] = useState('');
  const [selectedFieldId, setSelectedFieldId] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const availableModules = useMemo(() => ['Core', ...moduleOptions.filter((item) => item && item !== 'Core')], [moduleOptions]);
  const allFields = useMemo(() => flattenCanvasFields(sections), [sections]);
  const serializedFields = useMemo(() => serializeCanvasFields(sections), [sections]);
  const availableImageFields = useMemo(() => allFields.filter((field) => isImageField(field)), [allFields]);
  const selectedField = useMemo(() => allFields.find((field) => field.id === selectedFieldId) || null, [allFields, selectedFieldId]);

  const payloadPreview = useMemo(() => ({
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
    fields: serializedFields.map((field, index) => normalizeFieldPayload(field, index)),
  }), [allowRename, description, doctype, imageField, isChildTable, isSingle, label, maxAttachments, moduleName, quickEntry, serializedFields]);

  useEffect(() => {
    if (!activeSectionId || !activeColumnId) {
      const firstSection = sections[0];
      const firstColumn = firstSection?.columns[0];
      if (firstSection) {
        setActiveSectionId((current) => current || firstSection.id);
      }
      if (firstColumn) {
        setActiveColumnId((current) => current || firstColumn.id);
      }
    }
  }, [activeColumnId, activeSectionId, sections]);

  useEffect(() => {
    if (imageField && !availableImageFields.some((field) => field.fieldname === imageField)) {
      setImageField('');
    }
  }, [availableImageFields, imageField]);

  const updateSections = (updater) => {
    setSections((current) => updater(current));
  };

  const updateField = (fieldId, updater) => {
    updateSections((current) => current.map((section) => ({
      ...section,
      columns: section.columns.map((column) => ({
        ...column,
        fields: column.fields.map((field) => (field.id === fieldId ? { ...field, ...updater(field) } : field)),
      })),
    })));
  };

  const addSection = () => {
    const created = createSection(`Section ${sections.length + 1}`);
    setSections((current) => [...current, created]);
    setActiveSectionId(created.id);
    setActiveColumnId(created.columns[0].id);
  };

  const addColumn = (sectionId) => {
    if (!sectionId) {
      return;
    }

    const created = createColumn();
    setSections((current) => current.map((section) => (
      section.id === sectionId
        ? { ...section, columns: [...section.columns, { ...created, title: `Column ${section.columns.length + 1}` }] }
        : section
    )));
    setActiveSectionId(sectionId);
    setActiveColumnId(created.id);
  };

  const addField = (template) => {
    const { section, column } = findTargetColumn(sections, activeSectionId, activeColumnId);
    if (!section || !column) {
      return;
    }
    const created = createFieldFromTemplate(template, allFields.length + 1);
    setSections((current) => current.map((sectionItem) => (
      sectionItem.id === section.id
        ? {
            ...sectionItem,
            columns: sectionItem.columns.map((columnItem) => (
              columnItem.id === column.id
                ? { ...columnItem, fields: [...columnItem.fields, created] }
                : columnItem
            )),
          }
        : sectionItem
    )));
    setSelectedFieldId(created.id);
  };

  const removeField = (fieldId) => {
    setSections((current) => current.map((section) => ({
      ...section,
      columns: section.columns.map((column) => ({
        ...column,
        fields: column.fields.filter((field) => field.id !== fieldId),
      })),
    })));
    if (selectedFieldId === fieldId) {
      setSelectedFieldId('');
    }
  };

  const resetForm = ({ keepSuccess = false } = {}) => {
    const fresh = createSection();
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
    setSections([fresh]);
    setActiveSectionId(fresh.id);
    setActiveColumnId(fresh.columns[0].id);
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
    if (!normalizedDocType) {
      setError('DocType name is required.');
      return;
    }

    const existingMatch = existingDocTypes.some((item) => item.toLowerCase() === normalizedDocType.toLowerCase());
    if (existingMatch) {
      setError(`DocType ${normalizedDocType} already exists.`);
      return;
    }

    if (allFields.length === 0) {
      setError('Add at least one field from the left sidebar.');
      return;
    }

    const normalizedFields = serializedFields.map((field, index) => normalizeFieldPayload(field, index));
    const validationMessage = validateBuilderFields(normalizedFields);
    if (validationMessage) {
      setError(validationMessage);
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
        fields: normalizedFields,
      };
      const created = await createDocType(payload);
      onCreated?.(created || payload);
      resetForm({ keepSuccess: true });
      setSuccess(`DocType ${created?.doctype || normalizedDocType} created successfully.`);
    } catch (submitError) {
      setError(submitError.message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <section className="space-y-5">
      <div className="panel overflow-hidden">
        <div className="border-b border-white/10 px-5 py-4 lg:px-6">
          <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <div className="mb-2 flex flex-wrap gap-2">
                <span className="badge">Simple Builder</span>
                <span className="badge">Frappe Style</span>
              </div>
              <h2 className="text-xl font-semibold text-white">DocType Builder</h2>
              <p className="mt-1 text-sm text-slate-400">Add fields from the left, select a field, then edit its properties on the right.</p>
            </div>
            <div className="flex gap-2">
              <button type="button" onClick={addSection} className="rounded-2xl border border-white/10 bg-white/[0.05] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-white/[0.1]">+ Section</button>
              <button type="button" onClick={() => addColumn(activeSectionId || sections[0]?.id)} className="rounded-2xl border border-white/10 bg-white/[0.05] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-white/[0.1]">+ Column</button>
            </div>
          </div>
        </div>

        <form className="grid gap-5 p-5 xl:grid-cols-[320px_minmax(0,1fr)_360px] xl:p-6" onSubmit={handleSubmit}>
          <aside className="space-y-4">
            <section className="rounded-[26px] border border-white/10 bg-white/[0.03] p-4">
              <div className="text-[11px] font-semibold uppercase tracking-[0.22em] text-slate-500">DocType</div>
              <div className="mt-4 grid gap-3">
                <BaseInput label="Name" value={doctype} onChange={(event) => setDocType(event.target.value)} placeholder="sales_invoice" />
                <BaseInput label="Label" value={label} onChange={(event) => setLabel(event.target.value)} placeholder="Sales Invoice" />
                <BaseSelect label="Module" value={moduleName} onChange={(event) => setModuleName(event.target.value)}>
                  {availableModules.map((option) => <option key={option} value={option}>{option}</option>)}
                </BaseSelect>
                <BaseTextArea label="Description" value={description} onChange={(event) => setDescription(event.target.value)} placeholder="Brief doctype description" />
              </div>
            </section>

            <section className="rounded-[26px] border border-white/10 bg-white/[0.03] p-4">
              <div className="text-[11px] font-semibold uppercase tracking-[0.22em] text-slate-500">Behavior</div>
              <div className="mt-4 grid gap-2">
                <ToggleRow label="Is Single" checked={isSingle} onChange={(value) => {
                  setIsSingle(value);
                  if (value) {
                    setIsChildTable(false);
                    setQuickEntry(false);
                  }
                }} />
                <ToggleRow label="Is Child Table" checked={isChildTable} disabled={isSingle} onChange={(value) => {
                  setIsChildTable(value);
                  if (value) {
                    setIsSingle(false);
                    setQuickEntry(false);
                  }
                }} />
                <ToggleRow label="Allow Rename" checked={allowRename} onChange={setAllowRename} />
                <ToggleRow label="Quick Entry" checked={isSingle || isChildTable ? false : quickEntry} disabled={isSingle || isChildTable} onChange={setQuickEntry} />
                <BaseInput label="Max attachments" type="number" min="0" value={maxAttachments} onChange={(event) => setMaxAttachments(event.target.value)} placeholder="0" />
                <BaseSelect label="Image field" value={imageField} onChange={(event) => setImageField(event.target.value)}>
                  <option value="">No image field</option>
                  {availableImageFields.map((field) => <option key={field.id} value={field.fieldname}>{field.label}</option>)}
                </BaseSelect>
              </div>
            </section>

            <section className="rounded-[26px] border border-white/10 bg-white/[0.03] p-4">
              <div className="flex items-center justify-between gap-2">
                <div className="text-[11px] font-semibold uppercase tracking-[0.22em] text-slate-500">Add field</div>
                <div className="rounded-full border border-cyan-400/20 bg-cyan-500/10 px-2 py-1 text-[10px] uppercase tracking-[0.18em] text-cyan-100">{allFields.length}</div>
              </div>
              <div className="mt-4 grid gap-2">
                {fieldPalette.map((template) => (
                  <PaletteButton key={template.id} template={template} onClick={() => addField(template)} />
                ))}
              </div>
            </section>

            <section className="rounded-[26px] border border-white/10 bg-white/[0.03] p-4">
              <div className="text-[11px] font-semibold uppercase tracking-[0.22em] text-slate-500">Fields</div>
              <div className="mt-4 space-y-3">
                {sections.map((section) => (
                  <div key={`nav-${section.id}`} className="space-y-2">
                    <div className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{section.title}</div>
                    {section.columns.flatMap((column) => column.fields).length > 0 ? section.columns.flatMap((column) => column.fields).map((field) => (
                      <button
                        key={`nav-field-${field.id}`}
                        type="button"
                        onClick={() => setSelectedFieldId(field.id)}
                        className={`block w-full rounded-xl border px-3 py-2 text-left text-sm transition ${selectedFieldId === field.id ? 'border-cyan-400/25 bg-cyan-500/10 text-white' : 'border-white/10 bg-white/[0.03] text-slate-300 hover:bg-white/[0.06]'}`}
                      >
                        <div className="truncate font-medium">{field.label || 'Untitled field'}</div>
                        <div className="mt-1 text-[11px] uppercase tracking-[0.18em] text-slate-500">{field.fieldtype}</div>
                      </button>
                    )) : <div className="rounded-xl border border-dashed border-white/10 px-3 py-3 text-sm text-slate-500">No fields yet</div>}
                  </div>
                ))}
              </div>
            </section>
          </aside>

          <div className="space-y-4">
            {error ? <div className="rounded-2xl border border-rose-400/20 bg-rose-500/10 px-4 py-3 text-sm text-rose-100">{error}</div> : null}
            {success ? <div className="rounded-2xl border border-emerald-400/20 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-100">{success}</div> : null}

            {sections.map((section, sectionIndex) => (
              <section
                key={section.id}
                className={`rounded-[30px] border p-4 transition lg:p-5 ${activeSectionId === section.id ? 'border-cyan-400/25 bg-cyan-500/[0.04]' : 'border-white/10 bg-white/[0.03]'}`}
                onClick={() => {
                  setActiveSectionId(section.id);
                  setActiveColumnId(section.columns[0]?.id || '');
                }}
              >
                <div className="flex items-center justify-between gap-3">
                  <div className="min-w-0 flex-1">
                    <div className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500">Section {sectionIndex + 1}</div>
                    <div className="mt-3">
                      <BaseInput label="Section title" value={section.title} onChange={(event) => setSections((current) => current.map((item) => item.id === section.id ? { ...item, title: event.target.value } : item))} placeholder="Section title" />
                    </div>
                  </div>
                  <div className="rounded-full border border-white/10 px-3 py-1 text-xs text-slate-300">{section.columns.reduce((count, column) => count + column.fields.length, 0)} fields</div>
                </div>

                <div className={`mt-4 grid gap-4 ${section.columns.length === 1 ? 'lg:grid-cols-1' : section.columns.length === 2 ? 'lg:grid-cols-2' : 'lg:grid-cols-3'}`}>
                  {section.columns.map((column, columnIndex) => (
                    <div
                      key={column.id}
                      className={`rounded-[24px] border p-4 transition ${activeColumnId === column.id ? 'border-cyan-400/25 bg-slate-950/60' : 'border-white/10 bg-slate-950/40'}`}
                      onClick={(event) => {
                        event.stopPropagation();
                        setActiveSectionId(section.id);
                        setActiveColumnId(column.id);
                      }}
                    >
                      <div className="flex items-center justify-between gap-2">
                        <div>
                          <div className="text-sm font-semibold text-white">{column.title || `Column ${columnIndex + 1}`}</div>
                          <div className="mt-1 text-[11px] uppercase tracking-[0.18em] text-slate-500">Target column</div>
                        </div>
                        <span className="rounded-full border border-white/10 px-2 py-1 text-[10px] uppercase tracking-[0.18em] text-slate-300">{column.fields.length}</span>
                      </div>

                      <div className="mt-3">
                        <BaseInput label="Column title" value={column.title} onChange={(event) => setSections((current) => current.map((sectionItem) => sectionItem.id === section.id ? { ...sectionItem, columns: sectionItem.columns.map((columnItem) => columnItem.id === column.id ? { ...columnItem, title: event.target.value } : columnItem) } : sectionItem))} placeholder={`Column ${columnIndex + 1}`} />
                      </div>

                      <div className="mt-4 space-y-2">
                        {column.fields.length > 0 ? column.fields.map((field) => (
                          <div key={field.id} className={`rounded-2xl border p-3 transition ${selectedFieldId === field.id ? 'border-cyan-400/25 bg-cyan-500/10' : 'border-white/10 bg-white/[0.03]'}`}>
                            <div className="flex items-start justify-between gap-3">
                              <button type="button" onClick={() => setSelectedFieldId(field.id)} className="min-w-0 flex-1 text-left">
                                <div className="truncate text-sm font-semibold text-white">{field.label || 'Untitled field'}</div>
                                <div className="mt-1 flex flex-wrap gap-2 text-[11px] uppercase tracking-[0.16em] text-slate-500">
                                  <span>{field.fieldtype}</span>
                                  <span>{field.fieldname}</span>
                                </div>
                                {field.description ? <div className="mt-2 text-xs leading-5 text-slate-500">{field.description}</div> : null}
                              </button>
                              <button type="button" onClick={() => removeField(field.id)} className="rounded-full border border-rose-400/20 bg-rose-500/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.16em] text-rose-100">Remove</button>
                            </div>
                          </div>
                        )) : (
                          <div className="rounded-2xl border border-dashed border-white/10 px-4 py-6 text-center text-sm text-slate-500">Select a field from the left sidebar to add it here.</div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </section>
            ))}

            <div className="flex flex-wrap items-center justify-between gap-3 rounded-[26px] border border-white/10 bg-white/[0.03] px-4 py-4">
              <div>
                <div className="text-sm font-semibold text-white">Ready to save</div>
                <div className="mt-1 text-sm text-slate-400">The builder sends a clean flat field list to the current metadata API.</div>
              </div>
              <div className="flex gap-3">
                <button type="button" onClick={resetForm} className="rounded-2xl border border-white/10 bg-white/[0.05] px-4 py-2.5 text-sm font-medium text-slate-300 transition hover:bg-white/[0.1]">Reset</button>
                <button type="submit" disabled={busy} className="rounded-2xl bg-cyan-400 px-4 py-2.5 text-sm font-semibold text-slate-950 transition hover:bg-cyan-300 disabled:cursor-not-allowed disabled:opacity-60">{busy ? 'Saving…' : 'Save DocType'}</button>
              </div>
            </div>
          </div>

          <aside className="space-y-4">
            <section className="rounded-[26px] border border-white/10 bg-white/[0.03] p-4 lg:p-5">
              <div className="text-[11px] font-semibold uppercase tracking-[0.22em] text-slate-500">Field properties</div>
              {selectedField ? (
                <div className="mt-4 grid gap-4">
                  <div>
                    <h3 className="text-lg font-semibold text-white">{selectedField.label || 'Selected field'}</h3>
                    <p className="mt-1 text-sm text-slate-400">Edit the selected field like Frappe/Odoo properties.</p>
                  </div>

                  <BaseInput label="Field label" value={selectedField.label} onChange={(event) => updateField(selectedField.id, (current) => ({ label: event.target.value, fieldname: current.fieldname || slugifyFieldName(event.target.value) }))} placeholder="Customer Name" />
                  <BaseInput label="Field name" value={selectedField.fieldname} onChange={(event) => updateField(selectedField.id, () => ({ fieldname: slugifyFieldName(event.target.value) }))} placeholder="customer_name" />
                  <BaseSelect label="Field type" value={selectedField.fieldtype} onChange={(event) => updateField(selectedField.id, () => ({ fieldtype: event.target.value }))}>
                    {fieldTypeOptions.map((option) => <option key={option} value={option}>{option}</option>)}
                  </BaseSelect>
                  <BaseTextArea label="Description" value={selectedField.description || ''} onChange={(event) => updateField(selectedField.id, () => ({ description: event.target.value }))} placeholder="Helper text shown under the field" />
                  <BaseInput label="Default" value={selectedField.default} onChange={(event) => updateField(selectedField.id, () => ({ default: event.target.value }))} placeholder="Optional default value" />
                  {selectedField.fieldtype === 'JSON' ? (
                    <BaseTextArea label="Options" value={selectedField.options} onChange={(event) => updateField(selectedField.id, () => ({ options: event.target.value }))} placeholder='{"columns":[]}' />
                  ) : (
                    <BaseInput label="Options" value={selectedField.options} onChange={(event) => updateField(selectedField.id, () => ({ options: event.target.value }))} placeholder={selectedField.fieldtype === 'Link' ? 'Target DocType' : 'Option 1'} />
                  )}

                  <div className="grid gap-2">
                    <ToggleRow label="Mandatory" checked={selectedField.reqd} onChange={(value) => updateField(selectedField.id, () => ({ reqd: value }))} />
                    <ToggleRow label="Unique" checked={selectedField.unique} onChange={(value) => updateField(selectedField.id, () => ({ unique: value }))} />
                    <ToggleRow label="Read Only" checked={selectedField.read_only} onChange={(value) => updateField(selectedField.id, () => ({ read_only: value }))} />
                    <ToggleRow label="Hidden" checked={selectedField.hidden} onChange={(value) => updateField(selectedField.id, () => ({ hidden: value }))} />
                    <ToggleRow label="In List View" checked={selectedField.in_list_view} onChange={(value) => updateField(selectedField.id, () => ({ in_list_view: value }))} />
                  </div>
                </div>
              ) : (
                <div className="mt-4 rounded-2xl border border-dashed border-white/10 px-4 py-6 text-sm text-slate-500">
                  Select a field from the left sidebar or the canvas to edit its properties.
                </div>
              )}
            </section>

            <details className="rounded-[26px] border border-white/10 bg-white/[0.03] p-4 lg:p-5">
              <summary className="cursor-pointer list-none text-sm font-semibold text-white">Payload preview</summary>
              <pre className="mt-4 max-h-[300px] overflow-auto rounded-3xl border border-white/10 bg-slate-950 p-4 text-xs leading-6 text-cyan-200">
{JSON.stringify(payloadPreview, null, 2)}
              </pre>
            </details>
          </aside>
        </form>
      </div>
    </section>
  );
}
