<script lang="ts">
  import { onMount } from 'svelte';
  import * as d3 from 'd3';
  import { overviewStats, currentView, selectedCommit, selectedHotspotFile } from '../../stores/app';
  import SkeletonLoader from '../common/SkeletonLoader.svelte';

  let sparklineSvg: SVGSVGElement;

  $: if (sparklineSvg && $overviewStats?.timeline?.length) {
    renderVelocityChart();
  }

  function renderVelocityChart() {
    if (!sparklineSvg || !$overviewStats?.timeline?.length) return;

    const data = $overviewStats.timeline;
    const width = sparklineSvg.clientWidth || 600;
    const height = 180;
    const margin = { top: 20, right: 20, bottom: 30, left: 40 };

    const svg = d3.select(sparklineSvg);
    svg.selectAll('*').remove();

    svg.attr('viewBox', `0 0 ${width} ${height}`);

    const x = d3.scaleTime()
      .domain(d3.extent(data, d => new Date(d.date)) as [Date, Date])
      .range([margin.left, width - margin.right]);

    const y = d3.scaleLinear()
      .domain([0, (d3.max(data, d => d.commitCount) || 1) * 1.2])
      .nice()
      .range([height - margin.bottom, margin.top]);

    // Gradient
    const defs = svg.append('defs');
    const gradient = defs.append('linearGradient')
      .attr('id', 'area-gradient')
      .attr('x1', '0%').attr('y1', '0%')
      .attr('x2', '0%').attr('y2', '100%');
    gradient.append('stop').attr('offset', '0%').attr('stop-color', '#06b6d4').attr('stop-opacity', 0.4);
    gradient.append('stop').attr('offset', '100%').attr('stop-color', '#06b6d4').attr('stop-opacity', 0.0);

    // Grid lines
    svg.append('g')
      .attr('class', 'grid')
      .attr('transform', `translate(${margin.left},0)`)
      .call(d3.axisLeft(y).ticks(4).tickSize(-(width - margin.left - margin.right)).tickFormat(() => ''))
      .selectAll('line')
      .attr('stroke', '#1e293b')
      .attr('stroke-dasharray', '2,2');

    // Axes
    const xAxis = d3.axisBottom(x).ticks(5).tickSize(0);
    svg.append('g')
      .attr('transform', `translate(0,${height - margin.bottom})`)
      .call(xAxis)
      .selectAll('text')
      .attr('fill', '#64748b')
      .attr('font-size', '10px');

    const yAxis = d3.axisLeft(y).ticks(4).tickSize(0);
    svg.append('g')
      .attr('transform', `translate(${margin.left},0)`)
      .call(yAxis)
      .selectAll('text')
      .attr('fill', '#64748b')
      .attr('font-size', '10px');

    svg.selectAll('.domain').remove();

    // Area
    const area = d3.area<any>()
      .x(d => x(new Date(d.date)))
      .y0(height - margin.bottom)
      .y1(d => y(d.commitCount))
      .curve(d3.curveMonotoneX);

    svg.append('path')
      .datum(data)
      .attr('fill', 'url(#area-gradient)')
      .attr('d', area);

    // Line
    const line = d3.line<any>()
      .x(d => x(new Date(d.date)))
      .y(d => y(d.commitCount))
      .curve(d3.curveMonotoneX);

    svg.append('path')
      .datum(data)
      .attr('fill', 'none')
      .attr('stroke', '#06b6d4')
      .attr('stroke-width', 2)
      .attr('d', line);

    // Dots
    svg.selectAll('.dot')
      .data(data)
      .enter()
      .append('circle')
      .attr('cx', d => x(new Date(d.date)))
      .attr('cy', d => y(d.commitCount))
      .attr('r', 3)
      .attr('fill', '#06b6d4')
      .attr('stroke', '#0f172a')
      .attr('stroke-width', 1.5);
  }

  function viewHotspot(path: string) {
    selectedHotspotFile.set(path);
    currentView.set('hotspots');
  }

  function viewCommit(c: any) {
    selectedCommit.set(c);
    currentView.set('commits');
  }
</script>

<div class="h-full overflow-y-auto p-6 space-y-6">
  {#if !$overviewStats}
    <SkeletonLoader count={4} height="h-24" />
  {:else}
    <!-- Top KPI Grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="p-5 rounded-xl bg-slate-900/60 border border-slate-800/80 shadow-sm relative overflow-hidden group hover:border-slate-700/80 transition">
        <div class="flex items-center justify-between text-slate-400 mb-2">
          <span class="text-xs font-semibold uppercase tracking-wider">Total Commits</span>
          <span class="p-1.5 rounded-lg bg-cyan-950/60 text-cyan-400 border border-cyan-800/40">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </span>
        </div>
        <div class="text-2xl font-bold font-mono text-slate-100">
          {$overviewStats.repository.totalCommits.toLocaleString()}
        </div>
        <p class="text-[11px] text-slate-500 mt-1">Across all branches & tags</p>
      </div>

      <div class="p-5 rounded-xl bg-slate-900/60 border border-slate-800/80 shadow-sm relative overflow-hidden group hover:border-slate-700/80 transition">
        <div class="flex items-center justify-between text-slate-400 mb-2">
          <span class="text-xs font-semibold uppercase tracking-wider">Active Authors</span>
          <span class="p-1.5 rounded-lg bg-emerald-950/60 text-emerald-400 border border-emerald-800/40">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
          </span>
        </div>
        <div class="text-2xl font-bold font-mono text-slate-100">
          {$overviewStats.repository.totalAuthors}
        </div>
        <p class="text-[11px] text-slate-500 mt-1">Unique Git commit authors</p>
      </div>

      <div class="p-5 rounded-xl bg-slate-900/60 border border-slate-800/80 shadow-sm relative overflow-hidden group hover:border-slate-700/80 transition">
        <div class="flex items-center justify-between text-slate-400 mb-2">
          <span class="text-xs font-semibold uppercase tracking-wider">Tracked Files</span>
          <span class="p-1.5 rounded-lg bg-indigo-950/60 text-indigo-400 border border-indigo-800/40">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </span>
        </div>
        <div class="text-2xl font-bold font-mono text-slate-100">
          {$overviewStats.repository.totalFiles.toLocaleString()}
        </div>
        <p class="text-[11px] text-slate-500 mt-1">Source & asset inventory</p>
      </div>

      <div class="p-5 rounded-xl bg-slate-900/60 border border-slate-800/80 shadow-sm relative overflow-hidden group hover:border-slate-700/80 transition">
        <div class="flex items-center justify-between text-slate-400 mb-2">
          <span class="text-xs font-semibold uppercase tracking-wider">AST Dependencies</span>
          <span class="p-1.5 rounded-lg bg-purple-950/60 text-purple-400 border border-purple-800/40">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101" />
            </svg>
          </span>
        </div>
        <div class="text-2xl font-bold font-mono text-slate-100">
          {$overviewStats.totalDependencies}
        </div>
        <p class="text-[11px] text-slate-500 mt-1">Parsed cross-file imports</p>
      </div>
    </div>

    <!-- Velocity Timeline Chart -->
    <div class="p-5 rounded-xl bg-slate-900/60 border border-slate-800/80 shadow-sm">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h3 class="text-sm font-semibold text-slate-100">Commit Velocity & Activity Trends</h3>
          <p class="text-xs text-slate-500">Historical commit frequency over time</p>
        </div>
        <button
          class="text-xs font-medium text-cyan-400 hover:text-cyan-300 transition"
          on:click={() => currentView.set('timeline')}
        >
          View Full Timeline →
        </button>
      </div>
      <div class="w-full h-44">
        <svg bind:this={sparklineSvg} class="w-full h-full"></svg>
      </div>
    </div>

    <!-- Bottom Split: Top Hotspots vs Recent Activity -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Top Code Hotspots -->
      <div class="p-5 rounded-xl bg-slate-900/60 border border-slate-800/80 shadow-sm">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-sm font-semibold text-slate-100">Top Code Hotspots (High Churn)</h3>
          <button
            class="text-xs font-medium text-cyan-400 hover:text-cyan-300 transition"
            on:click={() => currentView.set('hotspots')}
          >
            All Hotspots →
          </button>
        </div>
        <div class="divide-y divide-slate-800/60">
          {#if $overviewStats.topHotspots.length === 0}
            <p class="py-6 text-center text-xs text-slate-500">No hotspot metrics computed yet.</p>
          {:else}
            {#each $overviewStats.topHotspots.slice(0, 5) as h}
              <div
                role="button"
                tabindex="0"
                class="py-2.5 flex items-center justify-between hover:bg-slate-800/30 px-2 rounded cursor-pointer transition"
                on:click={() => viewHotspot(h.filePath)}
                on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ') viewHotspot(h.filePath); }}
              >
                <div class="min-w-0 pr-3">
                  <span class="block text-xs font-mono text-slate-200 truncate">{h.filePath}</span>
                  <span class="text-[10px] text-slate-500">
                    {h.commitCount} commits • {h.authorCount} authors • {h.linesOfCode} LOC
                  </span>
                </div>
                <div class="shrink-0 flex items-center gap-2">
                  <span class="px-2 py-0.5 rounded text-[10px] font-mono font-medium bg-rose-950/60 text-rose-400 border border-rose-800/50">
                    Churn: {h.churnScore.toFixed(1)}
                  </span>
                </div>
              </div>
            {/each}
          {/if}
        </div>
      </div>

      <!-- Recent Commits -->
      <div class="p-5 rounded-xl bg-slate-900/60 border border-slate-800/80 shadow-sm">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-sm font-semibold text-slate-100">Recent Commits</h3>
          <button
            class="text-xs font-medium text-cyan-400 hover:text-cyan-300 transition"
            on:click={() => currentView.set('commits')}
          >
            Commit Explorer →
          </button>
        </div>
        <div class="divide-y divide-slate-800/60">
          {#if $overviewStats.recentCommits.length === 0}
            <p class="py-6 text-center text-xs text-slate-500">No commits found.</p>
          {:else}
            {#each $overviewStats.recentCommits.slice(0, 5) as c}
              <div
                role="button"
                tabindex="0"
                class="py-2.5 flex items-center justify-between hover:bg-slate-800/30 px-2 rounded cursor-pointer transition"
                on:click={() => viewCommit(c)}
                on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ') viewCommit(c); }}
              >
                <div class="min-w-0 pr-3">
                  <span class="block text-xs font-medium text-slate-200 truncate">{c.subject}</span>
                  <div class="flex items-center gap-2 text-[10px] text-slate-500 mt-0.5">
                    <span class="font-mono text-cyan-400">{c.shortHash}</span>
                    <span>•</span>
                    <span>{c.authorName}</span>
                    <span>•</span>
                    <span>{new Date(c.authorDate).toLocaleDateString()}</span>
                  </div>
                </div>
                <div class="shrink-0 text-right text-[11px] font-mono">
                  <span class="text-emerald-400">+{c.insertions}</span>
                  <span class="text-rose-400 ml-1">-{c.deletions}</span>
                </div>
              </div>
            {/each}
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>
