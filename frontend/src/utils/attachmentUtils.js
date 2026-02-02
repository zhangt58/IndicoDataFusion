// Utility helpers for attachment handling
// Exported so components/pages can reuse deduplication logic.

export function deduplicateAttachments(arr) {
  if (!arr || arr.length === 0) return [];

  const titleMap = new Map();
  for (const attachment of arr) {
    const rawKey = attachment.title || attachment.filename || 'untitled';
    let key = String(rawKey).trim().toLowerCase();
    const existing = titleMap.get(key);

    if (!existing) {
      titleMap.set(key, attachment);
    } else {
      const existingDate = existing.modified_dt ? new Date(existing.modified_dt) : new Date(0);
      const currentDate = attachment.modified_dt ? new Date(attachment.modified_dt) : new Date(0);
      if (currentDate > existingDate) {
        titleMap.set(key, attachment);
      }
    }
  }

  return Array.from(titleMap.values());
}
