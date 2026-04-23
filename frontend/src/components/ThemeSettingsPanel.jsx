function FieldLabel({ title, hint }) {
  return (
    <div>
      <div className="text-sm font-semibold text-white">{title}</div>
      {hint ? <div className="mt-1 text-xs leading-5 text-slate-500">{hint}</div> : null}
    </div>
  );
}

export default function ThemeSettingsPanel({ open, theme, onChange, onClose, onReset }) {
  return (
    <>
      <div
        className={`fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm transition ${open ? 'opacity-100' : 'pointer-events-none opacity-0'}`}
        onClick={onClose}
        aria-hidden="true"
      />
      <aside
        className={`fixed right-0 top-0 z-50 flex h-screen w-full max-w-[420px] transform flex-col border-l border-white/10 bg-slate-950/96 shadow-2xl transition duration-300 ${
          open ? 'translate-x-0' : 'translate-x-full'
        }`}
        aria-hidden={!open}
      >
        <div className="flex items-start justify-between gap-4 border-b border-white/10 px-5 py-5">
          <div>
            <div className="badge">UI Studio settings</div>
            <h2 className="mt-3 text-xl font-semibold text-white">Theme & layout editor</h2>
            <p className="mt-2 text-sm leading-6 text-slate-400">
              Tune color, surface feel, text scale, sidebar width, and popup/dialogue styling without touching code.
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded-2xl border border-white/10 bg-white/[0.05] px-3 py-2 text-sm font-semibold text-slate-200 transition hover:bg-white/[0.1]"
          >
            Close
          </button>
        </div>

        <div className="flex-1 space-y-5 overflow-y-auto px-5 py-5">
          <section className="panel p-4">
            <div className="flex items-center justify-between gap-4">
              <FieldLabel title="Accent color" hint="Controls badges, focus rings, glows, and highlighted controls." />
              <input type="color" value={theme.accentColor} onChange={(event) => onChange('accentColor', event.target.value)} className="h-11 w-16 rounded-xl border border-white/10 bg-transparent" />
            </div>
          </section>

          <section className="panel p-4 space-y-4">
            <FieldLabel title="Surface tint" hint="Makes cards and panels feel lighter or denser." />
            <input type="range" min="0.03" max="0.16" step="0.01" value={theme.surfaceOpacity} onChange={(event) => onChange('surfaceOpacity', event.target.value)} className="w-full" />
            <div className="text-xs text-slate-400">Current opacity: {Number(theme.surfaceOpacity).toFixed(2)}</div>
          </section>

          <section className="panel p-4 space-y-4">
            <FieldLabel title="Text scale" hint="Makes the desk easier to read without zooming the whole browser." />
            <select className="field" value={theme.textScale} onChange={(event) => onChange('textScale', event.target.value)}>
              <option value="0.95">Compact</option>
              <option value="1">Default</option>
              <option value="1.05">Comfortable</option>
              <option value="1.1">Large</option>
            </select>
          </section>

          <section className="panel p-4 space-y-4">
            <FieldLabel title="Popup dialogue style" hint="Changes the roundness and feel of panels, overlays, and future modal dialogs." />
            <select className="field" value={theme.dialogStyle} onChange={(event) => onChange('dialogStyle', event.target.value)}>
              <option value="soft">Soft</option>
              <option value="modern">Modern</option>
              <option value="sharp">Sharp</option>
            </select>
          </section>

          <section className="panel p-4 space-y-4">
            <FieldLabel title="Sidebar width" hint="Set how wide the navigation drawer feels on larger screens." />
            <input type="range" min="280" max="420" step="4" value={theme.sidebarWidth} onChange={(event) => onChange('sidebarWidth', event.target.value)} className="w-full" />
            <div className="text-xs text-slate-400">Current width: {theme.sidebarWidth}px</div>
          </section>

          <section className="panel p-4 space-y-4">
            <FieldLabel title="Layout density" hint="Compact mode reduces visual breathing room for faster desk scanning." />
            <select className="field" value={theme.layoutDensity} onChange={(event) => onChange('layoutDensity', event.target.value)}>
              <option value="comfortable">Comfortable</option>
              <option value="compact">Compact</option>
            </select>
          </section>
        </div>

        <div className="border-t border-white/10 px-5 py-4">
          <div className="flex items-center justify-between gap-3">
            <button
              type="button"
              onClick={onReset}
              className="rounded-2xl border border-white/10 bg-white/[0.05] px-4 py-2.5 text-sm font-semibold text-slate-200 transition hover:bg-white/[0.1]"
            >
              Reset theme
            </button>
            <button
              type="button"
              onClick={onClose}
              className="rounded-2xl bg-cyan-400 px-4 py-2.5 text-sm font-semibold text-slate-950 transition hover:bg-cyan-300"
            >
              Done
            </button>
          </div>
        </div>
      </aside>
    </>
  );
}
