import { useEffect, useMemo, useState } from 'react';
import {
  closestCenter,
  DndContext,
  DragOverlay,
  KeyboardSensor,
  PointerSensor,
	useDraggable,
  useDroppable,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { childTableOptionsTemplate, isImageField } from '../lib/metadata.js';
import { createDocType } from '../lib/api.js';

const canvasDropZoneId = 'builder-canvas';

const fieldTypeOptions = ['Data', 'Text', 'Check', 'Int', 'Float', 'Select', 'Link', 'Attach', 'Attach Image', 'Image', 'JSON'];

const fieldPalette = [
  {
    id: 'palette-data',
    title: 'Data Field',
    fieldtype: 'Data',
    label: 'New Data Field',
    description: 'Single-line text for names, titles, references, or search keys.',
    accent: 'cyan',
  },
  {
    id: 'palette-check',
    title: 'Check Field',
    fieldtype: 'Check',
    label: 'New Check Field',
    description: 'Boolean toggle for flags like Active, Approved, or Is Paid.',
    accent: 'emerald',
  },
  {
    id: 'palette-int',
    title: 'Int Field',
    fieldtype: 'Int',
    label: 'New Integer Field',
    description: 'Whole-number values for quantities, priorities, or counters.',
    accent: 'violet',
  },
  {
    id: 'palette-link',
    title: 'Link Field',
    fieldtype: 'Link',
    label: 'Linked Record',
    description: 'Lookup another DocType and store a live relationship, similar to ERP-style links.',
    accent: 'cyan',
    options: 'doctype_name',
  },
  {
    id: 'palette-attach',
    title: 'Attach File',
    fieldtype: 'Attach',
    label: 'Attachment',
    description: 'Store a file URL or uploaded file reference for contracts, PDFs, or supporting evidence.',
    accent: 'violet',
  },
  {
    id: 'palette-attach-image',
    title: 'Attach Image',
    fieldtype: 'Attach Image',
    label: 'Profile Image',
    description: 'Store an image path and surface it in form and sidebar previews like a profile or logo field.',
    accent: 'emerald',
  },
  {
    id: 'palette-child-table',
    title: 'Child Table',
    fieldtype: 'JSON',
    label: 'Line Items',
    description: 'Render repeated nested rows using a JSON-backed child-table schema for invoices, schedules, or checklists.',
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

function PaletteCard({ template, onAdd }) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({ id: template.id });

  const style = {
    transform: CSS.Translate.toString(transform),
  };

  const paletteTone = {
    cyan: 'border-cyan-400/20 bg-cyan-500/10 text-cyan-100',
    emerald: 'border-emerald-400/20 bg-emerald-500/10 text-emerald-100',
    violet: 'border-violet-400/20 bg-violet-500/10 text-violet-100',
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`rounded-3xl border border-white/10 bg-white/[0.04] p-4 transition hover:-translate-y-0.5 hover:border-white/15 hover:bg-white/[0.06] ${isDragging ? 'opacity-70' : ''}`}
    >
      <div className="flex items-start justify-between gap-3">
        <div>
          <div className="text-sm font-semibold text-white">{template.title}</div>
          <p className="mt-2 text-sm leading-6 text-slate-400">{template.description}</p>
        </div>
        <span className={`rounded-full border px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.2em] ${paletteTone[template.accent]}`}>
          {template.fieldtype}
        </span>
      </div>
      <div className="mt-4 flex items-center justify-between gap-3">
        <div className="text-xs uppercase tracking-[0.24em] text-slate-500">Drag to canvas or quick add</div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            className="rounded-full border border-white/10 bg-white/[0.06] px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-white/[0.12] active:scale-95"
            {...listeners}
            {...attributes}
          >
            Drag
          </button>
          <button
            type="button"
            onClick={onAdd}
            className="rounded-full border border-white/10 bg-white/[0.06] px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-white/[0.12] active:scale-95"
          >
            Add
          </button>
        </div>
      </div>
    </div>
  );
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

function BuilderDropZone({ children, isEmpty }) {
  const { isOver, setNodeRef } = useDroppable({ id: canvasDropZoneId });

  return (
    <div
      ref={setNodeRef}
      className={`rounded-[28px] border-2 border-dashed p-4 transition ${
        isOver ? 'border-cyan-400/50 bg-cyan-500/10' : 'border-white/10 bg-white/[0.03]'
      } ${isEmpty ? 'min-h-[260px]' : ''}`}
    >
      {children}
    </div>
  );
}

function SortableFieldRow({ field, active, onSelect, onRemove }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: field.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`rounded-3xl border px-4 py-4 transition ${
        active ? 'border-cyan-400/30 bg-cyan-500/10 shadow-lg shadow-cyan-500/10' : 'border-white/10 bg-white/[0.04]'
      } ${isDragging ? 'opacity-70' : ''}`}
    >
      <div className="flex items-center justify-between gap-3">
        <button
          type="button"
          onClick={onSelect}
          className="flex flex-1 items-center gap-3 text-left"
        >
          <span
            type="button"
            className="inline-flex h-10 w-10 items-center justify-center rounded-2xl border border-white/10 bg-slate-950/70 text-lg text-slate-300"
            {...attributes}
            {...listeners}
          >
            ⋮⋮
          </span>
          <div>
            <div className="text-sm font-semibold text-white">{field.label || 'Untitled field'}</div>
            <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-slate-500">
              <span>{field.fieldname || 'field_name'}</span>
              <span>•</span>
              <span>{field.fieldtype}</span>
              {field.reqd ? <><span>•</span><span>Mandatory</span></> : null}
              {field.unique ? <><span>•</span><span>Unique</span></> : null}
            </div>
          </div>
        </button>
        <button
          type="button"
          onClick={onRemove}
          className="rounded-full border border-white/10 bg-white/[0.05] px-3 py-1.5 text-xs font-semibold text-slate-300 transition hover:bg-white/[0.1] active:scale-95"
        >
          Remove
        </button>
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
  const [activeDragId, setActiveDragId] = useState(null);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

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

  const activePalette = fieldPalette.find((item) => item.id === activeDragId);

  const updateField = (fieldId, updater) => {
    setFields((current) =>
      current.map((field) => (field.id === fieldId ? { ...field, ...updater(field) } : field)),
    );
  };

  const addFieldFromTemplate = (template, targetIndex = fields.length) => {
    const created = createFieldFromTemplate(template, fields.length + 1);
    setFields((current) => {
      const next = [...current];
      next.splice(targetIndex, 0, created);
      return next;
    });
    setSelectedFieldId(created.id);
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

  const handleFieldDrop = (activeId, overId) => {
    if (activeId.startsWith('palette-')) {
      const template = fieldPalette.find((item) => item.id === activeId);
      if (!template) {
        return;
      }

      const overIndex = overId === canvasDropZoneId ? fields.length : fields.findIndex((field) => field.id === overId);
      addFieldFromTemplate(template, overIndex >= 0 ? overIndex : fields.length);
      return;
    }

    if (activeId !== overId) {
      const oldIndex = fields.findIndex((field) => field.id === activeId);
      const newIndex = fields.findIndex((field) => field.id === overId);
      if (oldIndex === -1 || newIndex === -1) {
        return;
      }
      setFields((current) => arrayMove(current, oldIndex, newIndex));
    }
  };

  const handleDragStart = (event) => {
    setActiveDragId(String(event.active.id));
  };

  const handleDragEnd = (event) => {
    const activeId = String(event.active.id);
    const overId = event.over ? String(event.over.id) : '';
    setActiveDragId(null);

    if (!overId) {
      return;
    }

    handleFieldDrop(activeId, overId);
  };

  const handleDragCancel = () => {
    setActiveDragId(null);
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
      setError('Add at least one field by dragging a palette card into the canvas.');
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
    <section className="panel overflow-hidden">
      <div className="border-b border-white/10 px-6 py-5">
        <div className="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
          <div>
            <div className="mb-2 flex flex-wrap gap-2">
              <span className="badge">UI Studio</span>
              <span className="badge">Metadata-first</span>
              <span className="badge">POST /api/doctypes</span>
            </div>
            <h2 className="text-2xl font-semibold text-white">DocType Builder</h2>
            <p className="mt-2 max-w-3xl text-sm leading-6 text-slate-400">
              Drag in field primitives, reorder them visually, tune properties from the side panel, and ship a clean JSON schema directly to the Go backend.
            </p>
          </div>

          <div className="rounded-3xl border border-emerald-400/15 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-100">
            <div className="font-semibold">Touch + animation aware</div>
            <div className="mt-1 text-emerald-100/75">Large targets, live payload preview, and future-ready hooks for scheduler/automation modules.</div>
          </div>
        </div>
      </div>

      <form className="grid gap-0 xl:grid-cols-[340px_minmax(0,1fr)_380px]" onSubmit={handleSubmit}>
        <div className="border-b border-white/10 p-6 xl:border-b-0 xl:border-r">
          <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">Field palette</div>
          <div className="mt-4 space-y-4">
            {fieldPalette.map((template) => (
              <PaletteCard
                key={template.id}
                template={template}
                onAdd={() => addFieldFromTemplate(template)}
              />
            ))}
          </div>

          <div className="mt-6 rounded-3xl border border-white/10 bg-white/[0.03] p-4">
            <div className="text-sm font-semibold text-white">Builder setup</div>
            <div className="mt-4 grid gap-4">
              <BaseInput
                label="DocType name"
                value={doctype}
                onChange={(event) => setDocType(event.target.value)}
                placeholder="CustomerInvoice"
                hint="Used as the API-facing DocType identifier."
              />
              <BaseInput
                label="Label"
                value={label}
                onChange={(event) => setLabel(event.target.value)}
                placeholder="Customer Invoice"
                hint="Human-readable label used in desk navigation and forms."
              />
              <BaseSelect label="Module" value={moduleName} onChange={(event) => setModuleName(event.target.value)}>
                {availableModules.map((moduleOption) => (
                  <option key={moduleOption} value={moduleOption}>{moduleOption}</option>
                ))}
              </BaseSelect>
              <BaseTextArea
                label="Description"
                value={description}
                onChange={(event) => setDescription(event.target.value)}
                placeholder="What workflow or business object does this DocType represent?"
              />
              <div className="grid gap-3">
                <ToggleCard
                  label="Is Single"
                  description="Create a singleton settings-style DocType. The builder disables conflicting options now so the runtime can support single CRUD cleanly next."
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
                  description="Use this DocType as nested rows inside a parent document through a Table field, Frappe-style but with explicit Go runtime constraints."
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
                <ToggleCard
                  label="Allow Rename"
                  description="Prepare this DocType for future rename support without changing the schema contract later."
                  checked={allowRename}
                  onChange={setAllowRename}
                />
                <ToggleCard
                  label="Quick Entry"
                  description="Reserve a compact create flow for normal doctypes. This stays disabled for single and child-table doctypes."
                  checked={isSingle || isChildTable ? false : quickEntry}
                  onChange={setQuickEntry}
                  disabled={isSingle || isChildTable}
                />
              </div>
              <BaseInput
                label="Max attachments"
                type="number"
                min="0"
                value={maxAttachments}
                onChange={(event) => setMaxAttachments(event.target.value)}
                placeholder="0"
                hint="Use 0 when you do not want to enforce a limit yet."
              />
              <BaseSelect label="Image field" value={imageField} onChange={(event) => setImageField(event.target.value)}>
                <option value="">No profile/sidebar image</option>
                {availableImageFields.map((fieldOption) => (
                  <option key={fieldOption.id} value={fieldOption.fieldname}>{fieldOption.label} ({fieldOption.fieldname})</option>
                ))}
              </BaseSelect>
            </div>
          </div>
        </div>

        <div className="border-b border-white/10 p-6 xl:border-b-0 xl:border-r">
          {error ? <div className="mb-4 rounded-2xl border border-rose-400/20 bg-rose-500/10 px-4 py-3 text-sm text-rose-100">{error}</div> : null}
          {success ? <div className="mb-4 rounded-2xl border border-emerald-400/20 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-100">{success}</div> : null}

          <div className="mb-4 flex items-center justify-between gap-3">
            <div>
              <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">Canvas</div>
              <h3 className="mt-2 text-xl font-semibold text-white">Arrange fields visually</h3>
              <p className="mt-2 text-sm leading-6 text-slate-400">Drag field cards into the canvas, then reorder them to define the final metadata sequence.</p>
            </div>
            <div className="rounded-2xl border border-white/10 bg-white/[0.03] px-3 py-2 text-right">
              <div className="text-xs uppercase tracking-[0.24em] text-slate-500">Fields</div>
              <div className="mt-1 text-lg font-semibold text-white">{fields.length}</div>
            </div>
          </div>

          <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragStart={handleDragStart}
            onDragEnd={handleDragEnd}
            onDragCancel={handleDragCancel}
          >
            <BuilderDropZone isEmpty={fields.length === 0}>
              {fields.length === 0 ? (
                <div className="flex min-h-[220px] flex-col items-center justify-center rounded-[24px] border border-dashed border-white/10 bg-slate-950/40 px-6 py-10 text-center">
                  <div className="rounded-full border border-cyan-400/20 bg-cyan-500/10 px-4 py-1.5 text-xs font-semibold uppercase tracking-[0.24em] text-cyan-100">
                    Drop zone ready
                  </div>
                  <h4 className="mt-4 text-xl font-semibold text-white">Drag fields into this canvas to shape your DocType</h4>
                  <p className="mt-3 max-w-xl text-sm leading-6 text-slate-400">
                    This builder now supports link, file, and image-focused fields while Gogal closes the remaining gaps to a fuller Frappe-style metadata model.
                  </p>
                </div>
              ) : (
                <SortableContext items={fields.map((field) => field.id)} strategy={verticalListSortingStrategy}>
                  <div className="space-y-3">
                    {fields.map((field) => (
                      <SortableFieldRow
                        key={field.id}
                        field={field}
                        active={selectedField?.id === field.id}
                        onSelect={() => setSelectedFieldId(field.id)}
                        onRemove={() => removeField(field.id)}
                      />
                    ))}
                  </div>
                </SortableContext>
              )}
            </BuilderDropZone>

            <DragOverlay>
              {activePalette ? (
                <div className="w-[280px] rounded-3xl border border-cyan-300/30 bg-slate-950/90 p-4 shadow-2xl shadow-cyan-500/20">
                  <div className="text-sm font-semibold text-white">{activePalette.title}</div>
                  <div className="mt-2 text-sm text-slate-400">{activePalette.description}</div>
                </div>
              ) : null}
            </DragOverlay>
          </DndContext>

          <div className="mt-6 flex flex-wrap items-center justify-between gap-3 rounded-3xl border border-white/10 bg-white/[0.03] px-4 py-4">
            <div>
              <div className="text-sm font-semibold text-white">Ready to save to Gogal?</div>
              <div className="mt-1 text-sm text-slate-400">The builder sends clean JSON directly to your Go metadata endpoint.</div>
            </div>
            <div className="flex items-center gap-3">
              <button
                type="button"
                onClick={resetForm}
                className="rounded-2xl border border-white/10 bg-white/[0.05] px-4 py-2.5 text-sm font-medium text-slate-300 transition hover:bg-white/[0.1] active:scale-95"
              >
                Reset
              </button>
              <button
                type="submit"
                disabled={busy}
                className="rounded-2xl bg-cyan-400 px-4 py-2.5 text-sm font-semibold text-slate-950 transition hover:bg-cyan-300 disabled:cursor-not-allowed disabled:opacity-60 active:scale-95"
              >
                {busy ? 'Saving…' : 'Save DocType'}
              </button>
            </div>
          </div>
        </div>

        <div className="bg-slate-950/60 p-6">
          <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">Property panel</div>
          <h3 className="mt-2 text-xl font-semibold text-white">{selectedField ? selectedField.label || 'Selected field' : 'Select a field'}</h3>
          <p className="mt-2 text-sm leading-6 text-slate-400">
            Configure the active field and preview the exact JSON payload before it goes to the backend.
          </p>

          {selectedField ? (
            <div className="mt-5 space-y-4">
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
              <BaseSelect
                label="Field type"
                value={selectedField.fieldtype}
                onChange={(event) => updateField(selectedField.id, () => ({ fieldtype: event.target.value }))}
              >
                {fieldTypeOptions.map((option) => (
                  <option key={option} value={option}>{option}</option>
                ))}
              </BaseSelect>
              <BaseInput
                label="Default value"
                value={selectedField.default}
                onChange={(event) => updateField(selectedField.id, () => ({ default: event.target.value }))}
                placeholder={selectedField.fieldtype === 'Check' ? '0 or 1' : 'Optional default'}
              />
        {selectedField.fieldtype === 'JSON' ? (
        <BaseTextArea
          label="Options"
          value={selectedField.options}
          onChange={(event) => updateField(selectedField.id, () => ({ options: event.target.value }))}
          placeholder="JSON schema for child-table rendering"
          hint="For child tables, keep a JSON schema with mode=child-table and a columns array."
        />
        ) : (
        <BaseInput
          label="Options"
          value={selectedField.options}
          onChange={(event) => updateField(selectedField.id, () => ({ options: event.target.value }))}
          placeholder={selectedField.fieldtype === 'Link' ? 'Target DocType, e.g. customer' : 'Reserved for Select / Link evolution'}
          hint={selectedField.fieldtype === 'Link' ? 'Link fields use options as the target DocType name for live lookups.' : undefined}
        />
        )}

              <div className="grid gap-3">
                <ToggleCard
                  label="Mandatory"
                  description="Make the field required during document creation."
                  checked={selectedField.reqd}
                  onChange={(value) => updateField(selectedField.id, () => ({ reqd: value }))}
                />
                <ToggleCard
                  label="Read-only"
                  description="Render as a non-editable field in generated forms."
                  checked={selectedField.read_only}
                  onChange={(value) => updateField(selectedField.id, () => ({ read_only: value }))}
                />
                <ToggleCard
                  label="Unique"
                  description="Hint that values should remain unique in this DocType."
                  checked={selectedField.unique}
                  onChange={(value) => updateField(selectedField.id, () => ({ unique: value }))}
                />
                <ToggleCard
                  label="Hidden"
                  description="Keep the field in metadata while hiding it from normal entry screens."
                  checked={selectedField.hidden}
                  onChange={(value) => updateField(selectedField.id, () => ({ hidden: value }))}
                />
                <ToggleCard
                  label="List view"
                  description="Prioritize this field for future generated tables and desk list views."
                  checked={selectedField.in_list_view}
                  onChange={(value) => updateField(selectedField.id, () => ({ in_list_view: value }))}
                />
              </div>
            </div>
          ) : (
            <div className="mt-5 rounded-3xl border border-dashed border-white/10 bg-white/[0.03] px-4 py-5 text-sm leading-6 text-slate-400">
              Drop a field into the canvas, then select it to configure field properties like mandatory, read-only, unique, hidden, defaults, and future list behavior.
            </div>
          )}

          <div className="mt-6">
            <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">Payload preview</div>
            <pre className="mt-3 max-h-[360px] overflow-auto rounded-3xl border border-white/10 bg-slate-950 p-4 text-xs leading-6 text-cyan-200">
{JSON.stringify(payloadPreview, null, 2)}
            </pre>
          </div>
        </div>
      </form>
    </section>
  );
}