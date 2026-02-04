/**
 * Parse a date string YYYY-MM-DD into numeric [year, month, day]
 */
function parseDateYMD(dateStr) {
  if (!dateStr) return null;
  const parts = dateStr.split('-').map((p) => parseInt(p, 10));
  if (parts.length < 3 || parts.some((n) => Number.isNaN(n))) return null;
  return parts; // [year, month, day]
}

/**
 * Parse a time string HH:MM or HH:MM:SS into numeric [hour, minute, second]
 */
function parseTimeHms(timeStr) {
  if (!timeStr) return null;
  const parts = timeStr.split(':').map((p) => parseInt(p, 10));
  if (parts.length < 2 || parts.some((n) => Number.isNaN(n))) return null;
  const [hour, minute, second] = [parts[0], parts[1], parts[2] || 0];
  return [hour, minute, second];
}

function pad2(n) {
  return String(n).padStart(2, '0');
}

/**
 * Convert a datetime expressed in an IANA timezone (tz) into the local timezone.
 * Input:
 *   date: 'YYYY-MM-DD' (e.g. '2025-04-07')
 *   time: 'HH:MM:SS' (e.g. '08:30:00') or 'HH:MM'
 *   tz:   IANA timezone name (e.g. 'Europe/London')
 * Returns formatted string: 'YYYY-MM-DD HH:MM:SS (Local/Zone)'
 *
 * This uses Intl.DateTimeFormat.formatToParts to compute the offset for the
 * provided timezone at the given wall-clock time and converts to an absolute
 * UTC timestamp, which is then formatted in the runtime's local timezone.
 */
export function convertDateTimeToLocal(date, time, tz) {
  if (!date || !time) return '';

  const dateParts = parseDateYMD(date);
  const timeParts = parseTimeHms(time);
  if (!dateParts || !timeParts) {
    // fallback: return original inputs
    return tz ? `${date} ${time} (${tz})` : `${date} ${time}`;
  }

  const [year, month, day] = dateParts;
  const [hour, minute, second] = timeParts;

  try {
    // Step 1: create a Date object for the same wall-clock time but treated as UTC
    // (i.e. Date.UTC(year, month-1, day, hour, minute, second))
    const assumedUtcMillis = Date.UTC(year, month - 1, day, hour, minute, second);
    const assumedUtcDate = new Date(assumedUtcMillis);

    // Step 2: format that UTC instant into the target timezone to discover
    // how the timezone maps instants to wall-clock fields.
    const dtf = new Intl.DateTimeFormat('en-US', {
      timeZone: tz,
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    });
    const parts = dtf.formatToParts(assumedUtcDate);
    const map = {};
    for (const p of parts) {
      if (p.type !== 'literal') map[p.type] = p.value;
    }

    const tzYear = parseInt(map.year, 10);
    const tzMonth = parseInt(map.month, 10);
    const tzDay = parseInt(map.day, 10);
    const tzHour = parseInt(map.hour, 10);
    const tzMinute = parseInt(map.minute, 10);
    const tzSecond = parseInt(map.second, 10);

    if ([tzYear, tzMonth, tzDay, tzHour, tzMinute, tzSecond].some((n) => Number.isNaN(n))) {
      throw new Error('Failed to obtain timezone parts');
    }

    // Step 3: compute delta between the assumed UTC instant and the UTC
    // equivalent of what that instant looks like in the target timezone.
    const asUtcFromTz = Date.UTC(tzYear, tzMonth - 1, tzDay, tzHour, tzMinute, tzSecond);
    const delta = assumedUtcDate.getTime() - asUtcFromTz;

    // Step 4: to convert the provided wall-clock (year,month,day,hour,minute,second)
    // which is expressed in `tz` into an absolute UTC timestamp, add delta.
    const targetUtcMillis = Date.UTC(year, month - 1, day, hour, minute, second) + delta;
    const localDate = new Date(targetUtcMillis);

    // Format output as YYYY-MM-DD HH:MM:SS (LocalZone)
    const localYear = localDate.getFullYear();
    const localMonth = pad2(localDate.getMonth() + 1);
    const localDay = pad2(localDate.getDate());
    const localHour = pad2(localDate.getHours());
    const localMinute = pad2(localDate.getMinutes());
    const localSecond = pad2(localDate.getSeconds());
    const localTz = Intl.DateTimeFormat().resolvedOptions().timeZone || 'local';

    return `${localYear}-${localMonth}-${localDay} ${localHour}:${localMinute}:${localSecond} (${localTz})`;
  } catch (e) {
    console.warn('convertDateTimeToLocal failed:', e);
    return tz ? `${date} ${time} (${tz})` : `${date} ${time}`;
  }
}
