import LinkFieldInput from './LinkFieldInput.jsx';
import { coerceChildTableValue, createChildTableRow, fieldInputType, getChildTableConfig, normalizeChildTableRows } from '../lib/metadata.js';

function ChildCellInput({ column, value, onChange }) {
	if (column.fieldtype === 'Check') {
		return (
			<button
				type="button"
				onClick={() => onChange(!value)}
				className={`flex h-11 items-center justify-between rounded-xl border px-3 text-sm transition ${
					value ? 'border-emerald-400/30 bg-emerald-500/10 text-emerald-100' : 'border-white/10 bg-slate-950/70 text-slate-300'
				}`}
			>
				<span>{value ? 'Yes' : 'No'}</span>
				<span>{value ? '✓' : '○'}</span>
			</button>
		);
	}

	if (column.fieldtype === 'Link') {
		return <LinkFieldInput field={column} value={value} onChange={onChange} placeholder={`Link ${column.label.toLowerCase()}`} />;
	}

	if (column.fieldtype === 'Text' || column.fieldtype === 'Long Text' || column.fieldtype === 'Small Text') {
		return (
			<textarea
				className="field min-h-24"
				value={value ?? ''}
				onChange={(event) => onChange(event.target.value)}
				placeholder={`Enter ${column.label.toLowerCase()}`}
			/>
		);
	}

	return (
		<input
			className="field"
			type={fieldInputType(column)}
			step={column.fieldtype === 'Int' ? '1' : 'any'}
			value={value ?? ''}
			onChange={(event) => onChange(event.target.value)}
			placeholder={`Enter ${column.label.toLowerCase()}`}
		/>
	);
}

export default function ChildTableField({ field, value, onChange }) {
	const config = getChildTableConfig(field);
	const rows = coerceChildTableValue(field, value);

	if (!config) {
		return (
			<div className="rounded-2xl border border-amber-400/20 bg-amber-500/10 px-4 py-3 text-sm text-amber-100">
				This JSON field is marked for child-table rendering, but its <span className="font-mono">options</span> schema is missing or invalid.
			</div>
		);
	}

	const updateRows = (nextRows) => {
		onChange(normalizeChildTableRows(config.columns, nextRows));
	};

	const handleCellChange = (rowIndex, fieldname, nextValue) => {
		const nextRows = rows.map((row, index) => (index === rowIndex ? { ...row, [fieldname]: nextValue } : row));
		updateRows(nextRows);
	};

	const addRow = () => {
		updateRows([...rows, createChildTableRow(config.columns)]);
	};

	const removeRow = (rowIndex) => {
		updateRows(rows.filter((_, index) => index !== rowIndex));
	};

	return (
		<div className="rounded-3xl border border-white/10 bg-white/[0.03] p-4">
			<div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
				<div>
					<div className="text-sm font-semibold text-white">{config.label || field.label}</div>
					<p className="mt-1 text-xs leading-5 text-slate-500">JSON-backed child rows for line items, schedule rows, checklist steps, or any repeated nested data.</p>
				</div>
				<button
					type="button"
					onClick={addRow}
					className="rounded-xl border border-cyan-400/20 bg-cyan-500/10 px-3 py-2 text-sm font-semibold text-cyan-100 transition hover:bg-cyan-500/20"
				>
					Add row
				</button>
			</div>

			<div className="mt-4 space-y-4">
				{rows.length > 0 ? (
					rows.map((row, rowIndex) => (
						<div key={`${field.fieldname}-row-${rowIndex}`} className="rounded-3xl border border-white/10 bg-slate-950/60 p-4">
							<div className="mb-4 flex items-center justify-between gap-3">
								<div className="text-sm font-semibold text-white">Row {rowIndex + 1}</div>
								<button
									type="button"
									onClick={() => removeRow(rowIndex)}
									className="rounded-full border border-rose-400/20 bg-rose-500/10 px-3 py-1.5 text-xs font-semibold text-rose-100 transition hover:bg-rose-500/20"
								>
									Remove
								</button>
							</div>
							<div className="grid gap-4 xl:grid-cols-2">
								{config.columns.map((column) => (
									<label key={column.fieldname} className="grid gap-2">
										<span className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">{column.label}</span>
										<ChildCellInput column={column} value={row[column.fieldname]} onChange={(nextValue) => handleCellChange(rowIndex, column.fieldname, nextValue)} />
									</label>
								))}
							</div>
						</div>
					))
				) : (
					<div className="rounded-2xl border border-dashed border-white/10 bg-white/[0.03] px-4 py-6 text-center text-sm text-slate-400">
						No child rows yet. Add the first one to start building nested business data.
					</div>
				)}
			</div>
		</div>
	);
}