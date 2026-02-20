// Shared review utilities used by multiple components

export function formatDate(dateStr) {
  if (!dateStr) return '';
  try {
    const d = new Date(dateStr);
    return d.toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch (e) {
    return dateStr;
  }
}

export function formatRatingValue(value) {
  if (typeof value === 'boolean') {
    return value ? 'Yes' : 'No';
  }
  return value;
}

export const ACTION_STYLES = {
  accept: {
    icon: 'mdi:arrow-up',
    badgeClass:
      'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 border-green-300 dark:border-green-700',
    label: 'Accept',
  },
  reject: {
    icon: 'mdi:arrow-down',
    badgeClass:
      'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200 border-red-300 dark:border-red-700',
    label: 'Reject',
  },
  change_tracks: {
    icon: 'mdi:repeat',
    badgeClass:
      'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 border-blue-300 dark:border-blue-700',
    label: 'Change Tracks',
  },
  mark_as_duplicate: {
    icon: 'mdi:content-copy',
    badgeClass:
      'bg-orange-100 dark:bg-orange-900 text-orange-800 dark:text-orange-200 border-orange-300 dark:border-orange-700',
    label: 'Mark as Duplicate',
  },
  merge: {
    icon: 'mdi:merge',
    badgeClass:
      'bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200 border-purple-300 dark:border-purple-700',
    label: 'Merge',
  },
};
