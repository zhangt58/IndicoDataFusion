/**
 * Format date info object to string in local timezone
 * @param {Object} dateInfo - The date info object with date, time, tz
 * @returns {string} Formatted date string in local timezone
 */
export function formatDate(dateInfo) {
  if (!dateInfo) return '';
  const { date, time, tz } = dateInfo;

  // Try to convert to local timezone
  try {
    // Combine date and time into ISO format
    const dateTimeStr = `${date}T${time}`;
    const dateObj = new Date(dateTimeStr);

    if (!isNaN(dateObj.getTime())) {
      // Format in local timezone
      const localDate = dateObj.toLocaleDateString(undefined, {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
      });
      const localTime = dateObj.toLocaleTimeString(undefined, {
        hour: '2-digit',
        minute: '2-digit',
        hour12: false,
      });
      const localTz = Intl.DateTimeFormat().resolvedOptions().timeZone;

      return `${localDate} ${localTime} (${localTz})`;
    }
  } catch (e) {
    // Fall back to original format if conversion fails
    console.warn('Failed to convert date to local timezone:', e);
  }

  return tz ? `${date} ${time} (${tz})` : `${date} ${time}`;
}
