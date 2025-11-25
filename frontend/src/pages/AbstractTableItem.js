/**
 * Helper to format timestamp (remove timezone and microseconds)
 * @param {string} dt - The datetime string to format
 * @returns {string} Formatted datetime string
 */
export function formatTimestamp(dt) {
  if (!dt) return '';
  // Remove microseconds (.123456) and timezone suffix like +00:00 or Z
  return dt.replace(/\.\d{6}/, '').replace(/([+-]\d{2}:\d{2}|Z)$/, '').replace('T', ' ');
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

/**
 * Get all track titles from an abstract (both accepted and reviewed)
 * @param {Object} abstract - The abstract data object
 * @returns {Array} Array of track objects with title and type
 */
export function getAllTracks(abstract) {
  const tracks = [];
  
  if (abstract.accepted_track) {
    tracks.push({
      title: abstract.accepted_track.title,
      type: 'accepted'
    });
  }
  
  if (abstract.reviewed_for_tracks && abstract.reviewed_for_tracks.length > 0) {
    abstract.reviewed_for_tracks.forEach(track => {
      // Don't add duplicates
      if (!tracks.some(t => t.title === track.title)) {
        tracks.push({
          title: track.title,
          type: 'reviewed'
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
  
  const primaryAuthors = persons.filter(p => p.author_type === 'primary');
  
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
  return persons.map(p => {
    const name = `${p.first_name} ${p.last_name}`;
    const type = p.author_type === 'primary' ? ' (Primary)' : '';
    const speaker = p.is_speaker ? ' 🎤' : '';
    return `${name}${type}${speaker}`;
  }).join('\n');
}

/**
 * Transform a single abstract into a table row object
 * @param {Object} abstract - The abstract data object
 * @returns {Object} Table row data
 */
export function transformAbstractToTableItem(abstract) {
  const trackTitle = abstract.accepted_track?.title || abstract.reviewed_for_tracks?.[0]?.title || '';
  const allTracks = getAllTracks(abstract);
  
  return {
    ID: abstract.friendly_id || abstract.id,
    Title: abstract.title || '',
    State: abstract.state || '',
    Submitter: abstract.submitter?.full_name || '',
    Affiliation: abstract.submitter?.affiliation || '',
    Track: getShortTrackName(trackTitle),
    TrackFull: JSON.stringify(allTracks),  // Store all tracks as JSON for the dialog
    TrackType: abstract.accepted_track ? 'accepted' : 'reviewed',
    Type: abstract.accepted_contrib_type?.name || '',
    Score: abstract.score ?? '',
    Submitted: formatTimestamp(abstract.submitted_dt),
    Authors: getPrimaryAuthorsDisplay(abstract.persons),
    AuthorsTooltip: getAllAuthorsTooltip(abstract.persons)  // All authors for tooltip
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

/**
 * Custom column rendering for Title with clickable link
 * @param {Array} tableItems - The table items for ID lookup
 * @returns {Function} Render function
 */
export function createRenderTitle(tableItems) {
  return function(data, cell, dataIndex, cellIndex) {
    const titleStr = String(data || '');
    // Get the ID from the first column of the same row
    const rowData = tableItems[dataIndex];
    const id = rowData?.ID || '';
    
    cell.childNodes = [
      {
        nodeName: 'A',
        attributes: { 
          class: 'title-link',
          href: '#',
          'data-id': String(id)
        },
        childNodes: [{ nodeName: '#text', data: titleStr }]
      }
    ];
  };
}

/**
 * Custom column rendering for State with badge styling
 */
export function renderState(data, cell, dataIndex, cellIndex) {
  const stateStr = String(data || '');
  const state = stateStr.toLowerCase();
  let bgClass = 'state-badge state-other';
  if (state === 'accepted') bgClass = 'state-badge state-accepted';
  else if (state === 'rejected') bgClass = 'state-badge state-rejected';
  
  cell.childNodes = [
    {
      nodeName: 'SPAN',
      attributes: { class: bgClass },
      childNodes: [{ nodeName: '#text', data: stateStr }]
    }
  ];
}

/**
 * Custom column rendering for Track with badge styling and clickable link
 * @param {Array} tableItems - The table items for track data lookup
 * @returns {Function} Render function
 */
export function createRenderTrack(tableItems) {
  return function(data, cell, dataIndex, cellIndex) {
    const trackStr = String(data || '');
    if (!trackStr) return;
    
    // Get the full track data from the row
    const rowData = tableItems[dataIndex];
    const trackFull = rowData?.TrackFull || '[]';
    
    // Default to reviewed style, will be updated based on TrackType
    const bgClass = 'track-badge track-reviewed track-link';
    
    cell.childNodes = [
      {
        nodeName: 'A',
        attributes: { 
          class: bgClass,
          href: '#',
          'data-tracks': trackFull
        },
        childNodes: [{ nodeName: '#text', data: trackStr }]
      }
    ];
  };
}

/**
 * Custom column rendering for Type with badge styling
 */
export function renderType(data, cell, dataIndex, cellIndex) {
  const typeStr = String(data || '');
  if (!typeStr) return;
  
  cell.childNodes = [
    {
      nodeName: 'SPAN',
      attributes: { class: 'type-badge' },
      childNodes: [{ nodeName: '#text', data: typeStr }]
    }
  ];
}

/**
 * Row render to style track based on TrackType
 */
export function rowRender(row, tr, index) {
  // Check TrackType (column 7) to style Track column (column 5)
  const trackType = row.cells[7]?.data;
  if (trackType === 'accepted' && tr.childNodes && tr.childNodes[5]) {
    // Find the track cell and update its class
    const trackCell = tr.childNodes[5];
    if (trackCell.childNodes && trackCell.childNodes[0]?.attributes) {
      trackCell.childNodes[0].attributes.class = 'track-badge track-accepted track-link';
    }
  }
  return tr;
}

/**
 * Custom column rendering for Authors with tooltip
 * @param {Array} tableItems - The table items for tooltip lookup
 * @returns {Function} Render function
 */
export function createRenderAuthors(tableItems) {
  return function(data, cell, dataIndex, cellIndex) {
    const authorsStr = String(data || '');
    if (!authorsStr) return;
    
    // Get the tooltip from the row
    const rowData = tableItems[dataIndex];
    const tooltip = rowData?.AuthorsTooltip || '';
    
    cell.childNodes = [
      {
        nodeName: 'SPAN',
        attributes: { 
          class: 'authors-cell',
          title: tooltip
        },
        childNodes: [{ nodeName: '#text', data: authorsStr }]
      }
    ];
  };
}

/**
 * Create DataTable options with column customization
 * @param {Array} tableItems - The table items for title rendering
 * @returns {Object} DataTable options configuration
 */
export function createDataTableOptions(tableItems) {
  return {
    searchable: true,
    sortable: true,
    paging: true,
    perPage: 25,
    perPageSelect: [10, 25, 50, 100],
    rowRender: rowRender,
    columns: [
      { select: 0, sortable: true, type: 'number' },  // ID
      { select: 1, render: createRenderTitle(tableItems), sortable: true, type: 'string' },  // Title
      { select: 2, render: renderState, sortable: true, type: 'string' },  // State
      { select: 3, sortable: true, type: 'string' },  // Submitter
      { select: 4, sortable: true, type: 'string' },  // Affiliation
      { select: 5, render: createRenderTrack(tableItems), sortable: true, type: 'string' },  // Track
      { select: 6, hidden: true, type: 'string' },  // TrackFull (hidden - JSON of all tracks)
      { select: 7, hidden: true, type: 'string' },  // TrackType (hidden helper column)
      { select: 8, render: renderType, sortable: true, type: 'string' },  // Type
      { select: 9, sortable: true, type: 'number' },  // Score
      { select: 10, sortable: true, type: 'string' },  // Submitted
      { select: 11, render: createRenderAuthors(tableItems), sortable: true, type: 'string' },  // Authors
      { select: 12, hidden: true, type: 'string' }  // AuthorsTooltip (hidden)
    ]
  };
}
