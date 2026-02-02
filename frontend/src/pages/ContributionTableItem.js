import { formatDate } from '../utils/dateUtils.js';
import { deduplicateAttachments } from '../utils/attachmentUtils.js';

/**
 * Get speakers display text (first name, with ... if more exist)
 * @param {Array} speakers - Array of speaker objects
 * @returns {string} Display text for speakers
 */
export function getSpeakersDisplay(speakers) {
  if (!speakers || speakers.length === 0) return '';

  const first = speakers[0];
  const name = first.fullName || `${first.first_name} ${first.last_name}`;
  return speakers.length > 1 ? `${name} ...` : name;
}

/**
 * Get speakers affiliations (first affiliation, with ... if more exist)
 * @param {Array} speakers - Array of speaker objects
 * @returns {string} Display text for speakers' affiliations
 */
export function getSpeakersAffiliations(speakers) {
  if (!speakers || speakers.length === 0) return '';

  const first = speakers[0];
  const affiliation = first.affiliation || '';
  return speakers.length > 1 ? `${affiliation} ...` : affiliation;
}

/**
 * Get all speakers as a string for tooltip
 * @param {Array} speakers - Array of speaker objects
 * @returns {string} All speakers joined by newlines
 */
export function getSpeakersTooltip(speakers) {
  if (!speakers || speakers.length === 0) return '';
  return speakers
    .map((s) => {
      const name = s.fullName || `${s.first_name} ${s.last_name}`;
      const affiliation = s.affiliation ? ` (${s.affiliation})` : '';
      return `🎤 ${name}${affiliation}`;
    })
    .join('\n');
}

/**
 * Get primary authors display text (first name, with ... if more exist)
 * @param {Array} authors - Array of author objects
 * @returns {string} Display text for primary authors
 */
export function getPrimaryAuthorsDisplay(authors) {
  if (!authors || authors.length === 0) return '';

  const first = authors[0];
  const name = first.fullName || `${first.first_name} ${first.last_name}`;
  return authors.length > 1 ? `${name} ...` : name;
}

/**
 * Get all authors as a string for tooltip (primary and co-authors)
 * @param {Array} primaryauthors - Array of primary author objects
 * @param {Array} coauthors - Array of co-author objects
 * @returns {string} All authors joined by newlines
 */
export function getAllAuthorsTooltip(primaryauthors, coauthors) {
  const lines = [];

  if (primaryauthors && primaryauthors.length > 0) {
    lines.push('Primary Authors:');
    primaryauthors.forEach((a) => {
      const name = a.fullName || `${a.first_name} ${a.last_name}`;
      const affiliation = a.affiliation ? ` (${a.affiliation})` : '';
      lines.push(`  ${name}${affiliation}`);
    });
  }

  if (coauthors && coauthors.length > 0) {
    if (lines.length > 0) lines.push('');
    lines.push('Co-Authors:');
    coauthors.forEach((a) => {
      const name = a.fullName || `${a.first_name} ${a.last_name}`;
      const affiliation = a.affiliation ? ` (${a.affiliation})` : '';
      lines.push(`  ${name}${affiliation}`);
    });
  }

  return lines.join('\n');
}

/**
 * Extract attachments from contribution.folders (flatten)
 * Returns array of simplified attachment objects: { title, filename, download_url, content_type }
 *
 * This implementation delegates deduplication to the shared `deduplicateAttachments` helper
 * (used by `AttachmentGrid`). We first flatten raw attachments, call the helper when dedupe is
 * desired, and then map to the simplified shape expected by table consumers.
 */
function extractAttachments(contribution) {
  const raw = [];
  if (!contribution || !contribution.folders) return raw;

  for (const folder of contribution.folders) {
    if (!folder || !Array.isArray(folder.attachments)) continue;
    for (const att of folder.attachments) {
      raw.push(att);
    }
  }

  // Use the shared dedupe helper to normalize duplicates by title/filename.
  const deduped = deduplicateAttachments(raw);

  // Map to simplified objects expected by the table components
  return (deduped || []).map((att) => ({
    title: att && (att.title || att.filename) ? (att.title || att.filename) : '',
    filename: att && (att.filename || '') || '',
    download_url: att && (att.download_url || att.url || att.downloadUrl || '') || '',
    content_type: att && (att.content_type || att.contentType || '') || '',
  }));
}

/**
 * Transform a single contribution into a table row object
 * @param {Object} contribution - The contribution data object
 * @returns {Object} Table row data
 */
export function transformContributionToTableItem(contribution) {
  // compute numeric ID if possible
  const rawId = contribution.friendly_id ?? contribution.id;
  const idNum = Number(rawId);

  // compute duration minutes (backend provides duration as number in minutes)
  const durationMinutes =
    typeof contribution.duration === 'number'
      ? contribution.duration
      : contribution.duration
        ? Number(contribution.duration)
        : NaN;

  // compute ISO datetime and millis for Start (if startDate object present)
  let startISO = '';
  let startMillis = NaN;
  if (contribution.startDate && contribution.startDate.date && contribution.startDate.time) {
    // combine to ISO-like string: YYYY-MM-DDTHH:MM[:SS]
    startISO = `${contribution.startDate.date}T${contribution.startDate.time}`;
    const d = new Date(startISO);
    if (!isNaN(d.getTime())) startMillis = d.getTime();
  }

  const attachments = extractAttachments(contribution);
  const attachmentsTooltip = attachments.map((a) => a.title || a.filename || '').join('\n');

  return {
    ID: rawId,
    IDNumber: isNaN(idNum) ? null : idNum,
    Code: contribution.code || '',
    Title: contribution.title || '',
    Type: contribution.type || '',
    Session: contribution.session || '',
    Track: contribution.track || '',
    // preserve both Start (formatted) and StartDate for compatibility
    Start: formatDate(contribution.startDate),
    StartDate: formatDate(contribution.startDate),
    StartISO: startISO,
    StartMillis: startMillis,
    Duration: contribution.duration ? `${contribution.duration} min` : '',
    DurationMinutes: isNaN(durationMinutes) ? null : durationMinutes,
    Location: contribution.location || '',
    Room: contribution.roomFullname || contribution.room || '',
    Speakers: getSpeakersDisplay(contribution.speakers),
    SpeakersAffiliations: getSpeakersAffiliations(contribution.speakers),
    // Add an alias 'Affiliations' so table columns that refer to 'Affiliations' can access data
    Affiliations: getSpeakersAffiliations(contribution.speakers),
    SpeakersTooltip: getSpeakersTooltip(contribution.speakers),
    Authors: getPrimaryAuthorsDisplay(contribution.primaryauthors),
    AuthorsTooltip: getAllAuthorsTooltip(contribution.primaryauthors, contribution.coauthors),
    URL: contribution.url || '',
    // Attachments: array of simplified attachment objects (may be empty)
    Attachments: attachments,
    AttachmentsCount: attachments.length,
    AttachmentsTooltip: attachmentsTooltip,
  };
}

/**
 * Transform an array of contributions into table items
 * @param {Array} data - Array of contribution objects
 * @returns {Array} Array of table row objects
 */
export function getTableItems(data) {
  return data.map(transformContributionToTableItem);
}
