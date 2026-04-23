const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '').replace(/\/$/, '');

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
    ...options,
  });

  const isJson = response.headers.get('content-type')?.includes('application/json');
  const payload = isJson ? await response.json() : null;

  if (!response.ok) {
    const message = payload?.error?.message || `Request failed with status ${response.status}`;
    throw new Error(message);
  }

  return payload;
}

export async function fetchDocTypes() {
  const payload = await request('/api/doctypes');
  return payload.data || [];
}

export async function fetchDocTypeMeta(name) {
  const payload = await request(`/api/doctypes/${encodeURIComponent(name)}/meta`);
  return payload.data;
}

export async function createDocType(document) {
  const payload = await request('/api/doctypes', {
    method: 'POST',
    body: JSON.stringify(document),
  });
  return payload.data;
}

export async function fetchResources(doctype, query = {}) {
  const params = new URLSearchParams();
  Object.entries(query).forEach(([key, value]) => {
    if (value === undefined || value === null || value === '') {
      return;
    }
    params.set(key, value);
  });

  const suffix = params.size > 0 ? `?${params.toString()}` : '';
  const payload = await request(`/api/resource/${encodeURIComponent(doctype)}${suffix}`);
  return payload;
}

export async function createResource(doctype, document) {
  const payload = await request(`/api/resource/${encodeURIComponent(doctype)}`, {
    method: 'POST',
    body: JSON.stringify(document),
  });
  return payload.data;
}

export async function updateResource(doctype, name, document) {
  const payload = await request(`/api/resource/${encodeURIComponent(doctype)}/${encodeURIComponent(name)}`, {
    method: 'PUT',
    body: JSON.stringify(document),
  });
  return payload.data;
}

export async function deleteResource(doctype, name) {
  return request(`/api/resource/${encodeURIComponent(doctype)}/${encodeURIComponent(name)}`, {
    method: 'DELETE',
  });
}
