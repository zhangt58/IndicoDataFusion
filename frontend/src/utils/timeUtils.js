/**
 * Format a timestamp into a human-readable relative time string
 * @param {string|Date} timestamp - The timestamp to format
 * @returns {string} Human-readable relative time (e.g., "2 minutes ago", "1 hour ago")
 */
export function formatRelativeTime(timestamp) {
  if (!timestamp) return 'Never';

  const date = typeof timestamp === 'string' ? new Date(timestamp) : timestamp;
  const now = new Date();
  const diffMs = now - date;
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSecs < 5) {
    return 'Just now';
  } else if (diffSecs < 60) {
    return `${diffSecs} second${diffSecs === 1 ? '' : 's'} ago`;
  } else if (diffMins < 60) {
    return `${diffMins} minute${diffMins === 1 ? '' : 's'} ago`;
  } else if (diffHours < 24) {
    return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`;
  } else if (diffDays < 7) {
    return `${diffDays} day${diffDays === 1 ? '' : 's'} ago`;
  } else {
    // For older dates, show the actual date and time
    return date.toLocaleString();
  }
}

/**
 * Format a timestamp into a full date/time string
 * @param {string|Date} timestamp - The timestamp to format
 * @returns {string} Full date and time string
 */
export function formatFullDateTime(timestamp) {
  if (!timestamp) return 'Never';

  const date = typeof timestamp === 'string' ? new Date(timestamp) : timestamp;
  return date.toLocaleString();
}

