import { useEffect, useMemo, useState } from 'react';
import { fetchDocTypes, fetchDocTypeMeta } from './lib/api.js';
import DocTypeBuilder from './components/DocTypeBuilder.jsx';
import ResourceWorkbench from './components/ResourceWorkbench.jsx';
import ThemeSettingsPanel from './components/ThemeSettingsPanel.jsx';

const STUDIO_THEME_STORAGE_KEY = 'gogal-ui-studio-theme';

const defaultStudioTheme = {
  accentColor: '#22d3ee',
  surfaceOpacity: '0.05',
  textScale: '1',
  dialogStyle: 'soft',
  sidebarWidth: '320',
  layoutDensity: 'comfortable',
};

const platformModules = [
  {
    id: 'builder',
    title: 'DocType Builder',
    description: 'Visual schema design with drag and drop, touch-friendly controls, and metadata-first saves.',
    badge: 'Live',
  },
  {
    id: 'automation',
    title: 'Automation Tasks',
    description: 'Queue jobs, event hooks, and workflow actions that will run across every installed app.',
    badge: 'Planned',
  },
  {
    id: 'scheduler',
    title: 'Scheduler',
    description: 'Timezone-aware schedules, cron orchestration, and recurring desk tasks.',
    badge: 'Planned',
  },
  {
    id: 'codes',
    title: 'QR & Barcode',
    description: 'Generate scannable labels, inventory cards, and print-ready machine-readable assets.',
    badge: 'Planned',
  },
  {
    id: 'time',
    title: 'Date & Timezone Lab',
    description: 'Convert, preview, and validate local versus UTC time before workflows ever go live.',
    badge: 'Planned',
  },
];

const roadmapCards = [
  {
    title: 'Print & Reporting UI Studio',
    detail: 'Letterheads, invoices, charts, dashboards, and export-ready report surfaces.',
  },
  {
    title: 'Client & Server Scripts',
    detail: 'Monaco-powered editing for UI event logic and backend hook orchestration.',
  },
  {
    title: 'Version History',
    detail: 'Per-record audit trail, timeline diffs, and who-changed-what accountability.',
  },
];

function formatTime(date, options) {
  return new Intl.DateTimeFormat('en-US', options).format(date);
}

function hexToRgbChannels(hex) {
  const normalized = hex.replace('#', '').trim();
  if (!/^[0-9a-fA-F]{6}$/.test(normalized)) {
    return '34 211 238';
  }

  const r = Number.parseInt(normalized.slice(0, 2), 16);
  const g = Number.parseInt(normalized.slice(2, 4), 16);
  const b = Number.parseInt(normalized.slice(4, 6), 16);
  return `${r} ${g} ${b}`;
}

function resolveDialogRadius(dialogStyle) {
  switch (dialogStyle) {
    case 'sharp':
      return '0.75rem';
    case 'modern':
      return '1.25rem';
    default:
      return '1.6rem';
  }
}

function readStoredStudioTheme() {
  if (typeof window === 'undefined') {
    return defaultStudioTheme;
  }

  try {
    const raw = window.localStorage.getItem(STUDIO_THEME_STORAGE_KEY);
    if (!raw) {
      return defaultStudioTheme;
    }

    return {
      ...defaultStudioTheme,
      ...JSON.parse(raw),
    };
  } catch {
    return defaultStudioTheme;
  }
}

function SidebarSkeleton() {
  return (
    <div className="space-y-3">
      {Array.from({ length: 5 }).map((_, index) => (
        <div key={index} className="skeleton-line h-12 rounded-2xl" />
      ))}
    </div>
  );
}

function NavButton({ title, description, badge, active = false, onClick }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`group w-full rounded-2xl border px-4 py-3 text-left transition duration-200 hover:-translate-y-0.5 active:scale-[0.99] ${
        active
          ? 'border-cyan-400/40 bg-cyan-500/10 shadow-lg shadow-cyan-500/10'
          : 'border-white/8 bg-white/[0.04] hover:border-white/15 hover:bg-white/[0.06]'
      }`}
    >
      <div className="flex items-start justify-between gap-3">
        <div>
          <div className="text-sm font-semibold text-white">{title}</div>
          <div className="mt-1 text-xs leading-5 text-slate-400">{description}</div>
        </div>
        <span className={`badge ${active ? 'border-cyan-300/30 bg-cyan-400/15 text-cyan-100' : ''}`}>{badge}</span>
      </div>
    </button>
  );
}

function RecentDocTypeButton({ docType, active, onClick }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`w-full rounded-2xl border px-4 py-3 text-left transition hover:border-white/15 hover:bg-white/[0.06] active:scale-[0.99] ${
        active ? 'border-emerald-400/30 bg-emerald-500/10' : 'border-white/8 bg-white/[0.03]'
      }`}
    >
      <div className="flex items-center justify-between gap-3">
        <div>
          <div className="text-sm font-semibold text-white">{docType.label || docType.doctype}</div>
          <div className="mt-1 text-xs uppercase tracking-[0.2em] text-slate-500">{docType.module || 'Core'}</div>
        </div>
        <div className="rounded-full border border-white/10 px-2 py-1 text-[11px] text-slate-300">{docType.doctype}</div>
      </div>
    </button>
  );
}

function StatusCard({ title, children, accent = 'cyan' }) {
  const accentClasses = {
    cyan: 'border-cyan-400/20 bg-cyan-500/10 text-cyan-100',
    emerald: 'border-emerald-400/20 bg-emerald-500/10 text-emerald-100',
    violet: 'border-violet-400/20 bg-violet-500/10 text-violet-100',
  };

  return (
    <section className="panel p-5">
      <div className={`inline-flex rounded-full border px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.24em] ${accentClasses[accent]}`}>
        {title}
      </div>
      <div className="mt-4">{children}</div>
    </section>
  );
}

export default function App() {
  const [docTypes, setDocTypes] = useState([]);
  const [selectedName, setSelectedName] = useState('');
  const [selectedMeta, setSelectedMeta] = useState(null);
  const [loading, setLoading] = useState(true);
  const [metaLoading, setMetaLoading] = useState(false);
  const [error, setError] = useState('');
  const [search, setSearch] = useState('');
  const [activeModule, setActiveModule] = useState('builder');
  const [now, setNow] = useState(() => new Date());
	const [navigationOpen, setNavigationOpen] = useState(false);
	const [themeEditorOpen, setThemeEditorOpen] = useState(false);
	const [studioTheme, setStudioTheme] = useState(() => readStoredStudioTheme());

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError('');

    fetchDocTypes()
      .then((items) => {
        if (!active) {
          return;
        }
        setDocTypes(items);
        const preferred = items.find((item) => !item.is_system)?.doctype || items[0]?.doctype || '';
        setSelectedName((current) => current || preferred);
      })
      .catch((requestError) => {
        if (active) {
          setError(requestError.message);
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
  }, []);

  useEffect(() => {
    const timer = window.setInterval(() => setNow(new Date()), 60_000);
    return () => window.clearInterval(timer);
  }, []);

	useEffect(() => {
		window.localStorage.setItem(STUDIO_THEME_STORAGE_KEY, JSON.stringify(studioTheme));
	}, [studioTheme]);

  useEffect(() => {
    if (!selectedName) {
      setSelectedMeta(null);
      return;
    }

    let active = true;
    setMetaLoading(true);
    fetchDocTypeMeta(selectedName)
      .then((meta) => {
        if (active) {
          setSelectedMeta(meta);
        }
      })
      .catch((requestError) => {
        if (active) {
          setError(requestError.message);
        }
      })
      .finally(() => {
        if (active) {
          setMetaLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, [selectedName]);

  useEffect(() => {
    if (!selectedName) {
      return;
    }
    setNavigationOpen(false);
  }, [selectedName]);

    const tenantName = import.meta.env.VITE_TENANT_NAME || 'Gogal';
  const timezone = useMemo(() => Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC', []);

  const filteredDocTypes = useMemo(() => {
    const query = search.trim().toLowerCase();
    if (!query) {
      return docTypes;
    }

    return docTypes.filter((item) => {
      const haystack = [item.doctype, item.label, item.module].filter(Boolean).join(' ').toLowerCase();
      return haystack.includes(query);
    });
  }, [docTypes, search]);

  const installedModules = useMemo(() => {
    const seen = new Set();
    return filteredDocTypes
      .map((docType) => docType.module || 'Core')
      .filter((moduleName) => {
        const normalized = moduleName.toLowerCase();
        if (seen.has(normalized)) {
          return false;
        }
        seen.add(normalized);
        return true;
      })
      .map((moduleName) => ({
        id: moduleName.toLowerCase(),
        title: moduleName,
        description: `Live module discovered from DocTypes assigned to ${moduleName}.`,
        badge: 'Installed',
      }));
  }, [filteredDocTypes]);

  const highlightedDocTypes = useMemo(() => filteredDocTypes.slice(0, 8), [filteredDocTypes]);

  const localTime = useMemo(
    () => formatTime(now, { dateStyle: 'medium', timeStyle: 'short', timeZone: timezone }),
    [now, timezone],
  );
  const utcTime = useMemo(
    () => formatTime(now, { dateStyle: 'medium', timeStyle: 'short', timeZone: 'UTC' }),
    [now],
  );

  const handleCreated = (created) => {
    const createdName = created?.doctype || created?.name || '';
    if (!createdName) {
      return;
    }

    setDocTypes((current) => {
      const summary = {
        doctype: createdName,
        label: created?.label || createdName,
        module: created?.module || 'Core',
        is_system: Boolean(created?.is_system),
      };

      const remaining = current.filter((item) => item.doctype !== createdName);
      return [summary, ...remaining];
    });
    setSelectedName(createdName);
    if (created?.fields) {
      setSelectedMeta(created);
    }
  };

  const themeVars = useMemo(() => ({
    '--studio-accent-rgb': hexToRgbChannels(studioTheme.accentColor),
    '--studio-surface-opacity': studioTheme.surfaceOpacity,
    '--studio-radius': resolveDialogRadius(studioTheme.dialogStyle),
    '--studio-text-scale': studioTheme.textScale,
    '--studio-sidebar-width': `${studioTheme.sidebarWidth}px`,
  }), [studioTheme]);

  const updateStudioTheme = (key, value) => {
    setStudioTheme((current) => ({ ...current, [key]: value }));
  };

  const resetStudioTheme = () => {
    setStudioTheme(defaultStudioTheme);
  };

  return (
    <div style={themeVars} className={`min-h-screen text-slate-100 ${studioTheme.layoutDensity === 'compact' ? 'studio-density-compact' : ''}`}>
      <div className="mx-auto flex min-h-screen max-w-[1800px] flex-col lg:flex-row">
			<div
				className={`fixed inset-0 z-30 bg-slate-950/60 backdrop-blur-sm transition lg:hidden ${navigationOpen ? 'opacity-100' : 'pointer-events-none opacity-0'}`}
				onClick={() => setNavigationOpen(false)}
				aria-hidden="true"
			/>
			<aside
				style={{ width: `min(88vw, ${studioTheme.sidebarWidth}px)` }}
				className={`studio-sidebar fixed inset-y-0 left-0 z-40 overflow-y-auto border-r border-white/8 px-4 py-5 transition duration-300 lg:static lg:min-h-screen lg:translate-x-0 lg:px-5 lg:py-6 ${navigationOpen ? 'translate-x-0' : '-translate-x-full'}`}
			>
          <div className="panel relative overflow-hidden p-5">
            <div className="studio-glow" />
            <div className="relative flex items-start gap-4">
              <div className="flex h-14 w-14 items-center justify-center rounded-2xl border border-cyan-300/20 bg-cyan-400/10 text-lg font-semibold text-cyan-100 shadow-lg shadow-cyan-500/10">
                GG
              </div>
              <div>
                <div className="text-xs font-semibold uppercase tracking-[0.28em] text-cyan-200/80">UI Studio</div>
                <h1 className="mt-2 text-2xl font-semibold text-white">{tenantName}</h1>
                <p className="mt-2 text-sm leading-6 text-slate-400">
                  Headless app builder, admin desk, automation cockpit, and metadata-first control center for Gogal.
                </p>
              </div>
            </div>
        <button
          type="button"
          onClick={() => setNavigationOpen(false)}
          className="mt-4 rounded-2xl border border-white/10 bg-white/[0.05] px-3 py-2 text-sm font-semibold text-slate-200 transition hover:bg-white/[0.1] lg:hidden"
        >
          Close drawer
        </button>
          </div>

          <div className="mt-5 space-y-5">
            <div>
              <div className="mb-3 flex items-center justify-between">
                <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">UI Studio modules</div>
                <span className="badge">{platformModules.length + installedModules.length} total</span>
              </div>
              <div className="space-y-3">
                {platformModules.map((moduleItem) => (
                  <NavButton
                    key={moduleItem.id}
                    title={moduleItem.title}
                    description={moduleItem.description}
                    badge={moduleItem.badge}
                    active={activeModule === moduleItem.id}
                    onClick={() => setActiveModule(moduleItem.id)}
                  />
                ))}
              </div>
            </div>

            <div>
              <div className="mb-3 text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">Installed apps & modules</div>
              <div className="space-y-3">
                {loading ? (
                  <SidebarSkeleton />
                ) : installedModules.length > 0 ? (
                  installedModules.map((moduleItem) => (
                    <NavButton
                      key={moduleItem.id}
                      title={moduleItem.title}
                      description={moduleItem.description}
                      badge={moduleItem.badge}
                    />
                  ))
                ) : (
                  <div className="rounded-2xl border border-dashed border-white/10 bg-white/[0.03] px-4 py-5 text-sm text-slate-400">
                    Installed modules will appear here automatically once your bench starts adding DocTypes by module.
                  </div>
                )}
              </div>
            </div>

            <div>
              <div className="mb-3 text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">Recent DocTypes</div>
              <div className="space-y-3">
                {loading ? (
                  <SidebarSkeleton />
                ) : highlightedDocTypes.length > 0 ? (
                  highlightedDocTypes.map((docType) => (
                    <RecentDocTypeButton
                      key={docType.doctype}
                      docType={docType}
                      active={selectedName === docType.doctype}
                      onClick={() => setSelectedName(docType.doctype)}
                    />
                  ))
                ) : (
                  <div className="rounded-2xl border border-dashed border-white/10 bg-white/[0.03] px-4 py-5 text-sm text-slate-400">
                    No DocTypes match your current search. Create one from the builder to see it show up here instantly.
                  </div>
                )}
              </div>
            </div>
          </div>
			</aside>

        <div className="flex-1">
          <header className="sticky top-0 z-20 border-b border-white/8 bg-slate-950/75 px-4 py-4 backdrop-blur-xl lg:px-6">
            <div className="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
              <div>
                <div className="flex flex-wrap items-center gap-2">
                  <button
						type="button"
						onClick={() => setNavigationOpen(true)}
						className="rounded-full border border-white/10 bg-white/[0.05] px-3 py-1.5 text-xs font-semibold text-slate-100 transition hover:bg-white/[0.1] lg:hidden"
					>
						Drawer
					</button>
                  <span className="badge">Go backend · /api</span>
                  <span className="badge">Tailwind shell</span>
                  <span className="badge">Touch friendly</span>
                  <span className="badge">Animated surfaces</span>
                </div>
                <h2 className="mt-3 text-3xl font-semibold tracking-tight text-white">UI Studio — App Builder & Admin Desk</h2>
                <p className="mt-2 max-w-3xl text-sm text-slate-400">
                  Build metadata visually, prepare automation-ready modules, and shape an enterprise desk that can outgrow ERP-era admin UX.
                </p>
              </div>

              <div className="flex flex-col gap-3 md:flex-row md:items-center">
                <label className="relative block min-w-[min(100%,28rem)]">
                  <span className="pointer-events-none absolute inset-y-0 left-4 flex items-center text-slate-500">⌕</span>
                  <input
                    className="field h-12 pl-10 pr-4"
                    value={search}
                    onChange={(event) => setSearch(event.target.value)}
                    placeholder="Search DocTypes, modules, or builder surfaces…"
                  />
                </label>

                <div className="flex items-center gap-3">
                  <div className="rounded-2xl border border-white/10 bg-white/[0.04] px-4 py-3 text-right">
                    <div className="text-xs font-semibold uppercase tracking-[0.24em] text-slate-500">Local / UTC</div>
                    <div className="mt-1 text-sm text-slate-200">{localTime}</div>
                    <div className="text-xs text-slate-500">UTC: {utcTime}</div>
                  </div>
                  <div className="rounded-2xl border border-cyan-300/15 bg-cyan-400/10 px-4 py-3">
                    <div className="text-xs font-semibold uppercase tracking-[0.24em] text-cyan-100/70">Timezone</div>
                    <div className="mt-1 text-sm text-cyan-50">{timezone}</div>
                  </div>
                  <button
						type="button"
						onClick={() => setThemeEditorOpen(true)}
						className="rounded-2xl border border-white/10 bg-white/[0.05] px-4 py-3 text-sm font-semibold text-white transition hover:-translate-y-0.5 hover:bg-white/[0.08] active:scale-95"
					>
						Theme
					</button>
                  <button
                    type="button"
                    className="flex h-12 w-12 items-center justify-center rounded-2xl border border-white/10 bg-white/[0.05] text-sm font-semibold text-white transition hover:-translate-y-0.5 hover:bg-white/[0.08] active:scale-95"
                    aria-label="Open profile menu"
                  >
                    UX
                  </button>
                </div>
              </div>
            </div>
          </header>

          <main className="grid gap-6 px-4 py-6 lg:px-6 xl:grid-cols-[minmax(0,1.45fr)_360px]">
            <section className="space-y-6">
              {error ? (
                <div className="rounded-2xl border border-rose-400/20 bg-rose-500/10 px-4 py-3 text-sm text-rose-100">
                  {error}
                </div>
              ) : null}

              <div className="grid gap-4 md:grid-cols-3">
                <StatusCard title="UI Studio readiness" accent="emerald">
                  <div className="text-3xl font-semibold text-white">{docTypes.length}</div>
                  <p className="mt-2 text-sm leading-6 text-slate-300">DocTypes available to drive metadata-first forms, lists, and desk navigation.</p>
                </StatusCard>

                <StatusCard title="Installed modules" accent="cyan">
                  <div className="text-3xl font-semibold text-white">{installedModules.length || 1}</div>
                  <p className="mt-2 text-sm leading-6 text-slate-300">The sidebar is already dynamic — it discovers modules directly from your live metadata.</p>
                </StatusCard>

                <StatusCard title="Builder focus" accent="violet">
                  <div className="text-lg font-semibold text-white">{activeModule === 'builder' ? 'Visual schema design' : 'Future studio surface'}</div>
                  <p className="mt-2 text-sm leading-6 text-slate-300">Today’s milestone centers the DocType Builder while previewing the modules you requested next.</p>
                </StatusCard>
              </div>

              <DocTypeBuilder
                moduleOptions={installedModules.map((item) => item.title)}
                existingDocTypes={docTypes.map((item) => item.doctype)}
                onCreated={handleCreated}
              />

				<ResourceWorkbench docType={selectedMeta} loading={metaLoading} />
            </section>

            <aside className="space-y-6">
              <section className="panel p-5">
                <div className="flex items-center justify-between gap-3">
                  <div>
                    <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">Selected metadata</div>
                    <h3 className="mt-2 text-xl font-semibold text-white">
                      {selectedMeta?.label || selectedName || 'Nothing selected yet'}
                    </h3>
                  </div>
                  <span className="badge">{metaLoading ? 'Loading' : selectedMeta ? `${selectedMeta.fields?.length || 0} fields` : 'Ready'}</span>
                </div>

                {metaLoading ? (
                  <div className="mt-4 space-y-3">
                    <div className="skeleton-line h-20 rounded-2xl" />
                    <div className="skeleton-line h-14 rounded-2xl" />
                    <div className="skeleton-line h-14 rounded-2xl" />
                  </div>
                ) : selectedMeta ? (
                  <div className="mt-4 space-y-3">
                    <p className="text-sm leading-6 text-slate-400">
                      {selectedMeta.description || 'This DocType is ready to power generated forms, live APIs, and future workflow rules.'}
                    </p>
                    <div className="grid gap-3">
                      {(selectedMeta.fields || []).slice(0, 6).map((field) => (
                        <div key={field.fieldname} className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
                          <div className="flex items-center justify-between gap-3">
                            <div>
                              <div className="text-sm font-semibold text-white">{field.label}</div>
                              <div className="mt-1 text-xs uppercase tracking-[0.24em] text-slate-500">{field.fieldname}</div>
                            </div>
                            <span className="rounded-full border border-white/10 px-2 py-1 text-[11px] text-slate-300">{field.fieldtype}</span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                ) : (
                  <div className="mt-4 rounded-2xl border border-dashed border-white/10 bg-white/[0.03] px-4 py-5 text-sm leading-6 text-slate-400">
                    Select a recent DocType from the sidebar or create a new one in the builder to preview its metadata here.
                  </div>
                )}
              </section>

              <section className="panel p-5">
                <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">Version history & audit trail</div>
                <div className="mt-4 space-y-3">
                  <div className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
                    <div className="text-sm font-semibold text-white">Schema drafted</div>
                    <div className="mt-1 text-xs text-slate-500">Visual builder session · now</div>
                  </div>
                  <div className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
                    <div className="text-sm font-semibold text-white">Automation hooks pending</div>
                    <div className="mt-1 text-xs text-slate-500">Client & server scripts module planned</div>
                  </div>
                  <div className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
                    <div className="text-sm font-semibold text-white">Reporting surfaces queued</div>
                    <div className="mt-1 text-xs text-slate-500">Print formats, dashboards, charts, and audit visualizations</div>
                  </div>
                </div>
              </section>

              <section className="panel p-5">
                <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">Requested power modules</div>
                <div className="mt-4 space-y-3">
                  {roadmapCards.map((card) => (
                    <div key={card.title} className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-4">
                      <div className="text-sm font-semibold text-white">{card.title}</div>
                      <p className="mt-2 text-sm leading-6 text-slate-400">{card.detail}</p>
                    </div>
                  ))}
                  <div className="rounded-2xl border border-cyan-400/15 bg-cyan-500/10 px-4 py-4 text-sm text-cyan-100">
                    Scheduler, touch animations, QR/barcode, timezone conversion, and automation tasks are now represented in the shell so the builder can grow into them instead of painting us into a corner later.
                  </div>
                </div>
              </section>
            </aside>
          </main>
        </div>
      </div>
    <ThemeSettingsPanel
      open={themeEditorOpen}
      theme={studioTheme}
      onChange={updateStudioTheme}
      onClose={() => setThemeEditorOpen(false)}
      onReset={resetStudioTheme}
    />
    </div>
  );
}
