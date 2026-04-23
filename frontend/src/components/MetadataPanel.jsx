const queryExamples = [
  '?search=alpha',
  '?sort_by=priority&sort_order=desc',
  '?filter_priority__gte=2',
  '?filter_is_done=true',
  '?filter_priority__in=1,2&search=task',
];

export default function MetadataPanel({ docType, loading }) {
  if (loading) {
    return <aside className="panel p-6 text-sm text-slate-300">Loading metadata…</aside>;
  }

  if (!docType) {
    return (
      <aside className="panel p-6 text-sm text-slate-300">
        Select a DocType to view metadata, supported filters, and live API help.
      </aside>
    );
  }

  return (
    <aside className="panel flex min-h-[70vh] flex-col overflow-hidden">
      <div className="border-b border-white/10 p-5">
        <h2 className="text-lg font-semibold text-white">Metadata & help</h2>
        <p className="mt-1 text-sm text-slate-400">Everything the frontend needs to render and query {docType.doctype}.</p>
      </div>

      <div className="flex-1 space-y-5 overflow-y-auto p-5 text-sm text-slate-300">
        <section>
          <div className="mb-2 flex flex-wrap items-center gap-2">
            <span className="badge">{docType.module}</span>
            <span className="badge">{docType.table_name}</span>
            {docType.is_system ? <span className="badge">system model</span> : null}
          </div>
          <p className="text-slate-400">{docType.description || 'No description yet. This doctype is ready for live record operations.'}</p>
        </section>

        <section>
          <h3 className="mb-3 text-sm font-semibold uppercase tracking-[0.18em] text-slate-400">Fields</h3>
          <div className="space-y-2">
            {(docType.fields || []).map((field) => (
              <div key={field.id ?? field.fieldname} className="rounded-2xl border border-white/10 bg-white/5 p-3">
                <div className="flex items-center justify-between gap-3">
                  <div>
                    <div className="font-medium text-white">{field.label}</div>
                    <div className="mt-1 font-mono text-xs text-slate-400">{field.fieldname}</div>
                  </div>
                  <span className="badge">{field.fieldtype}</span>
                </div>
                <div className="mt-3 flex flex-wrap gap-2 text-xs text-slate-300">
                  {field.reqd ? <span className="rounded-full bg-amber-400/10 px-2 py-1 text-amber-200">required</span> : null}
                  {field.unique ? <span className="rounded-full bg-fuchsia-400/10 px-2 py-1 text-fuchsia-200">unique</span> : null}
                  {field.read_only ? <span className="rounded-full bg-slate-400/10 px-2 py-1 text-slate-200">read only</span> : null}
                  {field.hidden ? <span className="rounded-full bg-slate-400/10 px-2 py-1 text-slate-200">hidden</span> : null}
                </div>
              </div>
            ))}
          </div>
        </section>

        <section>
          <h3 className="mb-3 text-sm font-semibold uppercase tracking-[0.18em] text-slate-400">Live query help</h3>
          <div className="rounded-2xl border border-white/10 bg-slate-950/70 p-4">
            <p className="text-slate-400">Try these directly in the browser or in the table controls:</p>
            <ul className="mt-3 space-y-2 font-mono text-xs text-cyan-200">
              {queryExamples.map((example) => (
                <li key={example}>/api/resource/{docType.doctype}{example}</li>
              ))}
            </ul>
          </div>
        </section>
      </div>
    </aside>
  );
}
