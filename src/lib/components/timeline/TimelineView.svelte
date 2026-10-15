<script lang="ts">
  import { onMount, afterUpdate } from 'svelte';
  import * as d3 from 'd3';
  import type { Commit } from '../../types';
  import { selectedRepo, selectedCommit, currentView } from '../../stores/app';
  import { api } from '../../api/client';
  import SkeletonLoader from '../common/SkeletonLoader.svelte';

  let container: HTMLDivElement;
  let svgElement: SVGSVGElement;
  let commits: Commit[] = [];
  let loading = true;
  let activeCommit: Commit | null = null;
  let authorColorMap = new Map<string, string>();

  const colorPalette = [
    '#06b6d4', '#10b981', '#f59e0b', '#a855f7',
    '#3b82f6', '#ec4899', '#14b8a6', '#8b5cf6'
  ];

  onMount(async () => {
    await fetchCommits();
  });

  $: if ($selectedRepo?.id) {
    fetchCommits();
  }

  async function fetchCommits() {
    if (!$selectedRepo) return;
    loading = true;
    try {
      commits = await api.getCommits({ repoId: $selectedRepo.id, limit: 150 });
      if (commits.length > 0 && !activeCommit) {
        activeCommit = commits[0];
      }
      buildColorMap();
      renderTimeline();
    } catch (err) {
      console.error('Failed to load commits for timeline:', err);
    } finally {
      loading = false;
    }
  }

  function buildColorMap() {
    let colorIdx = 0;
    for (const c of commits) {
      if (!authorColorMap.has(c.authorEmail)) {
        authorColorMap.set(c.authorEmail, colorPalette[colorIdx % colorPalette.length]);
        colorIdx++;
      }
    }
  }

  function renderTimeline() {
    if (!svgElement || commits.length === 0) return;

    const width = container?.clientWidth || 900;
    const height = 480;
    const margin = { top: 40, right: 40, bottom: 60, left: 60 };

    const svg = d3.select(svgElement);
    svg.selectAll('*').remove();
    svg.attr('viewBox', `0 0 ${width} ${height}`);

    const g = svg.append('g');

    // Zoom behavior
    const zoom = d3.zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.5, 8])
      .translateExtent([[0, 0], [width * 2, height]])
      .on('zoom', (event) => {
        g.attr('transform', event.transform);
      });

    svg.call(zoom);

    // Time scale
    const dates = commits.map(c => new Date(c.authorDate));
    const [minDate, maxDate] = d3.extent(dates) as [Date, Date];
    const paddedMin = new Date(minDate.getTime() - 86400000 * 2);
    const paddedMax = new Date(maxDate.getTime() + 86400000 * 2);

    const xScale = d3.scaleTime()
      .domain([paddedMin, paddedMax])
      .range([margin.left, width - margin.right]);

    // Distribute commits on swimlanes by author
    const authors = Array.from(authorColorMap.keys());
    const yScale = d3.scalePoint()
      .domain(authors)
      .range([margin.top + 30, height - margin.bottom - 30])
      .padding(0.5);

    // Grid lines for authors (swimlanes)
    authors.forEach(author => {
      const yPos = yScale(author) || 0;
      g.append('line')
        .attr('x1', margin.left)
        .attr('x2', width - margin.right)
        .attr('y1', yPos)
        .attr('y2', yPos)
        .attr('stroke', '#1e293b')
        .attr('stroke-dasharray', '4,4');

      // Author label
      const name = commits.find(c => c.authorEmail === author)?.authorName || author;
      g.append('text')
        .attr('x', margin.left - 10)
        .attr('y', yPos + 4)
        .attr('text-anchor', 'end')
        .attr('fill', '#64748b')
        .attr('font-size', '11px')
        .attr('font-family', 'sans-serif')
        .text(name.split(' ')[0]);
    });

    // Time Axis
    const xAxis = d3.axisBottom(xScale).ticks(6).tickSize(6);
    g.append('g')
      .attr('transform', `translate(0, ${height - margin.bottom})`)
      .call(xAxis)
      .selectAll('text')
      .attr('fill', '#94a3b8')
      .attr('font-size', '11px');

    g.selectAll('.domain').attr('stroke', '#334155');

    // Create lookup map for parent hashes
    const commitMap = new Map<string, Commit>();
    commits.forEach(c => commitMap.set(c.hash, c));

    // Draw parent links (DAG edges)
    commits.forEach(c => {
      if (!c.parentHashes) return;
      const parents = c.parentHashes.split(/[,\s]+/).map(p => p.trim());
      parents.forEach(pHash => {
        const parent = commitMap.get(pHash);
        if (parent) {
          const x1 = xScale(new Date(parent.authorDate));
          const y1 = yScale(parent.authorEmail) || 0;
          const x2 = xScale(new Date(c.authorDate));
          const y2 = yScale(c.authorEmail) || 0;

          const curvature = 0.5;
          const xi = d3.interpolateNumber(x1, x2);
          const x2Mid = xi(curvature);
          const x3Mid = xi(1 - curvature);

          g.append('path')
            .attr('d', `M ${x1} ${y1} C ${x2Mid} ${y1}, ${x3Mid} ${y2}, ${x2} ${y2}`)
            .attr('fill', 'none')
            .attr('stroke', '#334155')
            .attr('stroke-width', 1.5)
            .attr('opacity', 0.6);
        }
      });
    });

    // Commit Nodes
    const nodeGroup = g.selectAll('.commit-node')
      .data(commits)
      .enter()
      .append('g')
      .attr('class', 'commit-node')
      .attr('transform', d => `translate(${xScale(new Date(d.authorDate))}, ${yScale(d.authorEmail)})`)
      .style('cursor', 'pointer')
      .on('click', (_, d) => {
        activeCommit = d;
      });

    // Outer glow for active commit
    nodeGroup.append('circle')
      .attr('r', d => (activeCommit?.hash === d.hash ? 10 : 7))
      .attr('fill', d => authorColorMap.get(d.authorEmail) || '#06b6d4')
      .attr('fill-opacity', d => (activeCommit?.hash === d.hash ? 0.3 : 0.15))
      .attr('stroke', d => (activeCommit?.hash === d.hash ? '#38bdf8' : 'none'))
      .attr('stroke-width', 2);

    // Inner circle
    nodeGroup.append('circle')
      .attr('r', d => (activeCommit?.hash === d.hash ? 6 : 4))
      .attr('fill', d => authorColorMap.get(d.authorEmail) || '#06b6d4')
      .attr('stroke', '#0f172a')
      .attr('stroke-width', 1.5);

    // Commit short hash label
    nodeGroup.append('text')
      .attr('y', -10)
      .attr('text-anchor', 'middle')
      .attr('font-size', '9px')
      .attr('font-family', 'monospace')
      .attr('fill', '#94a3b8')
      .text(d => d.shortHash);
  }

  function inspectInExplorer() {
    if (activeCommit) {
      selectedCommit.set(activeCommit);
      currentView.set('commits');
    }
  }
</script>

<div class="flex h-full overflow-hidden" bind:this={container}>
  <!-- Timeline Canvas Area -->
  <div class="flex-1 flex flex-col h-full overflow-hidden p-5">
    <div class="flex items-center justify-between pb-3 border-b border-slate-800/80 mb-3">
      <div>
        <h2 class="text-base font-semibold text-slate-100">Interactive Commit Timeline</h2>
        <p class="text-xs text-slate-400">Zoom and pan to trace branch topology, merge connections, and commit cadences.</p>
      </div>
      <div class="flex items-center gap-2 text-xs text-slate-400">
        <span class="inline-flex items-center gap-1.5 px-2 py-1 rounded bg-slate-900 border border-slate-800">
          <svg class="w-3.5 h-3.5 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122" />
          </svg>
          Scroll to Zoom • Drag to Pan
        </span>
      </div>
    </div>

    {#if loading}
      <div class="flex-1 flex items-center justify-center">
        <SkeletonLoader count={3} height="h-28" width="w-3/4" />
      </div>
    {:else}
      <div class="flex-1 w-full bg-slate-950/60 rounded-xl border border-slate-800/80 relative overflow-hidden shadow-inner">
        <svg bind:this={svgElement} class="w-full h-full"></svg>
      </div>
    {/if}
  </div>

  <!-- Commit Inspector Drawer -->
  {#if activeCommit}
    <div class="w-80 md:w-96 border-l border-slate-800/80 bg-slate-900/90 backdrop-blur-md p-5 flex flex-col justify-between overflow-y-auto">
      <div class="space-y-4">
        <div class="flex items-center justify-between pb-3 border-b border-slate-800">
          <span class="text-xs font-semibold uppercase tracking-wider text-slate-400">Commit Details</span>
          <span class="px-2 py-0.5 rounded text-xs font-mono font-bold bg-cyan-950/80 text-cyan-400 border border-cyan-800/60">
            {activeCommit.shortHash}
          </span>
        </div>

        <div>
          <h3 class="text-sm font-semibold text-slate-100 leading-snug">{activeCommit.subject}</h3>
          {#if activeCommit.body}
            <p class="mt-2 text-xs text-slate-400 whitespace-pre-wrap bg-slate-950/60 p-2.5 rounded border border-slate-800 font-mono">
              {activeCommit.body}
            </p>
          {/if}
        </div>

        <div class="grid grid-cols-2 gap-3 text-xs bg-slate-950/40 p-3 rounded-lg border border-slate-800/60">
          <div>
            <span class="text-slate-500 block text-[10px] uppercase">Author</span>
            <span class="text-slate-200 font-medium truncate block">{activeCommit.authorName}</span>
            <span class="text-slate-500 text-[10px] truncate block">{activeCommit.authorEmail}</span>
          </div>
          <div>
            <span class="text-slate-500 block text-[10px] uppercase">Timestamp</span>
            <span class="text-slate-200 font-mono text-[11px] block">
              {new Date(activeCommit.authorDate).toLocaleString()}
            </span>
          </div>
        </div>

        <div class="grid grid-cols-3 gap-2 text-center text-xs">
          <div class="p-2 rounded bg-slate-800/50 border border-slate-800">
            <span class="block text-slate-500 text-[10px]">Files</span>
            <strong class="text-slate-200 font-mono">{activeCommit.filesChangedCount}</strong>
          </div>
          <div class="p-2 rounded bg-emerald-950/40 border border-emerald-900/40">
            <span class="block text-emerald-500 text-[10px]">Insertions</span>
            <strong class="text-emerald-400 font-mono">+{activeCommit.insertions}</strong>
          </div>
          <div class="p-2 rounded bg-rose-950/40 border border-rose-900/40">
            <span class="block text-rose-500 text-[10px]">Deletions</span>
            <strong class="text-rose-400 font-mono">-{activeCommit.deletions}</strong>
          </div>
        </div>

        {#if activeCommit.parentHashes}
          <div>
            <span class="text-[10px] font-semibold uppercase tracking-wider text-slate-500 block mb-1">
              Parent Commits
            </span>
            <div class="flex flex-wrap gap-1.5">
              {#each activeCommit.parentHashes.split(/[,\s]+/) as p}
                <span class="px-2 py-0.5 rounded text-[10px] font-mono bg-slate-800 text-slate-300 border border-slate-700">
                  {p.slice(0, 7)}
                </span>
              {/each}
            </div>
          </div>
        {/if}
      </div>

      <div class="pt-4 mt-6 border-t border-slate-800">
        <button
          class="w-full flex items-center justify-center gap-2 py-2 px-4 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-medium text-xs transition shadow-lg shadow-cyan-950/40"
          on:click={inspectInExplorer}
        >
          <span>Inspect Diff in Explorer</span>
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3"/>
          </svg>
        </button>
      </div>
    </div>
  {/if}
</div>
