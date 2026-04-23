export default function DocTypeSidebar({ docTypes, loading, selectedName, onSelect, onRefresh, onCreate }) {
  return (
    <aside className="panel flex min-h-[70vh] flex-col overflow-hidden">
      <div className="border-b border-white/10 p-5">
        <div className="flex items-center justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold text-white">DocTypes</h2>
            <p className="mt-1 text-sm text-slate-400">Pick a model to inspect and edit live data.</p>
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={onCreate}
              className="rounded-xl bg-cyan-500 px-3 py-2 text-xs font-semibold text-slate-950 transition hover:bg-cyan-400"
            >
              New
            </button>
            <button
              type="button"
              onClick={onRefresh}
              className="rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-xs font-medium text-slate-200 transition hover:bg-white/10"
            >
              Refresh
            </button>
          </div>
        </div>
      </div>

      <div className="flex-1 space-y-2 overflow-y-auto p-3">
        {loading ? (
          <div className="rounded-xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-slate-400">Loading doctypes…</div>
        ) : null}

        {!loading && docTypes.length === 0 ? (
          <div className="rounded-xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-slate-400">
            No doctypes found yet.
          </div>
        ) : null}

        {docTypes.map((docType) => {
          const selected = docType.doctype === selectedName;
          return (
            <button
              key={docType.id}
              type="button"
              onClick={() => onSelect(docType.doctype)}
              className={`w-full rounded-2xl border px-4 py-3 text-left transition ${
                selected
                  ? 'border-cyan-400/40 bg-cyan-400/10 text-white shadow-lg shadow-cyan-900/20'
                  : 'border-white/10 bg-white/5 text-slate-300 hover:bg-white/10'
              }`}
            >
              <div className="flex items-center justify-between gap-3">
                <div>
                  <div className="font-medium">{docType.label || docType.doctype}</div>
                  <div className="mt-1 text-xs text-slate-400">{docType.doctype}</div>
                </div>
                {docType.is_system ? <span className="badge">system</span> : null}
              </div>
            </button>
          );
        })}
      </div>
    </aside>
  );
}
