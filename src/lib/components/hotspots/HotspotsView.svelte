<script lang="ts">
  import { onMount } from 'svelte';
  import * as d3 from 'd3';
  import type { FileHotspot } from '../../types';
  import { selectedRepo, selectedHotspotFile, currentView, fileFilter } from '../../stores/app';
  import { api } from '../../api/client';
  import SkeletonLoader from '../common/SkeletonLoader.svelte';

  let hotspots: FileHotspot[] = [];
  let loading = true;
  let activeTab: 'matrix' | 'treemap' = 'treemap';
  let selectedExtension = 'all';
  let searchTerm = '';
  let treemapSvg: SVGSVGElement;
  let matrixSvg: SVGSVGElement;
  let hoveredHotspot: FileHotspot | null = null;
  let tooltipX = 0;
  let tooltipY = 0;

  onMount(async () => {
    await fetchHotspots();
  });

  $: if ($selectedRepo?.id) {
    fetchHotspots();
  }

  async function fetchHotspots() {
    if (!$selectedRepo) return;
    loading = true;
    try {
      hotspots = await api.getHotspots($selectedRepo.id, 100);
      renderActiveChart();
    } catch (err) {
      console.error('Failed to load hotspots:', err);
    } finally {
      loading = false;
    }
  }

  $: filteredHotspots = hotspots.filter(h => {
    const matchExt = selectedExtension === 'all' || h.extension === selectedExtension;
    const matchSearch = !searchTerm || h.filePath.toLowerCase().includes(searchTerm.toLowerCase());
    return matchExt && matchSearch;
  });

  $: if (filteredHotspots && (treemapSvg || matrixSvg)) {
    renderActiveChart();
  }

  function renderActiveChart() {
    if (activeTab === 'treemap') {
      renderTreemap();
    } else {
      renderScatterMatrix();
    }
  }

  function renderTreemap() {
    if (!treemapSvg || filteredHotspots.length === 0) return;

    const width = treemapSvg.clientWidth || 800;
    const height = 460;

    const svg = d3.select(treemapSvg);
    svg.selectAll('*').remove();
    svg.attr('viewBox', `0 0 ${width} ${height}`);

    const rootData = {
      name: 'root',
      children: filteredHotspots.map(h => ({
        ...h,
        value: Math.max(h.linesOfCode, 10)
      }))
    };

    const root = d3.hierarchy<any>(rootData)
      .sum(d => d.value)
      .sort((a, b) => (b.value || 0) - (a.value || 0));

    d3.treemap<any>()
      .size([width, height])
      .padding(3)
      .round(true)(root);

    const maxChurn = d3.max(filteredHotspots, d => d.churnScore) || 10;
    const colorScale = d3.scaleSequential(d3.interpolateWarm)
      .domain([0, maxChurn]);

    const leaf = svg.selectAll('g')
      .data(root.leaves())
      .enter()
      .append('g')
      .attr('transform', d => `translate(${d.x0},${d.y0})`)
      .style('cursor', 'pointer')
      .on('mouseenter', (event, d) => {
        hoveredHotspot = d.data as FileHotspot;
        tooltipX = event.clientX;
        tooltipY = event.clientY;
      })
      .on('mousemove', (event) => {
        tooltipX = event.clientX;
        tooltipY = event.clientY;
      })
      .on('mouseleave', () => {
        hoveredHotspot = null;
      })
      .on('click', (_, d) => {
        drillDownFile(d.data.filePath);
      });

    leaf.append('rect')
      .attr('width', d => Math.max(0, d.x1 - d.x0))
      .attr('height', d => Math.max(0, d.y1 - d.y0))
      .attr('fill', d => colorScale(d.data.churnScore))
      .attr('fill-opacity', 0.85)
      .attr('stroke', '#0f172a')
      .attr('stroke-width', 1.5)
      .attr('rx', 4);

    leaf.append('text')
      .attr('x', 6)
      .attr('y', 16)
      .attr('font-size', '10px')
      .attr('font-family', 'monospace')
      .attr('font-weight', '600')
      .attr('fill', '#ffffff')
      .text(d => {
        const w = d.x1 - d.x0;
        const name = d.data.filePath.split('/').pop() || '';
        return w > 50 ? name : '';
      });

    leaf.append('text')
      .attr('x', 6)
      .attr('y', 28)
      .attr('font-size', '9px')
      .attr('fill', '#e2e8f0')
      .text(d => {
        const w = d.x1 - d.x0;
        const h = d.y1 - d.y0;
        return w > 60 && h > 35 ? `${d.data.commitCount} commits` : '';
      });
  }

  function renderScatterMatrix() {
    if (!matrixSvg || filteredHotspots.length === 0) return;

    const width = matrixSvg.clientWidth || 800;
    const height = 460;
    const margin = { top: 30, right: 30, bottom: 50, left: 60 };

    const svg = d3.select(matrixSvg);
    svg.selectAll('*').remove();
    svg.attr('viewBox', `0 0 ${width} ${height}`);

    const maxLOC = (d3.max(filteredHotspots, d => d.linesOfCode) || 100) * 1.1;
    const maxCommits = (d3.max(filteredHotspots, d => d.commitCount) || 10) * 1.2;

    const x = d3.scaleLinear()
      .domain([0, maxLOC])
      .range([margin.left, width - margin.right]);

    const y = d3.scaleLinear()
      .domain([0, maxCommits])
      .range([height - margin.bottom, margin.top]);

    const r = d3.scaleSqrt()
      .domain([0, d3.max(filteredHotspots, d => d.authorCount) || 5])
      .range([4, 16]);

    // Grid
    svg.append('g')
      .attr('transform', `translate(0, ${height - margin.bottom})`)
      .call(d3.axisBottom(x).ticks(6))
      .selectAll('text').attr('fill', '#64748b').attr('font-size', '10px');

    svg.append('g')
      .attr('transform', `translate(${margin.left}, 0)`)
      .call(d3.axisLeft(y).ticks(5))
      .selectAll('text').attr('fill', '#64748b').attr('font-size', '10px');

    svg.selectAll('.domain').attr('stroke', '#334155');

    // Axis labels
    svg.append('text')
      .attr('x', width / 2)
      .attr('y', height - 15)
      .attr('text-anchor', 'middle')
      .attr('fill', '#94a3b8')
      .attr('font-size', '11px')
      .text('Lines of Code (Complexity)');

    svg.append('text')
      .attr('transform', 'rotate(-90)')
      .attr('x', -height / 2)
      .attr('y', 20)
      .attr('text-anchor', 'middle')
      .attr('fill', '#94a3b8')
      .attr('font-size', '11px')
      .text('Modification Frequency (Commits)');

    const circles = svg.selectAll('.bubble')
      .data(filteredHotspots)
      .enter()
      .append('circle')
      .attr('cx', d => x(d.linesOfCode))
      .attr('cy', d => y(d.commitCount))
      .attr('r', d => r(d.authorCount))
      .attr('fill', '#f43f5e')
      .attr('fill-opacity', 0.6)
      .attr('stroke', '#f43f5e')
      .attr('stroke-width', 1.5)
      .style('cursor', 'pointer')
      .on('mouseenter', (event, d) => {
        hoveredHotspot = d;
        tooltipX = event.clientX;
        tooltipY = event.clientY;
      })
      .on('mousemove', (event) => {
        tooltipX = event.clientX;
        tooltipY = event.clientY;
      })
      .on('mouseleave', () => {
        hoveredHotspot = null;
      })
      .on('click', (_, d) => {
        drillDownFile(d.filePath);
      });
  }

  function drillDownFile(path: string) {
    selectedHotspotFile.set(path);
    fileFilter.set(path);
    currentView.set('commits');
  }

  $: extensions = ['all', ...Array.from(new Set(hotspots.map(h => h.extension))).filter(Boolean)];
</script>

<div class="flex flex-col h-full overflow-hidden p-6 space-y-4">
  <!-- Header & Toolbar -->
  <div class="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-slate-800/80">
    <div>
      <h2 class="text-base font-semibold text-slate-100">Code Hotspots & Churn Analysis</h2>
      <p class="text-xs text-slate-400">
        High churn and complexity identify files susceptible to architectural debt and defect clustering.
      </p>
    </div>

    <div class="flex items-center gap-3">
      <!-- Tab Switcher -->
      <div class="flex items-center p-0.5 rounded-lg bg-slate-900 border border-slate-800 text-xs">
        <button
          class="px-3 py-1.5 rounded-md font-medium transition {activeTab === 'treemap' ? 'bg-cyan-500 text-slate-950 shadow-sm' : 'text-slate-400 hover:text-slate-200'}"
          on:click={() => { activeTab = 'treemap'; renderActiveChart(); }}
        >
          Treemap Heatmap
        </button>
        <button
          class="px-3 py-1.5 rounded-md font-medium transition {activeTab === 'matrix' ? 'bg-cyan-500 text-slate-950 shadow-sm' : 'text-slate-400 hover:text-slate-200'}"
          on:click={() => { activeTab = 'matrix'; renderActiveChart(); }}
        >
          Churn vs LOC Matrix
        </button>
      </div>

      <!-- Extension filter -->
      <select
        bind:value={selectedExtension}
        class="px-2.5 py-1.5 text-xs bg-slate-900 border border-slate-700/80 rounded-lg text-slate-200 font-sans focus:outline-none"
      >
        {#each extensions as ext}
          <option value={ext}>{ext === 'all' ? 'All Extensions' : ext}</option>
        {/each}
      </select>
    </div>
  </div>

  {#if loading}
    <SkeletonLoader count={4} height="h-24" />
  {:else if filteredHotspots.length === 0}
    <div class="flex-1 flex items-center justify-center text-slate-500 text-sm">
      No files match the selected filter.
    </div>
  {:else}
    <!-- Chart Canvas -->
    <div class="flex-1 w-full bg-slate-950/60 rounded-xl border border-slate-800/80 p-4 relative overflow-hidden shadow-inner">
      {#if activeTab === 'treemap'}
        <svg bind:this={treemapSvg} class="w-full h-full"></svg>
      {:else}
        <svg bind:this={matrixSvg} class="w-full h-full"></svg>
      {/if}
    </div>
  {/if}
</div>

<!-- Floating Tooltip -->
{#if hoveredHotspot}
  <div
    class="archeo-tooltip"
    style="left: {tooltipX + 15}px; top: {tooltipY + 15}px;"
  >
    <div class="font-mono text-xs font-semibold text-cyan-300 mb-1">{hoveredHotspot.filePath}</div>
    <div class="grid grid-cols-2 gap-x-3 gap-y-1 text-[11px] text-slate-300">
      <span>Churn Score:</span>
      <strong class="text-rose-400">{hoveredHotspot.churnScore.toFixed(1)}</strong>
      <span>Commits:</span>
      <strong>{hoveredHotspot.commitCount}</strong>
      <span>Authors:</span>
      <strong>{hoveredHotspot.authorCount}</strong>
      <span>Lines of Code:</span>
      <strong>{hoveredHotspot.linesOfCode}</strong>
      <span>Line Changes:</span>
      <span class="text-emerald-400">+{hoveredHotspot.totalInsertions} / <span class="text-rose-400">-{hoveredHotspot.totalDeletions}</span></span>
    </div>
    <div class="mt-2 text-[10px] text-slate-500 italic">Click to view commit diffs for this file</div>
  </div>
{/if}
