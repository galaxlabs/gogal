export const supportedFieldTypes = new Set([
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
]);

const childTableFallback = {
	mode: 'child-table',
	label: 'Line Items',
	columns: [
		{ fieldname: 'item_code', label: 'Item Code', fieldtype: 'Data', reqd: true },
		{ fieldname: 'qty', label: 'Qty', fieldtype: 'Int' },
		{ fieldname: 'is_active', label: 'Active', fieldtype: 'Check' },
	],
};

function safeParseJSON(raw) {
	if (typeof raw !== 'string') {
		return null;
	}

	try {
		return JSON.parse(raw);
	} catch {
		return null;
	}
}

export function childTableOptionsTemplate() {
	return JSON.stringify(childTableFallback, null, 2);
}

export function getFieldOptionsObject(field) {
	return safeParseJSON(field?.options || '');
}

export function getLinkTargetDocType(field) {
	const parsed = getFieldOptionsObject(field);
	if (parsed && typeof parsed.target_doctype === 'string') {
		return parsed.target_doctype.trim();
	}

	return String(field?.options || '').trim();
}

export function getSelectChoices(field) {
	const parsed = getFieldOptionsObject(field);
	if (Array.isArray(parsed)) {
		return parsed.map(String);
	}
	if (parsed && Array.isArray(parsed.options)) {
		return parsed.options.map(String);
	}

	return String(field?.options || '')
		.split(/\n|,/)
		.map((item) => item.trim())
		.filter(Boolean);
}

export function getChildTableConfig(field) {
	if (!field || field.fieldtype !== 'JSON') {
		return null;
	}

	const parsed = getFieldOptionsObject(field);
	if (!parsed || parsed.mode !== 'child-table' || !Array.isArray(parsed.columns)) {
		return null;
	}

	return {
		label: parsed.label || field.label,
		columns: parsed.columns.map((column, index) => ({
			fieldname: column.fieldname || `column_${index + 1}`,
			label: column.label || column.fieldname || `Column ${index + 1}`,
			fieldtype: column.fieldtype || 'Data',
			reqd: Boolean(column.reqd),
			options: column.options || '',
		})),
	};
}

export function isChildTableField(field) {
	return Boolean(getChildTableConfig(field));
}

export function isEditableField(field) {
	return supportedFieldTypes.has(field.fieldtype) && !field.hidden;
}

export function defaultValueForField(field) {
	if (isChildTableField(field)) {
		return [];
	}

	if (field.default !== undefined && field.default !== null && field.default !== '') {
		if (field.fieldtype === 'Check') {
			return field.default === true || field.default === '1' || String(field.default).toLowerCase() === 'true';
		}

		return toFormValue(field, field.default);
	}

	if (field.fieldtype === 'Check') {
		return false;
	}

	return '';
}

export function fieldInputType(field) {
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

function toDatetimeLocalValue(value) {
	if (!value) {
		return '';
	}

	const date = new Date(value);
	if (Number.isNaN(date.getTime())) {
		return String(value);
	}

	const offsetDate = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
	return offsetDate.toISOString().slice(0, 16);
}

export function toFormValue(field, value) {
	if (value === undefined || value === null) {
		return defaultValueForField({ ...field, default: '' });
	}

	if (field.fieldtype === 'Check') {
		return Boolean(value);
	}

	if (isChildTableField(field)) {
		return coerceChildTableValue(field, value);
	}

	if (field.fieldtype === 'JSON') {
		return typeof value === 'string' ? value : JSON.stringify(value, null, 2);
	}

	if (field.fieldtype === 'Date') {
		return String(value).slice(0, 10);
	}

	if (field.fieldtype === 'Datetime') {
		return toDatetimeLocalValue(value);
	}

	if (field.fieldtype === 'Time') {
		return String(value).slice(0, 8);
	}

	return String(value);
}

export function buildInitialValues(docType, initialValues = {}) {
	return Object.fromEntries(
		(docType?.fields || [])
			.filter((field) => isEditableField(field) && !field.read_only)
			.map((field) => [field.fieldname, toFormValue(field, initialValues[field.fieldname])]),
	);
}

export function createChildTableRow(columns) {
	return Object.fromEntries(columns.map((column) => [column.fieldname, defaultValueForField(column)]));
}

export function coerceChildTableValue(field, value) {
	const config = getChildTableConfig(field);
	if (!config) {
		return [];
	}

	const rawRows = typeof value === 'string' ? safeParseJSON(value) : value;
	if (!Array.isArray(rawRows)) {
		return [];
	}

	return rawRows.map((row) => {
		const source = row && typeof row === 'object' ? row : {};
		return Object.fromEntries(
			config.columns.map((column) => [column.fieldname, source[column.fieldname] ?? defaultValueForField(column)]),
		);
	});
}

export function normalizeChildTableRows(columns, rows) {
	return rows.map((row) => {
		const normalized = {};
		columns.forEach((column) => {
			const rawValue = row?.[column.fieldname] ?? defaultValueForField(column);
			normalized[column.fieldname] = normalizeSubmissionValue(column, rawValue);
		});
		return normalized;
	});
}

export function normalizeSubmissionValue(field, value) {
	if (field.fieldtype === 'Check') {
		return Boolean(value);
	}

	if (isChildTableField(field)) {
		const config = getChildTableConfig(field);
		return normalizeChildTableRows(config.columns, Array.isArray(value) ? value : []);
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

export function formatFieldValue(field, value) {
	if (value === undefined || value === null || value === '') {
		return '—';
	}

	if (isChildTableField(field)) {
		const rows = coerceChildTableValue(field, value);
		return `${rows.length} row${rows.length === 1 ? '' : 's'}`;
	}

	if (field.fieldtype === 'Check') {
		return value ? 'Enabled' : 'Disabled';
	}

	if (field.fieldtype === 'JSON') {
		return typeof value === 'string' ? value : JSON.stringify(value, null, 2);
	}

	if (field.fieldtype === 'Datetime') {
		const date = new Date(value);
		return Number.isNaN(date.getTime()) ? String(value) : new Intl.DateTimeFormat('en-US', { dateStyle: 'medium', timeStyle: 'short' }).format(date);
	}

	if (field.fieldtype === 'Date') {
		const date = new Date(value);
		return Number.isNaN(date.getTime()) ? String(value) : new Intl.DateTimeFormat('en-US', { dateStyle: 'medium' }).format(date);
	}

	return String(value);
}