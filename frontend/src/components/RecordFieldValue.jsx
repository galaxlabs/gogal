import { coerceChildTableValue, formatFieldValue, getChildTableConfig } from '../lib/metadata.js';

export default function RecordFieldValue({ field, value }) {
	const childTableConfig = getChildTableConfig(field);

	if (childTableConfig) {
		const rows = coerceChildTableValue(field, value);
		return rows.length > 0 ? (
			<div className="overflow-x-auto rounded-2xl border border-white/10 bg-slate-950/50">
				<table className="min-w-full divide-y divide-white/10 text-sm">
					<thead className="bg-white/[0.04] text-left text-slate-400">
						<tr>
							{childTableConfig.columns.map((column) => (
								<th key={column.fieldname} className="px-3 py-2 font-medium uppercase tracking-[0.16em]">{column.label}</th>
							))}
						</tr>
					</thead>
					<tbody className="divide-y divide-white/5">
						{rows.map((row, rowIndex) => (
							<tr key={`${field.fieldname}-preview-${rowIndex}`}>
								{childTableConfig.columns.map((column) => (
									<td key={`${column.fieldname}-${rowIndex}`} className="px-3 py-2 text-slate-200">
										{formatFieldValue(column, row[column.fieldname])}
									</td>
								))}
							</tr>
						))}
					</tbody>
				</table>
			</div>
		) : (
			<span className="text-slate-500">No child rows</span>
		);
	}

	if (field.fieldtype === 'Link' || field.fieldtype === 'DynamicLink') {
		return value ? <span className="inline-flex rounded-full border border-cyan-400/20 bg-cyan-500/10 px-3 py-1 text-sm text-cyan-100">{String(value)}</span> : '—';
	}

	if (field.fieldtype === 'JSON') {
		return <pre className="whitespace-pre-wrap text-xs leading-6 text-cyan-200">{formatFieldValue(field, value)}</pre>;
	}

	return <span>{formatFieldValue(field, value)}</span>;
}