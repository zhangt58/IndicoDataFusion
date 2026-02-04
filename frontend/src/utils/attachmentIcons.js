// Utility to map attachment metadata to Iconify icon IDs and Tailwind color/bg classes
export function getAttachmentIcon(attachment) {
  const contentType =
    (attachment && (attachment.content_type || attachment.contentType || '')) + '';
  const filename = attachment && (attachment.filename || attachment.title || '') + '';
  const ext = (filename.split('.').pop() || '').toLowerCase();

  // default
  let icon = 'mdi:file';
  let color = 'text-gray-600 dark:text-gray-400';
  let bgColor = 'bg-gray-50 dark:bg-gray-700/50';

  if (contentType.includes('pdf') || ext === 'pdf') {
    icon = 'mdi:file-pdf';
    color = 'text-red-600 dark:text-red-400';
    bgColor = 'bg-red-50 dark:bg-red-900/20';
    return { icon, color, bgColor };
  }
  if (contentType.includes('image') || ['jpg', 'jpeg', 'png', 'gif', 'svg', 'webp'].includes(ext)) {
    icon = 'mdi:file-image';
    color = 'text-green-600 dark:text-green-400';
    bgColor = 'bg-green-50 dark:bg-green-900/20';
    return { icon, color, bgColor };
  }
  if (
    contentType.includes('word') ||
    contentType.includes('msword') ||
    ['doc', 'docx'].includes(ext)
  ) {
    icon = 'mdi:file-document';
    color = 'text-blue-600 dark:text-blue-400';
    bgColor = 'bg-blue-50 dark:bg-blue-900/20';
    return { icon, color, bgColor };
  }
  if (
    contentType.includes('excel') ||
    contentType.includes('spreadsheet') ||
    ['xls', 'xlsx', 'csv'].includes(ext)
  ) {
    icon = 'mdi:table';
    color = 'text-emerald-600 dark:text-emerald-400';
    bgColor = 'bg-emerald-50 dark:bg-emerald-900/20';
    return { icon, color, bgColor };
  }
  if (
    contentType.includes('presentation') ||
    contentType.includes('powerpoint') ||
    ['ppt', 'pptx'].includes(ext)
  ) {
    icon = 'mdi:file-powerpoint';
    color = 'text-orange-600 dark:text-orange-400';
    bgColor = 'bg-orange-50 dark:bg-orange-900/20';
    return { icon, color, bgColor };
  }
  if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) {
    icon = 'mdi:archive';
    color = 'text-purple-600 dark:text-purple-400';
    bgColor = 'bg-purple-50 dark:bg-purple-900/20';
    return { icon, color, bgColor };
  }
  if (contentType.includes('video') || ['mp4', 'avi', 'mov', 'wmv', 'flv', 'mkv'].includes(ext)) {
    icon = 'mdi:video';
    color = 'text-pink-600 dark:text-pink-400';
    bgColor = 'bg-pink-50 dark:bg-pink-900/20';
    return { icon, color, bgColor };
  }

  return { icon, color, bgColor };
}
