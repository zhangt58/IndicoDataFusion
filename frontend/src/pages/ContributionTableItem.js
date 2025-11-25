/**
 * Helper to format date info object to string
 * @param {Object} dateInfo - The date info object with date, time, tz
 * @returns {string} Formatted date string
 */
export function formatDate(dateInfo) {
  if (!dateInfo) return '';
  const { date, time, tz } = dateInfo;
//   return tz ? `${date} ${time} (${tz})` : `${date} ${time}`;
  return `${date} ${time}`;
}

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
 * Get all speakers as a string for tooltip
 * @param {Array} speakers - Array of speaker objects
 * @returns {string} All speakers joined by newlines
 */
export function getSpeakersTooltip(speakers) {
  if (!speakers || speakers.length === 0) return '';
  return speakers.map(s => {
    const name = s.fullName || `${s.first_name} ${s.last_name}`;
    const affiliation = s.affiliation ? ` (${s.affiliation})` : '';
    return `🎤 ${name}${affiliation}`;
  }).join('\n');
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
    primaryauthors.forEach(a => {
      const name = a.fullName || `${a.first_name} ${a.last_name}`;
      const affiliation = a.affiliation ? ` (${a.affiliation})` : '';
      lines.push(`  ${name}${affiliation}`);
    });
  }
  
  if (coauthors && coauthors.length > 0) {
    if (lines.length > 0) lines.push('');
    lines.push('Co-Authors:');
    coauthors.forEach(a => {
      const name = a.fullName || `${a.first_name} ${a.last_name}`;
      const affiliation = a.affiliation ? ` (${a.affiliation})` : '';
      lines.push(`  ${name}${affiliation}`);
    });
  }
  
  return lines.join('\n');
}

/**
 * Transform a single contribution into a table row object
 * @param {Object} contribution - The contribution data object
 * @returns {Object} Table row data
 */
export function transformContributionToTableItem(contribution) {
  return {
    ID: contribution.friendly_id || contribution.id,
    Code: contribution.code || '',
    Title: contribution.title || '',
    Type: contribution.type || '',
    Session: contribution.session || '',
    Track: contribution.track || '',
    StartDate: formatDate(contribution.startDate),
    Duration: contribution.duration ? `${contribution.duration} min` : '',
    Location: contribution.location || '',
    Room: contribution.roomFullname || contribution.room || '',
    Speakers: getSpeakersDisplay(contribution.speakers),
    SpeakersTooltip: getSpeakersTooltip(contribution.speakers),
    Authors: getPrimaryAuthorsDisplay(contribution.primaryauthors),
    AuthorsTooltip: getAllAuthorsTooltip(contribution.primaryauthors, contribution.coauthors),
    URL: contribution.url || ''
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

/**
 * Build a lookup map from ID to table item
 * @param {Array} tableItems - Array of table items
 * @returns {Map} Map from ID to table item
 */
export function buildTableItemsMap(tableItems) {
  const map = new Map();
  tableItems.forEach(item => {
    map.set(String(item.ID), item);
  });
  return map;
}

/**
 * Custom column rendering for Title with clickable link
 * @returns {Function} Render function
 */
export function createRenderTitle() {
  return function(data, cell, dataIndex, cellIndex) {
    const titleStr = String(data || '');
    cell.childNodes = [
      {
        nodeName: 'A',
        attributes: { 
          class: 'title-link',
          href: '#',
          'data-id': '',
          'data-title': titleStr
        },
        childNodes: [{ nodeName: '#text', data: titleStr }]
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
 * Custom column rendering for Session with badge styling
 */
export function renderSession(data, cell, dataIndex, cellIndex) {
  const sessionStr = String(data || '');
  if (!sessionStr) return;
  
  cell.childNodes = [
    {
      nodeName: 'SPAN',
      attributes: { class: 'session-badge' },
      childNodes: [{ nodeName: '#text', data: sessionStr }]
    }
  ];
}

/**
 * Custom column rendering for Track with badge styling
 */
export function renderTrack(data, cell, dataIndex, cellIndex) {
  const trackStr = String(data || '');
  if (!trackStr) return;
  
  cell.childNodes = [
    {
      nodeName: 'SPAN',
      attributes: { class: 'track-badge' },
      childNodes: [{ nodeName: '#text', data: trackStr }]
    }
  ];
}

/**
 * Custom column rendering for Speakers with tooltip
 * @returns {Function} Render function
 */
export function createRenderSpeakers() {
  return function(data, cell, dataIndex, cellIndex) {
    const speakersStr = String(data || '');
    if (!speakersStr) return;
    
    cell.childNodes = [
      {
        nodeName: 'SPAN',
        attributes: { 
          class: 'speakers-cell',
          title: ''
        },
        childNodes: [{ nodeName: '#text', data: speakersStr }]
      }
    ];
  };
}

/**
 * Custom column rendering for Authors with tooltip
 * @returns {Function} Render function
 */
export function createRenderAuthors() {
  return function(data, cell, dataIndex, cellIndex) {
    const authorsStr = String(data || '');
    if (!authorsStr) return;
    
    cell.childNodes = [
      {
        nodeName: 'SPAN',
        attributes: { 
          class: 'authors-cell',
          title: ''
        },
        childNodes: [{ nodeName: '#text', data: authorsStr }]
      }
    ];
  };
}

/**
 * Row render to populate data attributes from row data
 */
export function rowRender(row, tr, index) {
  // Data column indices (row.cells):
  // 0=ID, 1=Code, 2=Title, 3=Type, 4=Session, 5=Track, 6=StartDate, 7=Duration,
  // 8=Location(hidden), 9=Room, 10=Speakers, 11=SpeakersTooltip(hidden), 12=Authors(hidden), 13=AuthorsTooltip(hidden), 14=URL(hidden)
  //
  // Visible column indices (tr.childNodes) - hidden columns are excluded:
  // 0=ID, 1=Code, 2=Title, 3=Type, 4=Session, 5=Track, 6=StartDate, 7=Duration,
  // 8=Room, 9=Speakers
  
  const id = row.cells[0]?.data || '';
  const speakersTooltip = row.cells[11]?.data || '';
  
  // Update Title link (visible column 2) with data-id
  if (tr.childNodes && tr.childNodes[2]) {
    const titleCell = tr.childNodes[2];
    if (titleCell.childNodes && titleCell.childNodes[0]?.attributes) {
      titleCell.childNodes[0].attributes['data-id'] = String(id);
    }
  }
  
  // Update Speakers span (visible column 9) with tooltip
  if (tr.childNodes && tr.childNodes[9]) {
    const speakersCell = tr.childNodes[9];
    if (speakersCell.childNodes && speakersCell.childNodes[0]?.attributes) {
      speakersCell.childNodes[0].attributes.title = speakersTooltip;
    }
  }
  
  return tr;
}

// Column names for filter placeholders (visible columns only)
// Visible: 0=ID, 1=Code, 2=Title, 3=Type, 4=Session, 5=Track, 6=Start, 7=Duration, 8=Room, 9=Speakers
const visibleColumnNames = ['ID', 'Code', 'Title', 'Type', 'Session', 'Track', 'Start', 'Duration', 'Room', 'Speakers'];

// Mapping from visible column index to actual data column index
// Visible: 0=ID, 1=Code, 2=Title, 3=Type, 4=Session, 5=Track, 6=Start, 7=Duration, 8=Room, 9=Speakers
// Data:    0=ID, 1=Code, 2=Title, 3=Type, 4=Session, 5=Track, 6=StartDate, 7=Duration, 9=Room, 10=Speakers
// Hidden data columns: 8=Location, 11=SpeakersTooltip, 12=Authors, 13=AuthorsTooltip, 14=URL
const visibleToDataColumnIndex = [0, 1, 2, 3, 4, 5, 6, 7, 9, 10];

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
      class: 'search-filtering-row'
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
            'data-columns': `[${visibleToDataColumnIndex[index]}]`
          }
        }
      ]
    }))
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
      { select: 0, sortable: true, type: 'number' },  // ID
      { select: 1, sortable: true, type: 'string' },  // Code
      { select: 2, render: createRenderTitle(), sortable: true, type: 'string' },  // Title
      { select: 3, render: renderType, sortable: true, type: 'string' },  // Type
      { select: 4, render: renderSession, sortable: true, type: 'string' },  // Session
      { select: 5, render: renderTrack, sortable: true, type: 'string' },  // Track
      { select: 6, sortable: true, type: 'string' },  // StartDate
      { select: 7, sortable: true, type: 'string' },  // Duration
      { select: 8, sortable: true, hidden: true, type: 'string' },  // Location
      { select: 9, sortable: true, type: 'string' },  // Room
      { select: 10, render: createRenderSpeakers(), sortable: true, type: 'string' },  // Speakers
      { select: 11, hidden: true, type: 'string' },  // SpeakersTooltip (hidden)
      { select: 12, hidden: true, render: createRenderAuthors(), sortable: true, type: 'string' },  // Authors
      { select: 13, hidden: true, type: 'string' },  // AuthorsTooltip (hidden)
      { select: 14, hidden: true, type: 'string' }   // URL (hidden)
    ]
  };
}
