// Utility helpers for data source UI behavior (tags, favorite)
// Keep these pure and simple so they can be reused across components.

// Add a tag string to a data source object (mutates ds.tags)
export function addTagTo(ds, tag) {
  if (!tag || !ds) return false;
  const t = String(tag).trim();
  if (!t) return false;
  if (!ds.tags) ds.tags = [];
  if (!ds.tags.includes(t)) {
    ds.tags = [...ds.tags, t];
    return true;
  }
  return false;
}

// Remove a tag by index from a data source object (mutates ds.tags)
export function removeTagFrom(ds, idx) {
  if (!ds || !ds.tags) return false;
  if (idx < 0 || idx >= ds.tags.length) return false;
  ds.tags = ds.tags.filter((_, i) => i !== idx);
  return true;
}

// Toggle favorite flag on a data source object (mutates ds.favorite)
export function toggleFavoriteOn(ds) {
  if (!ds) return false;
  ds.favorite = !ds.favorite;
  return ds.favorite;
}

// Collect all unique tags from an array of data sources
export function collectAllTags(dataSources) {
  const s = new Set();
  if (!dataSources) return [];
  for (const ds of dataSources) {
    if (ds && ds.tags) {
      for (const t of ds.tags) {
        if (t) s.add(String(t));
      }
    }
  }
  return Array.from(s).sort();
}

// Collect all unique base URLs from data sources (for Indico type entries)
export function collectAllBaseUrls(dataSources) {
  const s = new Set();
  if (!dataSources) return [];
  for (const ds of dataSources) {
    if (ds && ds.type === 'indico' && ds.indico && ds.indico.baseUrl) {
      const b = String(ds.indico.baseUrl || '').trim();
      if (b) s.add(b);
    }
  }
  return Array.from(s).sort();
}
