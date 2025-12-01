// Cache refresh utilities
import { RefreshCache, GetCacheEntries } from '../../wailsjs/go/main/App';

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
 * Checks whether a given cache key is present in any data-source group.
 * Returns true when the key exists (i.e. cached and not expired), false otherwise.
 * This wraps the frontend Wails API GetCacheEntries which returns a map of groups -> CacheEntry[]
 */
export async function isCacheKeyPresent(cacheKey) {
  try {
    const entriesMap = await GetCacheEntries();
    for (const groupKey of Object.keys(entriesMap || {})) {
      const group = entriesMap[groupKey];
      if (!Array.isArray(group)) continue;
      for (const entry of group) {
        if (entry && entry.key === cacheKey) {
          return true;
        }
      }
    }
    return false;
  } catch (e) {
    console.error('isCacheKeyPresent failed', e);
    return false;
  }
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

    // Only reload when the cache was explicitly refreshed for this key, or when cache was cleared.
    // Do NOT auto-reload on expiration/eviction events (action === 'expired' or 'evicted').
    const shouldReload = (action === 'refreshed' && key === cacheKey) || action === 'cleared';

    if (!shouldReload) {
      // If this event is not a user-triggered refresh/clear, ignore for loading purposes.
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

  async function updateCacheStatus() {
    try {
      const present = await isCacheKeyPresent(cacheKey);
      return !present; // return true when expired
    } catch (e) {
      // Keep error logging minimal here; caller may log if needed
      console.error('createCachePage.updateCacheStatus failed', e);
      return true; // assume expired on error
    }
  }

  return {
    handleRefresh,
    handleCacheEvent,
    updateCacheStatus,
  };
}
