/**
 * Shared reactive vote stats store (Svelte 5 rune module).
 *
 * Consumers (AbstractMyReviews, AbstractReviewForm, etc.) import
 * `voteStatsStore` and read `voteStatsStore.data` reactively.
 * Any component can call `voteStatsStore.refresh()` to re-fetch and
 * the update is instantly visible everywhere.
 */
import { GetVoteStats } from '../../wailsjs/go/main/App';

function createVoteStatsStore() {
  let data = $state(null);
  let loading = $state(false);

  async function refresh() {
    if (loading) return;
    loading = true;
    try {
      data = await GetVoteStats();
    } catch {
      // vote stats are purely informational; ignore errors silently
    } finally {
      loading = false;
    }
  }

  return {
    get data() { return data; },
    set data(v) { data = v; },
    get loading() { return loading; },
    refresh,
  };
}

export const voteStatsStore = createVoteStatsStore();

