// Cache refresh utilities
import { RefreshCache } from '../../wailsjs/go/main/App';

/**
 * Creates a refresh handler for a specific cache key
 * @param {string} cacheKey - The cache key to refresh (e.g., 'event_info', 'abstracts', 'contributions')
 * @param {Function} setRefreshing - Callback to set refreshing state
 * @param {Function} setError - Callback to set error state
 * @returns {Function} - The refresh handler function
 */
export function createRefreshHandler(cacheKey, setRefreshing, setError) {
  return async function handleRefresh() {
    setRefreshing(true);
    try {
      await RefreshCache(cacheKey);
      // loadData() will be triggered by the cache:updated event
      // Don't call it here to avoid double loading
    } catch (e) {
      console.error(`Refresh failed for ${cacheKey}:`, e);
      setError(e);
      setRefreshing(false); // Clear on error
    }
    // Note: refreshing flag will be cleared by the event listener
  };
}

/**
 * Creates a cache event listener that prevents double loading
 * @param {string} cacheKey - The cache key to listen for
 * @param {Function} loadData - The function to call to reload data
 * @param {Function} setRefreshing - Callback to clear refreshing state
 * @returns {Function} - The event listener function
 */
export function createCacheEventListener(cacheKey, loadData, setRefreshing) {
  let isLoading = false; // Flag to prevent concurrent loads

  return async function handleCacheEvent(data) {
    // data is expected to be an object like { key: "abstracts", action: "refreshed" }
    const action = data && data.action ? data.action : null;
    const key = data && data.key ? data.key : null;

    // Only reload when the cache was explicitly refreshed/deleted for this key, or when cache was cleared.
    // Do NOT auto-reload on expiration/eviction events (action === 'expired' or 'evicted').
    const shouldReload = (action === 'refreshed' && key === cacheKey) ||
                        (action === 'deleted' && key === cacheKey) ||
                        action === 'cleared';

    if (!shouldReload) {
      // If this event is not a user-triggered refresh/clear/delete, ignore for loading purposes.
      // Optionally, we might still want to clear the refreshing flag if action indicates failure —
      // but in normal flows we only clear refreshing when a refresh completed (refreshed event).
      return;
    }

    // Prevent concurrent loads
    if (isLoading) {
      console.log(`Skipping redundant load for ${cacheKey} - already loading`);
      return;
    }

    isLoading = true;
    try {
      await loadData();
      setRefreshing(false); // Clear refresh indicator
    } finally {
      isLoading = false;
    }
  };
}

// Shared helper for cache-aware pages
export function createCachePage(cacheKey, loadData, setRefreshing, setError) {
  const handleRefresh = createRefreshHandler(cacheKey, setRefreshing, setError);
  const handleCacheEvent = createCacheEventListener(cacheKey, loadData, setRefreshing);

  return {
    handleRefresh,
    handleCacheEvent,
  };
}
