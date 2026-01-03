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
      type: 'accepted',
    });
  }

  if (abstract.reviewed_for_tracks && abstract.reviewed_for_tracks.length > 0) {
    abstract.reviewed_for_tracks.forEach((track) => {
      // Don't add duplicates
      if (!tracks.some((t) => t.title === track.title)) {
        tracks.push({
          title: track.title,
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
  const trackTitle =
    abstract.accepted_track?.title || abstract.reviewed_for_tracks?.[0]?.title || '';
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

  return {
    ID: rawId,
    IDNumber: isNaN(idNum) ? null : idNum,
    Title: abstract.title || '',
    State: abstract.state || '',
    Submitter: abstract.submitter?.full_name || '',
    Affiliation: affiliationDisplay,
    AffiliationFull: affiliationData ? JSON.stringify(affiliationData) : '',
    Track: getShortTrackName(trackTitle),
    TrackFull: JSON.stringify(allTracks), // Store all tracks as JSON for the dialog
    TrackType: abstract.accepted_track ? 'accepted' : 'reviewed',
    Type: abstract.accepted_contrib_type?.name || '',
    Score: abstract.score ?? '',
    Submitted: formatTimestamp(abstract.submitted_dt),
    SubmittedISO: submittedISO,
    SubmittedMillis: submittedMillis,
    Authors: getPrimaryAuthorsDisplay(abstract.persons),
    AuthorsTooltip: getAllAuthorsTooltip(abstract.persons), // All authors for tooltip
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
 * Build a lookup map from ID to table item
 * @param {Array} tableItems - Array of table items
 * @returns {Map} Map from ID to table item
 */
export function buildTableItemsMap(tableItems) {
  const map = new Map();
  tableItems.forEach((item) => {
    map.set(String(item.ID), item);
  });
  return map;
}

/**
 * Custom column rendering for Title with clickable link
 * data-id will be set by rowRender
 * @returns {Function} Render function
 */
export function createRenderTitle() {
  return function (data, cell, dataIndex, cellIndex) {
    const titleStr = String(data || '');
    // We need to find ID from the row. Use a custom attribute on the link
    // that will be updated by rowRender with the actual row ID
    cell.childNodes = [
      {
        nodeName: 'A',
        attributes: {
          class: 'title-link',
          href: '#',
          'data-id': '', // Will be filled by rowRender
          'data-title': titleStr,
        },
        childNodes: [{ nodeName: '#text', data: titleStr }],
      },
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
      childNodes: [{ nodeName: '#text', data: stateStr }],
    },
  ];
}

/**
 * Custom column rendering for Track with badge styling and clickable link
 * Badge class and data-tracks will be set by rowRender
 * @returns {Function} Render function
 */
export function createRenderTrack() {
  return function (data, cell, dataIndex, cellIndex) {
    const trackStr = String(data || '');
    if (!trackStr) return;

    // Default class, will be updated by rowRender based on TrackType
    cell.childNodes = [
      {
        nodeName: 'A',
        attributes: {
          class: 'track-badge track-reviewed track-link',
          href: '#',
          'data-tracks': '', // Will be filled by rowRender
        },
        childNodes: [{ nodeName: '#text', data: trackStr }],
      },
    ];
  };
}

/**
 * Custom column rendering for Affiliation with clickable link
 * data-affiliation will be set by rowRender
 * @returns {Function} Render function
 */
export function createRenderAffiliation() {
  return function (data, cell, dataIndex, cellIndex) {
    const affiliationStr = String(data || '');
    if (!affiliationStr) return;

    cell.childNodes = [
      {
        nodeName: 'A',
        attributes: {
          class: 'affiliation-link',
          href: '#',
          'data-affiliation': '', // Will be filled by rowRender
        },
        childNodes: [{ nodeName: '#text', data: affiliationStr }],
      },
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
      childNodes: [{ nodeName: '#text', data: typeStr }],
    },
  ];
}

/**
 * Row render to populate data attributes from row data
 * This runs after cell render functions, so we can access all row cells
 */
export function rowRender(row, tr, index) {
  // Get values from row cells
  const id = row.cells[0]?.data || ''; // ID column
  const affiliationFull = row.cells[5]?.data || ''; // AffiliationFull column
  const trackType = row.cells[7]?.data || ''; // TrackType column
  const trackFull = row.cells[6]?.data || '[]'; // TrackFull column
  const authorsTooltip = row.cells[13]?.data || ''; // AuthorsTooltip column

  // Update Title link (column 1) with data-id
  if (tr.childNodes && tr.childNodes[1]) {
    const titleCell = tr.childNodes[1];
    if (titleCell.childNodes && titleCell.childNodes[0]?.attributes) {
      titleCell.childNodes[0].attributes['data-id'] = String(id);
    }
  }

  // Update Affiliation link (column 4) with data-affiliation
  if (tr.childNodes && tr.childNodes[4]) {
    const affiliationCell = tr.childNodes[4];
    if (affiliationCell.childNodes && affiliationCell.childNodes[0]?.attributes) {
      affiliationCell.childNodes[0].attributes['data-affiliation'] = affiliationFull;
    }
  }

  // Update Track link (column 5) with data-tracks and proper class
  if (tr.childNodes && tr.childNodes[6]) {
    const trackCell = tr.childNodes[6];
    if (trackCell.childNodes && trackCell.childNodes[0]?.attributes) {
      trackCell.childNodes[0].attributes['data-tracks'] = trackFull;
      trackCell.childNodes[0].attributes.class =
        trackType === 'accepted'
          ? 'track-badge track-accepted track-link'
          : 'track-badge track-reviewed track-link';
    }
  }

  // Update Authors span (column 11) with tooltip
  if (tr.childNodes && tr.childNodes[12]) {
    const authorsCell = tr.childNodes[12];
    if (authorsCell.childNodes && authorsCell.childNodes[0]?.attributes) {
      authorsCell.childNodes[0].attributes.title = authorsTooltip;
    }
  }

  return tr;
}

/**
 * Custom column rendering for Authors with tooltip
 * Tooltip will be set by rowRender
 * @returns {Function} Render function
 */
export function createRenderAuthors() {
  return function (data, cell, dataIndex, cellIndex) {
    const authorsStr = String(data || '');
    if (!authorsStr) return;

    cell.childNodes = [
      {
        nodeName: 'SPAN',
        attributes: {
          class: 'authors-cell',
          title: '', // Will be filled by rowRender
        },
        childNodes: [{ nodeName: '#text', data: authorsStr }],
      },
    ];
  };
}

// Column names for filter placeholders (visible columns only)
const visibleColumnNames = [
  'ID',
  'Title',
  'State',
  'Submitter',
  'Affiliation',
  'Track',
  'Type',
  'Score',
  'Submitted',
  'Authors',
];

// Mapping from visible column index to actual data column index
// Visible: 0=ID, 1=Title, 2=State, 3=Submitter, 4=Affiliation, 5=Track, 6=Type, 7=Score, 8=Submitted, 9=Authors
// Data:    0=ID, 1=Title, 2=State, 3=Submitter, 4=Affiliation, 5=Track, 8=Type, 9=Score, 10=Submitted, 11=Authors
// Hidden data columns: 6=TrackFull, 7=TrackType, 12=AuthorsTooltip
const visibleToDataColumnIndex = [0, 1, 2, 3, 4, 5, 8, 9, 10, 11];

/**
 * Table render function to add column filtering row
 * @param {Array} _data - Table data
 * @param {Object} table - Table DOM structure
 * @param {string} type - Render type ('print' or default)
 * @returns {Object} Modified table structure
 */
export function tableRender(_data, table, type) {
  if (type === 'print') {
    return table;
  }

  const tHead = table.childNodes[0];
  const filterHeaders = {
    nodeName: 'TR',
    attributes: {
      class: 'search-filtering-row',
    },
    childNodes: tHead.childNodes[0].childNodes.map((_th, index) => ({
      nodeName: 'TH',
      childNodes: [
        {
          nodeName: 'INPUT',
          attributes: {
            class: 'datatable-input column-filter',
            type: 'search',
            placeholder: visibleColumnNames[index] || `Col ${index + 1}`,
            'data-columns': `[${visibleToDataColumnIndex[index]}]`,
          },
        },
      ],
    })),
  };

  tHead.childNodes.push(filterHeaders);
  return table;
}

/**
 * Create DataTable options with column customization
 * @returns {Object} DataTable options configuration
 */
export function createDataTableOptions() {
  return {
    searchable: true,
    sortable: true,
    paging: true,
    perPage: 25,
    perPageSelect: [10, 25, 50, 100],
    rowRender: rowRender,
    tableRender: tableRender,
    columns: [
      { select: 0, sortable: true, type: 'number' }, // ID
      { select: 1, render: createRenderTitle(), sortable: true, type: 'string' }, // Title
      { select: 2, render: renderState, sortable: true, type: 'string' }, // State
      { select: 3, sortable: true, type: 'string' }, // Submitter
      { select: 4, render: createRenderAffiliation(), sortable: true, type: 'string' }, // Affiliation
      { select: 5, hidden: true, type: 'string' }, // AffiliationFull (hidden - JSON of full affiliation)
      { select: 6, render: createRenderTrack(), sortable: true, type: 'string' }, // Track
      { select: 7, hidden: true, type: 'string' }, // TrackFull (hidden - JSON of all tracks)
      { select: 8, hidden: true, type: 'string' }, // TrackType (hidden helper column)
      { select: 9, render: renderType, sortable: true, type: 'string' }, // Type
      { select: 10, sortable: true, type: 'number' }, // Score
      { select: 11, sortable: true, type: 'string' }, // Submitted
      { select: 12, render: createRenderAuthors(), sortable: true, type: 'string' }, // Authors
      { select: 13, hidden: true, type: 'string' }, // AuthorsTooltip (hidden)
    ],
  };
}
