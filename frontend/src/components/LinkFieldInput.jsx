import { useEffect, useMemo, useState } from 'react';
import { searchLinkOptions } from '../lib/api.js';
import { getLinkTargetDocType } from '../lib/metadata.js';

export default function LinkFieldInput({ field, value, onChange, disabled = false, placeholder }) {
	const targetDocType = useMemo(() => getLinkTargetDocType(field), [field]);
	const [search, setSearch] = useState(value || '');
	const [options, setOptions] = useState([]);
	const [loading, setLoading] = useState(false);
	const [open, setOpen] = useState(false);

	useEffect(() => {
		setSearch(value || '');
	}, [value]);

	useEffect(() => {
		if (!targetDocType || !open) {
			setOptions([]);
			return;
		}

		let active = true;
		setLoading(true);
		searchLinkOptions(targetDocType, {
			search,
			limit: 8,
		})
			.then((payload) => {
				if (!active) {
					return;
				}
				setOptions(payload);
			})
			.catch(() => {
				if (active) {
					setOptions([]);
				}
			})
			.finally(() => {
				if (active) {
					setLoading(false);
				}
			});

		return () => {
			active = false;
		};
	}, [targetDocType, search, open]);

	if (!targetDocType) {
		return (
			<div className="rounded-2xl border border-amber-400/20 bg-amber-500/10 px-4 py-3 text-sm text-amber-100">
				Add a target DocType in field <span className="font-mono">options</span> to enable Link suggestions.
			</div>
		);
	}

	return (
		<div className="space-y-3">
			<label className="grid gap-2">
				<input
					className="field"
					value={search}
					disabled={disabled}
					onFocus={() => setOpen(true)}
					onBlur={() => window.setTimeout(() => setOpen(false), 140)}
					onChange={(event) => {
						setSearch(event.target.value);
						onChange(event.target.value);
					}}
					placeholder={placeholder || `Search ${targetDocType} records`}
				/>
			</label>
			<div className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3 text-xs text-slate-500">
				Target DocType: <span className="font-mono text-cyan-200">{targetDocType}</span>
			</div>
			{open ? (
				<div className="rounded-2xl border border-white/10 bg-slate-950/90 p-2 shadow-xl shadow-slate-950/30">
					{loading ? (
						<div className="px-3 py-2 text-sm text-slate-400">Loading linked records…</div>
					) : options.length > 0 ? (
						options.map((option) => (
							<button
								key={option.name}
								type="button"
								onMouseDown={(event) => event.preventDefault()}
								onClick={() => {
									onChange(option.name);
									setSearch(option.name);
									setOpen(false);
								}}
								className="flex w-full items-center justify-between rounded-xl px-3 py-2 text-left text-sm text-slate-200 transition hover:bg-white/[0.06]"
							>
								<div>
									<div>{option.label || option.name}</div>
									{option.description ? <div className="mt-1 text-xs text-slate-500">{option.description}</div> : null}
								</div>
								<span className="text-xs text-slate-500 font-mono">{option.name}</span>
							</button>
						))
					) : (
						<div className="px-3 py-2 text-sm text-slate-500">No linked records found. You can still type a manual value.</div>
					)}
				</div>
			) : null}
		</div>
	);
}