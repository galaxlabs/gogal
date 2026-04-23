import { lazy, Suspense, useEffect, useMemo, useState } from 'react';
import { fetchDocTypes, fetchDocTypeMeta } from './lib/api.js';
import DocTypeSidebar from './components/DocTypeSidebar.jsx';
import DocTypeBuilder from './components/DocTypeBuilder.jsx';

const MetadataPanel = lazy(() => import('./components/MetadataPanel.jsx'));
const ResourceWorkbench = lazy(() => import('./components/ResourceWorkbench.jsx'));

export default function App() {
  const [docTypes, setDocTypes] = useState([]);
  const [selectedName, setSelectedName] = useState('');
  const [selectedMeta, setSelectedMeta] = useState(null);
  const [loading, setLoading] = useState(true);
  const [metaLoading, setMetaLoading] = useState(false);
  const [error, setError] = useState('');
  const [refreshKey, setRefreshKey] = useState(0);
  const [builderOpen, setBuilderOpen] = useState(false);

  const triggerRefresh = () => setRefreshKey((current) => current + 1);

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
  }, [refreshKey]);

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

  const headline = useMemo(() => {
    if (!selectedMeta) {
      return 'Metadata-powered admin studio';
    }

    return `${selectedMeta.label} · ${selectedMeta.fields?.length || 0} fields`;
  }, [selectedMeta]);

  return (
    <div className="min-h-screen px-4 py-5 text-slate-100 sm:px-6 lg:px-8">
      <div className="mx-auto flex max-w-[1800px] flex-col gap-4">
        <header className="panel flex flex-col gap-5 p-6 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <div className="mb-3 flex flex-wrap items-center gap-2">
              <span className="badge">Go API live</span>
              <span className="badge">React studio</span>
              <span className="badge">Low-code foundation</span>
            </div>
            <h1 className="text-3xl font-semibold tracking-tight text-white">Gogal Framework Studio</h1>
            <p className="mt-2 max-w-3xl text-sm text-slate-300">
              Explore doctypes, inspect metadata, search records, and create documents from the live Go backend.
            </p>
          </div>
          <div className="rounded-2xl border border-emerald-400/20 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-100">
            <div className="font-medium">{headline}</div>
            <div className="mt-1 text-emerald-200/80">API target: <span className="font-mono">/api</span></div>
          </div>
        </header>

        {error ? (
          <div className="rounded-2xl border border-rose-400/20 bg-rose-500/10 px-4 py-3 text-sm text-rose-100">
            {error}
          </div>
        ) : null}

        <div className="grid min-h-[70vh] gap-4 lg:grid-cols-[300px_minmax(0,1fr)_420px]">
          <DocTypeSidebar
            docTypes={docTypes}
            loading={loading}
            selectedName={selectedName}
            onSelect={setSelectedName}
            onRefresh={triggerRefresh}
            onCreate={() => setBuilderOpen(true)}
          />

          <Suspense fallback={<div className="panel p-6 text-sm text-slate-300">Loading live workspace…</div>}>
            <ResourceWorkbench
              key={selectedName}
              docType={selectedMeta}
              loading={metaLoading}
            />
          </Suspense>

          <Suspense fallback={<div className="panel p-6 text-sm text-slate-300">Loading metadata panel…</div>}>
            <MetadataPanel docType={selectedMeta} loading={metaLoading} />
          </Suspense>
        </div>
      </div>

      <DocTypeBuilder
        open={builderOpen}
        onClose={() => setBuilderOpen(false)}
        onCreated={(name) => {
          setSelectedName(name);
          triggerRefresh();
        }}
      />
    </div>
  );
}
