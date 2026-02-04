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
 * Get all tracks from abstract(s) - handles both single abstract and array of abstracts
 * When passed an array, returns unique tracks across all abstracts
 * @param {Object|Array} abstractOrArray - Single abstract object or array of abstracts
 * @returns {Array} Array of track objects with label, id, code, title and type
 */
export function getAllTracks(abstractOrArray) {
  // Handle array of abstracts - return unique tracks across all
  if (Array.isArray(abstractOrArray)) {
    const acc = [];
    const seenIds = new Set();
    const seenLabels = new Set();

    abstractOrArray.forEach((abstract) => {
      const tracks = getAllTracks(abstract); // Recursive call for single abstract
      tracks.forEach((track) => {
        // Deduplicate by ID (if present) or by label
        if (track.id != null) {
          if (seenIds.has(track.id)) return;
          seenIds.add(track.id);
        } else {
          if (seenLabels.has(track.label)) return;
          seenLabels.add(track.label);
        }
        acc.push(track);
      });
    });

    return acc;
  }

  // Handle single abstract
  const abstract = abstractOrArray;
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
 * Get track label from track ID using the getAllTracks method
 * @param {Object} abstract - The abstract data object
 * @param {number|string} trackID - The track ID to search for
 * @returns {string} The track label, or empty string if not found
 */
export function getTrackLabelByID(abstract, trackID) {
  if (!trackID || !abstract) return '';
  const allTracks = getAllTracks(abstract);
  const track = allTracks.find((t) => t.id !== null && String(t.id) === String(trackID));
  return track ? track.label : '';
}

/**
 * Check if an abstract has a track with the given ID
 * @param {Object} abstract - The abstract data object
 * @param {number|string} trackID - The track ID to search for
 * @returns {boolean} True if the abstract has this track
 */
export function abstractHasTrackID(abstract, trackID) {
  if (!trackID || !abstract) return false;
  const allTracks = getAllTracks(abstract);
  return allTracks.some((t) => t.id !== null && String(t.id) === String(trackID));
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
    FriendlyID: abstract.friendly_id,
    DatabaseID: abstract.id, // Always the actual database ID for API calls
    Title: abstract.title || '',
    State: abstract.state || '',
    Submitter: abstract.submitter?.full_name || '',
    Affiliation: affiliationDisplay,
    AffiliationFull: affiliationData ? JSON.stringify(affiliationData) : '',
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
    IsMyReview: abstract.is_my_review === true ? 'Yes' : 'No',
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
