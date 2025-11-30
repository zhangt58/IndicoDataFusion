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
    // Check if this event is for our cache key or a clear all action
    if (data.key === cacheKey || data.action === 'cleared') {
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
    }
  };
}

