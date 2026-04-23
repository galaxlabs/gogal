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
  sidebarWidth: '300',
  layoutDensity: 'comfortable',
};

const workspaceOptions = [
  {
    id: 'home',
    title: 'Home',
    description: 'Workspace cards and quick links.',
    badge: 'Desk',
  },
  {
    id: 'records',
    title: 'Records',
    description: 'Dynamic DocTypes and records.',
    badge: 'Live',
  },
  {
    id: 'builder',
    title: 'Builder',
    description: 'DocType canvas designer.',
    badge: 'Live',
  },
];

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

function WorkspaceButton({ title, badge, active = false, onClick }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`group w-full rounded-2xl border px-3 py-2.5 text-left transition duration-200 hover:-translate-y-0.5 active:scale-[0.99] ${
        active
          ? 'border-cyan-400/30 bg-cyan-500/10 shadow-lg shadow-cyan-500/10'
          : 'border-white/8 bg-white/[0.04] hover:border-white/15 hover:bg-white/[0.06]'
      }`}
    >
      <div className="flex items-start justify-between gap-3">
        <div>
          <div className="text-sm font-semibold text-white">{title}</div>
        </div>
        <span className={`badge ${active ? 'border-cyan-300/30 bg-cyan-400/15 text-cyan-100' : ''}`}>{badge}</span>
      </div>
    </button>
  );
}

function DocTypeLinkButton({ docType, active, onClick }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`w-full rounded-xl border px-3 py-2 text-left transition ${
        active
          ? 'border-emerald-400/30 bg-emerald-500/10 text-white'
          : 'border-white/8 bg-white/[0.03] text-slate-300 hover:border-white/15 hover:bg-white/[0.06]'
      }`}
    >
      <div className="flex items-center justify-between gap-3">
        <div className="min-w-0">
          <div className="truncate text-sm font-semibold">{docType.label || docType.doctype}</div>
          <div className="mt-1 truncate text-[11px] uppercase tracking-[0.18em] text-slate-500">{docType.doctype}</div>
        </div>
        {docType.is_system ? <span className="rounded-full border border-white/10 px-2 py-1 text-[10px] text-slate-300">system</span> : null}
      </div>
    </button>
  );
}

function HomeCard({ title, description, actionLabel, onClick }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="panel w-full rounded-[28px] p-5 text-left transition hover:-translate-y-0.5 hover:border-white/15 hover:bg-white/[0.06]"
    >
      <div className="text-[11px] font-semibold uppercase tracking-[0.24em] text-slate-500">Workspace</div>
      <h3 className="mt-3 text-xl font-semibold text-white">{title}</h3>
      <p className="mt-3 text-sm leading-6 text-slate-400">{description}</p>
      <div className="mt-5 inline-flex rounded-full border border-cyan-400/20 bg-cyan-500/10 px-3 py-1.5 text-xs font-semibold text-cyan-100">
        {actionLabel}
      </div>
    </button>
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
  const [activeWorkspace, setActiveWorkspace] = useState('home');
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

  const groupedDocTypes = useMemo(() => {
    const groups = new Map();
    filteredDocTypes.forEach((docType) => {
      const moduleName = docType.module || 'Core';
      if (!groups.has(moduleName)) {
        groups.set(moduleName, []);
      }
      groups.get(moduleName).push(docType);
    });

    return Array.from(groups.entries())
      .map(([moduleName, items]) => ({
        moduleName,
        items: items.sort((left, right) => (left.label || left.doctype).localeCompare(right.label || right.doctype)),
      }))
      .sort((left, right) => left.moduleName.localeCompare(right.moduleName));
  }, [filteredDocTypes]);

  const totalDocTypes = filteredDocTypes.length;
  const totalModules = groupedDocTypes.length;
  const systemDocTypes = filteredDocTypes.filter((item) => item.is_system).length;

  const selectedWorkspaceMeta = useMemo(
    () => workspaceOptions.find((item) => item.id === activeWorkspace) || workspaceOptions[0],
    [activeWorkspace],
  );

  const selectedModuleName = selectedMeta?.module || docTypes.find((item) => item.doctype === selectedName)?.module || 'Core';

  const selectedHeaderTitle = useMemo(() => {
    if (activeWorkspace === 'home') {
      return 'Workspace Home';
    }
    if (activeWorkspace === 'builder') {
      return 'DocType Builder';
    }
    return selectedMeta?.label || selectedName || 'Records';
  }, [activeWorkspace, selectedMeta, selectedName]);

  const selectedHeaderSubtitle = useMemo(() => {
    if (activeWorkspace === 'home') {
      return 'Clean workspace with cards, modules, and quick links.';
    }
    if (activeWorkspace === 'builder') {
      return 'Canvas-based DocType designer.';
    }
    return `${selectedModuleName} module`;
  }, [activeWorkspace, selectedModuleName]);

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
    setSelectedMeta(created?.fields ? created : null);
    setActiveWorkspace('records');
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
          style={{ width: `min(90vw, ${studioTheme.sidebarWidth}px)` }}
          className={`studio-sidebar fixed inset-y-0 left-0 z-40 overflow-y-auto border-r border-white/8 px-3 py-4 transition duration-300 lg:static lg:min-h-screen lg:translate-x-0 lg:px-4 lg:py-4 ${navigationOpen ? 'translate-x-0' : '-translate-x-full'}`}
        >
          <div className="panel relative overflow-hidden p-3">
            <div className="relative flex items-center gap-3">
              <div className="flex h-11 w-11 items-center justify-center rounded-2xl border border-cyan-300/20 bg-cyan-400/10 text-sm font-semibold text-cyan-100 shadow-lg shadow-cyan-500/10">
                GG
              </div>
              <div className="min-w-0 flex-1">
                <div className="text-[11px] font-semibold uppercase tracking-[0.28em] text-cyan-200/80">UI Studio</div>
                <h1 className="mt-1 truncate text-base font-semibold text-white">{tenantName}</h1>
              </div>
              <button
                type="button"
                onClick={() => setThemeEditorOpen(true)}
                className="rounded-xl border border-white/10 bg-white/[0.05] px-2.5 py-2 text-xs font-semibold text-slate-200 transition hover:bg-white/[0.1]"
              >
                ⚙
              </button>
            </div>
            <button
              type="button"
              onClick={() => setNavigationOpen(false)}
              className="mt-4 rounded-2xl border border-white/10 bg-white/[0.05] px-3 py-2 text-sm font-semibold text-slate-200 transition hover:bg-white/[0.1] lg:hidden"
            >
              Close drawer
            </button>
          </div>

          <div className="mt-4 space-y-4">
            <div>
              <div className="mb-3 flex items-center justify-between">
                <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">Workspaces</div>
                <span className="badge">{workspaceOptions.length}</span>
              </div>
              <div className="space-y-3">
                {workspaceOptions.map((workspace) => (
                  <WorkspaceButton
                    key={workspace.id}
                    title={workspace.title}
                    badge={workspace.badge}
                    active={activeWorkspace === workspace.id}
                    onClick={() => setActiveWorkspace(workspace.id)}
                  />
                ))}
              </div>
            </div>

            <div>
              <div className="mb-3 flex items-center justify-between">
                <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">Modules → DocTypes</div>
                <span className="badge">{totalModules} modules</span>
              </div>

              {loading ? (
                <SidebarSkeleton />
              ) : groupedDocTypes.length > 0 ? (
                <div className="space-y-3">
                  {groupedDocTypes.map((group) => (
                    <section key={group.moduleName} className="panel p-3">
                      <div className="flex items-center justify-between gap-3">
                        <div className="text-sm font-semibold text-white">{group.moduleName}</div>
                        <div className="rounded-full border border-white/10 px-2 py-1 text-[10px] text-slate-300">
                          {group.items.length}
                        </div>
                      </div>
                      <div className="mt-3 space-y-2">
                        {group.items.map((docType) => (
                          <DocTypeLinkButton
                            key={docType.doctype}
                            docType={docType}
                            active={selectedName === docType.doctype && activeWorkspace === 'records'}
                            onClick={() => {
                              setSelectedName(docType.doctype);
                              setActiveWorkspace('records');
                            }}
                          />
                        ))}
                      </div>
                    </section>
                  ))}
                </div>
              ) : (
                <div className="rounded-2xl border border-dashed border-white/10 bg-white/[0.03] px-4 py-5 text-sm text-slate-400">
                  No module or doctype matched the current search.
                </div>
              )}
            </div>

          </div>
        </aside>

        <div className="flex-1">
          <header className="sticky top-0 z-20 border-b border-white/8 bg-slate-950/85 px-4 py-3 backdrop-blur-xl lg:px-6">
            <div className="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
              <div className="flex min-w-0 items-center gap-3">
                <button
                  type="button"
                  onClick={() => setNavigationOpen(true)}
                  className="rounded-full border border-white/10 bg-white/[0.05] px-3 py-1.5 text-xs font-semibold text-slate-100 transition hover:bg-white/[0.1] lg:hidden"
                >
                  Menu
                </button>
                <div className="min-w-0">
                  <div className="text-[11px] font-semibold uppercase tracking-[0.28em] text-slate-500">{selectedWorkspaceMeta.title}</div>
                  <h2 className="truncate text-lg font-semibold tracking-tight text-white">{selectedHeaderTitle}</h2>
                  <p className="truncate text-sm text-slate-400">{selectedHeaderSubtitle}</p>
                </div>
              </div>

              <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-end">
                <label className="relative block min-w-[min(100%,22rem)] md:w-[320px]">
                  <span className="pointer-events-none absolute inset-y-0 left-4 flex items-center text-slate-500">⌕</span>
                  <input
                    className="field h-10 rounded-2xl pl-10 pr-16"
                    value={search}
                    onChange={(event) => setSearch(event.target.value)}
                    placeholder="Search modules or DocTypes"
                  />
                  <span className="pointer-events-none absolute inset-y-0 right-3 flex items-center text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500">⌘K</span>
                </label>
              </div>
            </div>
          </header>

          <main className="px-4 py-6 lg:px-6">
            {error ? (
              <div className="mb-6 rounded-2xl border border-rose-400/20 bg-rose-500/10 px-4 py-3 text-sm text-rose-100">
                {error}
              </div>
            ) : null}

            <div className="mb-6 grid gap-3 md:grid-cols-3">
              <div className="panel rounded-[24px] px-4 py-3">
                <div className="text-[11px] font-semibold uppercase tracking-[0.22em] text-slate-500">DocTypes</div>
                <div className="mt-1 text-lg font-semibold text-white">{totalDocTypes}</div>
              </div>
              <div className="panel rounded-[24px] px-4 py-3">
                <div className="text-[11px] font-semibold uppercase tracking-[0.22em] text-slate-500">System</div>
                <div className="mt-1 text-lg font-semibold text-white">{systemDocTypes}</div>
              </div>
              <div className="panel rounded-[24px] px-4 py-3">
                <div className="text-[11px] font-semibold uppercase tracking-[0.22em] text-slate-500">Timezone</div>
                <div className="mt-1 truncate text-lg font-semibold text-white">{timezone}</div>
              </div>
            </div>

            {activeWorkspace === 'home' ? (
              <div className="space-y-6">
                <div className="grid gap-5 lg:grid-cols-3">
                  <HomeCard title="DocType Builder" description="Design DocTypes on a cleaner canvas with sections, columns, and add-field controls." actionLabel="Open Builder" onClick={() => setActiveWorkspace('builder')} />
                  <HomeCard title="Records Desk" description="Open dynamic DocTypes, list view, and form view from the grouped module sidebar." actionLabel="Open Records" onClick={() => setActiveWorkspace('records')} />
                  <HomeCard title="Modules" description={`${totalModules} modules and ${totalDocTypes} DocTypes are available dynamically in the sidebar.`} actionLabel="Browse Sidebar" onClick={() => setActiveWorkspace('records')} />
                </div>

                <section className="panel p-6">
                  <div className="flex items-center justify-between gap-3">
                    <div>
                      <div className="text-[11px] font-semibold uppercase tracking-[0.24em] text-slate-500">Workspace links</div>
                      <h3 className="mt-2 text-xl font-semibold text-white">Modules and DocTypes</h3>
                    </div>
                    <div className="rounded-full border border-white/10 px-3 py-1 text-xs text-slate-300">Dynamic</div>
                  </div>
                  <div className="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                    {groupedDocTypes.map((group) => (
                      <div key={`home-${group.moduleName}`} className="rounded-[24px] border border-white/10 bg-white/[0.03] p-4">
                        <div className="text-sm font-semibold text-white">{group.moduleName}</div>
                        <div className="mt-3 space-y-2">
                          {group.items.slice(0, 5).map((docType) => (
                            <button
                              key={`home-link-${docType.doctype}`}
                              type="button"
                              onClick={() => {
                                setSelectedName(docType.doctype);
                                setActiveWorkspace('records');
                              }}
                              className="block w-full rounded-xl border border-white/8 bg-white/[0.03] px-3 py-2 text-left text-sm text-slate-200 transition hover:bg-white/[0.06]"
                            >
                              {docType.label || docType.doctype}
                            </button>
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                </section>
              </div>
            ) : null}

            {activeWorkspace === 'builder' ? (
              <DocTypeBuilder
                moduleOptions={groupedDocTypes.map((group) => group.moduleName)}
                existingDocTypes={docTypes.map((item) => item.doctype)}
                onCreated={handleCreated}
              />
            ) : null}

            {activeWorkspace === 'records' ? (
              <ResourceWorkbench docType={selectedMeta} loading={metaLoading} />
            ) : null}
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
