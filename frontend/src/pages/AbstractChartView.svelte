<script>
  import DonutChart from '../components/DonutChart.svelte';
  import BarChart from '../components/BarChart.svelte';
  import AbstractSubmissionTrend from '../components/AbstractSubmissionTrend.svelte';
  import ReviewChartView from '../components/ReviewChartView.svelte';
  import WordCloud from '../components/WordCloud.svelte';
  import { Tabs, TabItem, Button } from 'flowbite-svelte';
  import AffiliationTableView from '../components/AffiliationTableView.svelte';
  import AffiliationSettings from '../components/AffiliationSettings.svelte';
  import Icon from '@iconify/svelte';

  // Props: array of abstract objects (same shape returned by GetAbstracts)
  let { abstractData = [], visibilityConfig = null } = $props();

  // Affiliation deduplication state — aliasToCanonical is managed by AffiliationSettings
  let useAffiliationMap = $state(false);
  /**
   * Flat lookup: alias (raw name) → canonical (display name).
   * Managed by AffiliationSettings (loaded from config, kept live).
   * @type {Record<string, string>}
   */
  let aliasToCanonical = $state({});
  let showAffiliationSettings = $state(false);

  // Color schemes for charts
  const instituteColors = [
    '#1E40AF',
    '#7C3AED',
    '#DB2777',
    '#DC2626',
    '#EA580C',
    '#D97706',
    '#65A30D',
    '#059669',
    '#0891B2',
    '#0284C7',
  ];

  const countryColors = [
    '#3B82F6',
    '#8B5CF6',
    '#EC4899',
    '#EF4444',
    '#F97316',
    '#F59E0B',
    '#84CC16',
    '#10B981',
  ];

  const continentColors = ['#2563EB', '#9333EA', '#DB2777', '#DC2626', '#EA580C', '#D97706'];

  // Helper: collect affiliation entries from abstracts (ONLY from persons now)
  function collectAffiliations(data) {
    const items = [];
    for (const a of data || []) {
      // persons array only
      if (Array.isArray(a.persons)) {
        for (const p of a.persons) {
          if (p.affiliation) {
            // Handle both string and object affiliation data
            const affiliationStr =
              typeof p.affiliation === 'string'
                ? p.affiliation
                : p.affiliation.name || p.affiliation.raw || String(p.affiliation);

            items.push({
              raw: affiliationStr,
              country_name: p.affiliation?.country_name || p.country_name,
              continent: p.affiliation?.continent || p.continent,
            });
          }
        }
      }
    }
    return items;
  }

  // Aggregate helpers: by institute (name), by country_name, by continent
  function aggregateByInstitute(items, applyMap, lookup) {
    const m = new Map();
    for (const it of items) {
      const raw = it.raw || '';
      if (!raw) continue;
      const key = applyMap && lookup[raw] ? lookup[raw] : raw;
      m.set(key, (m.get(key) || 0) + 1);
    }
    return m;
  }

  function aggregateByCountry(items) {
    const m = new Map();
    for (const it of items) {
      const key = it.country_name || guessCountryFromRaw(it.raw) || '';
      if (!key) continue;
      m.set(key, (m.get(key) || 0) + 1);
    }
    return m;
  }

  function aggregateByContinent(items) {
    const m = new Map();
    for (const it of items) {
      const key =
        it.continent ||
        guessContinentFromCountry(it.country_name || guessCountryFromRaw(it.raw)) ||
        'Unknown';
      m.set(key, (m.get(key) || 0) + 1);
    }
    return m;
  }

  // Very small heuristics to guess country from raw affiliation string using last token if it matches known country names.
  // For reliability, prefer structured country_name if available on affiliation_link or backend enhanced data.
  const knownCountries = new Set([
    'United States',
    'USA',
    'United Kingdom',
    'Germany',
    'France',
    'Italy',
    'Spain',
    'Canada',
    'China',
    'Japan',
    'Switzerland',
    'Australia',
    'Brazil',
    'India',
    'South Korea',
    'Netherlands',
    'Sweden',
    'Belgium',
    'Austria',
    'Poland',
    'Russia',
    'Mexico',
    'Argentina',
  ]);

  function guessCountryFromRaw(raw) {
    if (!raw) return '';
    // Ensure raw is a string
    const rawStr = typeof raw === 'string' ? raw : String(raw);
    // try to find a known country substring
    for (const c of knownCountries) {
      if (rawStr.includes(c)) return c;
    }
    return '';
  }

  // Minimal mapping country -> continent for a few common countries; fallback to 'Other'
  const countryToContinent = {
    'United States': 'North America',
    USA: 'North America',
    Canada: 'North America',
    Mexico: 'North America',
    China: 'Asia',
    Japan: 'Asia',
    'South Korea': 'Asia',
    India: 'Asia',
    Germany: 'Europe',
    France: 'Europe',
    Italy: 'Europe',
    Spain: 'Europe',
    Switzerland: 'Europe',
    'United Kingdom': 'Europe',
    Netherlands: 'Europe',
    Sweden: 'Europe',
    Belgium: 'Europe',
    Austria: 'Europe',
    Poland: 'Europe',
    Russia: 'Europe',
    Australia: 'Oceania',
    Brazil: 'South America',
    Argentina: 'South America',
  };

  function guessContinentFromCountry(country) {
    if (!country) return 'Other';
    return countryToContinent[country] || 'Other';
  }

  function buildChartOptions(map, colors) {
    const labels = Array.from(map.keys());
    const values = Array.from(map.values());

    return {
      labels,
      series: values,
      colors: colors.slice(0, labels.length),
    };
  }

  // Keep top N for readability; group rest into "Other"
  function topN(m, n = 8) {
    const arr = Array.from(m.entries()).sort((a, b) => b[1] - a[1]);
    if (arr.length <= n) return new Map(arr);
    const top = arr.slice(0, n);
    const rest = arr.slice(n);
    const otherCount = rest.reduce((s, r) => s + r[1], 0);
    const res = new Map(top);
    if (otherCount > 0) {
      res.set('Other', otherCount);
    }
    return res;
  }

  // Build all three charts with derived state
  const chartData = $derived.by(() => {
    const items = collectAffiliations(abstractData || []);
    const instMap = aggregateByInstitute(items, useAffiliationMap, aliasToCanonical);
    const countryMap = aggregateByCountry(items);
    const continentMap = aggregateByContinent(items);

    const instTop = topN(instMap, 10);
    const countryTop = topN(countryMap, 8);
    const continentTop = topN(continentMap, 8);

    // Prepare full institute list sorted desc for bar chart
    const instFullArr = Array.from(instMap.entries()).sort((a, b) => b[1] - a[1]);
    const instFullLabels = instFullArr.map((e) => e[0]);
    const instFullSeries = instFullArr.map((e) => e[1]);

    return {
      institute: buildChartOptions(instTop, instituteColors),
      country: buildChartOptions(countryTop, countryColors),
      continent: buildChartOptions(continentTop, continentColors),
      instituteFull: { labels: instFullLabels, series: instFullSeries, colors: instituteColors },
    };
  });

  const instituteOptions = $derived(chartData.institute);
  const countryOptions = $derived(chartData.country);
  const continentOptions = $derived(chartData.continent);
  const instituteFullOptions = $derived(chartData.instituteFull);

  const chartHeight = '50vh';
</script>

<div class="p-2 mb-1">
  <Tabs class="shadow-md rounded-md">
    <TabItem open title="Affiliation">
      <div
        class="flex items-center justify-between gap-2 px-1 py-1 border-b border-gray-200 dark:border-gray-700"
      >
        <div class="flex items-center gap-2 last:-mt-6">
          <label class="flex items-center gap-2 text-sm">
            <input type="checkbox" bind:checked={useAffiliationMap} class="rounded" />
            <span>Use Deduplication</span>
          </label>
          <Button
            size="xs"
            color="light"
            onclick={() => (showAffiliationSettings = true)}
            title="Manage affiliation deduplication"
          >
            <Icon icon="mdi:cog" class="w-4 h-4" />
          </Button>
        </div>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-1 gap-4">
        <div>
          <Tabs tabStyle="underline">
            <TabItem open title="By Institution">
              <div class="p-0.5">
                {#if instituteOptions && instituteOptions.series && instituteOptions.series.length}
                  <div class="flex flex-col md:flex-row gap-0.5">
                    <div class="w-full md:w-3/5">
                      <DonutChart
                        labels={instituteOptions.labels}
                        series={instituteOptions.series}
                        colors={instituteOptions.colors}
                        title={''}
                        height={chartHeight}
                        legendPosition={'bottom'}
                      />
                    </div>

                    {#if instituteFullOptions && instituteFullOptions.series && instituteFullOptions.series.length}
                      <div class="w-full md:w-2/5">
                        <BarChart
                          labels={instituteFullOptions.labels}
                          series={instituteFullOptions.series}
                          colors={instituteFullOptions.colors}
                          title={''}
                          height={chartHeight}
                        />
                      </div>
                    {/if}
                  </div>
                {:else}
                  <div class="text-sm text-gray-500 text-center py-8">No data available</div>
                {/if}
              </div>
            </TabItem>

            <TabItem title="By Country">
              <div class="p-0.5">
                {#if countryOptions && countryOptions.series && countryOptions.series.length}
                  <DonutChart
                    labels={countryOptions.labels}
                    series={countryOptions.series}
                    colors={countryOptions.colors}
                    title={''}
                    height={chartHeight}
                  />
                {:else}
                  <div class="text-sm text-gray-500 text-center py-8">No data available</div>
                {/if}
              </div>
            </TabItem>

            <TabItem title="By Continent">
              <div class="p-0.5">
                {#if continentOptions && continentOptions.series && continentOptions.series.length}
                  <DonutChart
                    labels={continentOptions.labels}
                    series={continentOptions.series}
                    colors={continentOptions.colors}
                    title={''}
                    height={chartHeight}
                  />
                {:else}
                  <div class="text-sm text-gray-500 text-center py-8">No data available</div>
                {/if}
              </div>
            </TabItem>

            <TabItem title="Table">
              <div class="p-0.5 last:-mt-7 overflow-auto" style="height:calc(100vh - 18rem);">
                {#if abstractData && abstractData.length}
                  <AffiliationTableView
                    {abstractData}
                    bind:useAffiliationMap
                    {aliasToCanonical}
                    {visibilityConfig}
                  />
                {:else}
                  <div class="text-sm text-gray-500 text-center py-8">
                    No affiliation data available
                  </div>
                {/if}
              </div>
            </TabItem>
          </Tabs>
        </div>
      </div>
    </TabItem>

    {#if visibilityConfig?.ShowSubmissionTab !== false}
      <TabItem title="Submission">
        <div class="p-0.5 last:-mt-4">
          {#if abstractData && abstractData.length}
            <AbstractSubmissionTrend submittedTimes={abstractData} title={''} height={'40vh'} />
          {:else}
            <div class="text-sm text-gray-500 text-center py-8">No abstracts to display.</div>
          {/if}
        </div>
      </TabItem>
    {/if}

    <TabItem title="Reviews">
      <div class="p-0.5 last:-mt-8">
        {#if abstractData && abstractData.length}
          <ReviewChartView {abstractData} {visibilityConfig} />
        {:else}
          <div class="text-sm text-gray-500 text-center py-8">No abstracts to display.</div>
        {/if}
      </div>
    </TabItem>

    <TabItem title="Word Cloud">
      <div class="p-0.5 last:-mt-4">
        {#if abstractData && abstractData.length}
          <WordCloud abstracts={abstractData} minLength={3} maxWords={200} height="70vh" title="" />
        {:else}
          <div class="text-sm text-gray-500 text-center py-8">No abstracts to display.</div>
        {/if}
      </div>
    </TabItem>
  </Tabs>
</div>

<AffiliationSettings bind:open={showAffiliationSettings} bind:aliasToCanonical />
