/**
 * Helper to format timestamp (remove timezone and microseconds)
 * @param {string} dt - The datetime string to format
 * @returns {string} Formatted datetime string
 */
export function formatTimestamp(dt) {
  if (!dt) return '';
  // Remove microseconds (.123456) and timezone suffix like +00:00 or Z
  return dt
    .replace(/\.\d{6}/, '')
    .replace(/([+-]\d{2}:\d{2}|Z)$/, '')
    .replace('T', ' ');
}

/**
 * Extract the short track name (part before ":")
 * @param {string} trackTitle - Full track title
 * @returns {string} Short track name
 */
export function getShortTrackName(trackTitle) {
  if (!trackTitle) return '';
  const colonIndex = trackTitle.indexOf(':');
  return colonIndex > 0 ? trackTitle.substring(0, colonIndex).trim() : trackTitle;
}

/*
 * Get a display label for a track: prefer non-empty code, otherwise title
 * @param {Object} track - track object with optional `code` and `title`
 * @returns {string} label to display
 */
export function getTrackLabel(track) {
  if (!track) return '';
  const code = track.code;
  if (typeof code === 'string' && code.trim() !== '') return code;
  return track.title ?? '';
}

/**
 * Get all track titles from an abstract (both accepted and reviewed)
 * @param {Object} abstract - The abstract data object
 * @returns {Array} Array of track objects with label, id, code, title and type
 */
export function getAllTracks(abstract) {
  const tracks = [];

  if (abstract.accepted_track) {
    const t = abstract.accepted_track;
    tracks.push({
      label: getTrackLabel(t),
      id: t.id ?? null,
      code: t.code ?? '',
      title: t.title ?? '',
      type: 'accepted',
    });
  }

  if (abstract.reviewed_for_tracks && abstract.reviewed_for_tracks.length > 0) {
    abstract.reviewed_for_tracks.forEach((track) => {
      // Don't add duplicates (compare by label)
      const trackLabel = getTrackLabel(track);
      if (!tracks.some((t) => t.label === trackLabel)) {
        tracks.push({
          label: trackLabel,
          id: track.id ?? null,
          code: track.code ?? '',
          title: track.title ?? '',
          type: 'reviewed',
        });
      }
    });
  }

  return tracks;
}

/**
 * Get primary authors display text (at most one name, with ... if more exist)
 * @param {Array} persons - Array of person objects
 * @returns {string} Display text for primary authors
 */
export function getPrimaryAuthorsDisplay(persons) {
  if (!persons || persons.length === 0) return '';

  const primaryAuthors = persons.filter((p) => p.author_type === 'primary');

  if (primaryAuthors.length === 0) {
    // Fall back to first person if no primary authors
    const first = persons[0];
    return persons.length > 1
      ? `${first.first_name} ${first.last_name} ...`
      : `${first.first_name} ${first.last_name}`;
  }

  const first = primaryAuthors[0];
  return primaryAuthors.length > 1
    ? `${first.first_name} ${first.last_name} ...`
    : `${first.first_name} ${first.last_name}`;
}

/**
 * Get all authors as a string for tooltip
 * @param {Array} persons - Array of person objects
 * @returns {string} All authors joined by newlines
 */
export function getAllAuthorsTooltip(persons) {
  if (!persons || persons.length === 0) return '';
  return persons
    .map((p) => {
      const name = `${p.first_name} ${p.last_name}`;
      const type = p.author_type === 'primary' ? ' (Primary)' : '';
      const speaker = p.is_speaker ? ' 🎤' : '';
      return `${name}${type}${speaker}`;
    })
    .join('\n');
}

/**
 * Transform a single abstract into a table row object
 * @param {Object} abstract - The abstract data object
 * @returns {Object} Table row data
 */
export function transformAbstractToTableItem(abstract) {
  // Prefer non-empty track.code if available, otherwise fall back to title
  let trackTitle = '';
  if (abstract.accepted_track) {
    trackTitle = getTrackLabel(abstract.accepted_track);
  } else if (abstract.reviewed_for_tracks && abstract.reviewed_for_tracks.length > 0) {
    const t0 = abstract.reviewed_for_tracks[0];
    trackTitle = getTrackLabel(t0);
  }
  const allTracks = getAllTracks(abstract);

  // compute numeric ID if possible
  const rawId = abstract.friendly_id ?? abstract.id;
  const idNum = Number(rawId);

  // compute submitted timestamp millis if possible
  let submittedISO = '';
  let submittedMillis = NaN;
  if (abstract.submitted_dt) {
    submittedISO = String(abstract.submitted_dt);
    const d = new Date(submittedISO);
    if (!isNaN(d.getTime())) submittedMillis = d.getTime();
  }

  // Extract affiliation display text and full data
  let affiliationDisplay = '';
  let affiliationData = null;
  if (abstract.submitter?.affiliation) {
    if (typeof abstract.submitter.affiliation === 'object') {
      // Structured affiliation object
      affiliationDisplay = abstract.submitter.affiliation.name || '';
      affiliationData = abstract.submitter.affiliation;
    } else {
      // Legacy string affiliation
      affiliationDisplay = abstract.submitter.affiliation;
    }
  }

  // Prepare reviewed/submitted track title arrays (prefer non-empty code)
  const reviewedTrackTitles = (abstract.reviewed_for_tracks || []).map((t) => getTrackLabel(t));
  const submittedTrackTitles = (abstract.submitted_for_tracks || []).map((t) => getTrackLabel(t));

  return {
    ID: rawId,
    IDNumber: isNaN(idNum) ? null : idNum,
    DatabaseID: abstract.id, // Always the actual database ID for API calls
    Title: abstract.title || '',
    State: abstract.state || '',
    Submitter: abstract.submitter?.full_name || '',
    Affiliation: affiliationDisplay,
    AffiliationFull: affiliationData ? JSON.stringify(affiliationData) : '',
    Track: getShortTrackName(trackTitle),
    TrackFull: JSON.stringify(allTracks), // Store all tracks as JSON for the dialog
    TrackType: abstract.accepted_track ? 'accepted' : 'reviewed',
    Type: abstract.accepted_contrib_type?.name || '',
    // New explicit fields requested (prefer non-empty code when available)
    AcceptedTrack: abstract.accepted_track ? getTrackLabel(abstract.accepted_track) : '',
    AcceptedContribType: abstract.accepted_contrib_type?.name || '',
    SubmittedContribType: abstract.submitted_contrib_type?.name || '',
    ReviewedForTracks: reviewedTrackTitles,
    SubmittedForTracks: submittedTrackTitles,

    Score: abstract.score ?? '',
    Submitted: formatTimestamp(abstract.submitted_dt),
    SubmittedISO: submittedISO,
    SubmittedMillis: submittedMillis,
    Authors: getPrimaryAuthorsDisplay(abstract.persons),
    AuthorsTooltip: getAllAuthorsTooltip(abstract.persons), // All authors for tooltip
    FirstPriority: abstract.first_priority ?? 0,
    SecondPriority: abstract.second_priority ?? 0,
  };
}

/**
 * Transform an array of abstracts into table items
 * @param {Array} data - Array of abstract objects
 * @returns {Array} Array of table row objects
 */
export function getTableItems(data) {
  return data.map(transformAbstractToTableItem);
}
