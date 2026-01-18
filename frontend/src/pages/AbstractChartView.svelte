<script>
  import AffiliationDonut from '../components/AffiliationDonut.svelte';
  import AffiliationBarChart from '../components/AffiliationBarChart.svelte';
  import AbstractSubmissionTrend from '../components/AbstractSubmissionTrend.svelte';
  import { Tabs, TabItem } from 'flowbite-svelte';

  // Props: array of abstract objects (same shape returned by GetAbstracts)
  let { abstractData = [] } = $props();

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
              affiliation: affiliationStr,
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
  function aggregateByInstitute(items) {
    const m = new Map();
    for (const it of items) {
      const key = it.affiliation || it.raw || '';
      if (!key) continue;
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
      colors: colors.slice(0, labels.length)
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
    const instMap = aggregateByInstitute(items);
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

  // Shared chart height for the paired donut/bar display.
  // Use a non-pixel unit so layout remains responsive. Adjust as needed.
  const chartHeight = '50vh';
</script>

<div class="p-2 mb-1">
  <!-- Top-level tabs: Affiliation vs Submission -->
  <Tabs class="shadow-md rounded-md">
    <TabItem open title="Affiliation">
      <div class="grid grid-cols-1 md:grid-cols-1 gap-4">
        <div>
          <Tabs tabStyle="underline">
            <TabItem open title="By Institution">
              <div class="p-0.5">
                {#if instituteOptions && instituteOptions.series && instituteOptions.series.length}
                  <div class="flex flex-col md:flex-row gap-0.5">
                    <div class="w-full md:w-3/5">
                      <AffiliationDonut
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
                        <AffiliationBarChart
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
                  <AffiliationDonut
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
                  <AffiliationDonut
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
          </Tabs>
        </div>
      </div>
    </TabItem>

    <TabItem title="Submission">
      <!-- Submission content -->
      <div class="p-0.5">
        {#if abstractData && abstractData.length}
          <AbstractSubmissionTrend submittedTimes={abstractData} title={''} height={'40vh'} />
        {:else}
          <div class="text-sm text-gray-500 text-center py-8">No abstracts to display.</div>
        {/if}
      </div>
    </TabItem>
  </Tabs>
</div>
