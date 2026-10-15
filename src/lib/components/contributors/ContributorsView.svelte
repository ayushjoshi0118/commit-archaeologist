<script lang="ts">
  import { onMount } from 'svelte';
  import * as d3 from 'd3';
  import type { Contributor, ContributorActivity } from '../../types';
  import { selectedRepo } from '../../stores/app';
  import { api } from '../../api/client';
  import SkeletonLoader from '../common/SkeletonLoader.svelte';

  let contributors: Contributor[] = [];
  let activities: ContributorActivity[] = [];
  let loading = true;
  let selectedAuthorEmail = '';
  let punchcardSvg: SVGSVGElement;

  const dayLabels = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

  onMount(async () => {
    await loadData();
  });

  $: if ($selectedRepo?.id) {
    loadData();
  }

  async function loadData() {
    if (!$selectedRepo) return;
    loading = true;
    try {
      contributors = await api.getContributors($selectedRepo.id);
      activities = await api.getActivity($selectedRepo.id, selectedAuthorEmail);
      renderPunchcard();
    } catch (err) {
      console.error('Failed to load contributor data:', err);
    } finally {
      loading = false;
    }
  }

  async function selectAuthor(email: string) {
    selectedAuthorEmail = selectedAuthorEmail === email ? '' : email;
    if ($selectedRepo) {
      activities = await api.getActivity($selectedRepo.id, selectedAuthorEmail);
      renderPunchcard();
    }
  }

  function renderPunchcard() {
    if (!punchcardSvg || activities.length === 0) return;

    const width = punchcardSvg.clientWidth || 800;
    const height = 280;
    const margin = { top: 30, right: 30, bottom: 40, left: 50 };

    const svg = d3.select(punchcardSvg);
    svg.selectAll('*').remove();
    svg.attr('viewBox', `0 0 ${width} ${height}`);

    const g = svg.append('g').attr('transform', `translate(${margin.left}, ${margin.top})`);

    const innerWidth = width - margin.left - margin.right;
    const innerHeight = height - margin.top - margin.bottom;

    const cellWidth = innerWidth / 24;
    const cellHeight = innerHeight / 7;

    const maxCount = d3.max(activities, d => d.commitCount) || 1;
    const rScale = d3.scaleSqrt().domain([0, maxCount]).range([2, Math.min(cellWidth, cellHeight) / 2 - 2]);
    const colorScale = d3.scaleSequential(d3.interpolateCool).domain([0, maxCount]);

    // Hours Axis (top or bottom)
    for (let h = 0; h < 24; h++) {
      g.append('text')
        .attr('x', h * cellWidth + cellWidth / 2)
        .attr('y', innerHeight + 20)
        .attr('text-anchor', 'middle')
        .attr('font-size', '10px')
        .attr('font-family', 'monospace')
        .attr('fill', '#64748b')
        .text(h % 3 === 0 ? `${h}:00` : '');
    }

    // Days Axis (left)
    for (let d = 0; d < 7; d++) {
      g.append('text')
        .attr('x', -12)
        .attr('y', d * cellHeight + cellHeight / 2 + 4)
        .attr('text-anchor', 'end')
        .attr('font-size', '10px')
        .attr('font-family', 'sans-serif')
        .attr('fill', '#94a3b8')
        .text(dayLabels[d]);
    }

    // Grid dots
    for (let d = 0; d < 7; d++) {
      for (let h = 0; h < 24; h++) {
        const item = activities.find(a => a.dayOfWeek === d && a.hourOfDay === h);
        const count = item ? item.commitCount : 0;
        const cx = h * cellWidth + cellWidth / 2;
        const cy = d * cellHeight + cellHeight / 2;

        // Base slot marker
        g.append('circle')
          .attr('cx', cx)
          .attr('cy', cy)
          .attr('r', 2)
          .attr('fill', '#1e293b');

        if (count > 0) {
          g.append('circle')
            .attr('cx', cx)
            .attr('cy', cy)
            .attr('r', rScale(count))
            .attr('fill', '#06b6d4')
            .attr('fill-opacity', 0.8)
            .attr('stroke', '#38bdf8')
            .attr('stroke-width', 1.5)
            .append('title')
            .text(`${dayLabels[d]} at ${h}:00 - ${count} commit${count > 1 ? 's' : ''}`);
        }
      }
    }
  }
</script>

<div class="h-full overflow-y-auto p-6 space-y-6">
  <div class="pb-3 border-b border-slate-800/80">
    <h2 class="text-base font-semibold text-slate-100">Contributor Analytics & Activity Punchcard</h2>
    <p class="text-xs text-slate-400">
      Temporal commit distribution across days of the week and hours of the day.
    </p>
  </div>

  {#if loading}
    <SkeletonLoader count={4} height="h-28" />
  {:else}
    <!-- 7x24 Punchcard Section -->
    <div class="p-5 rounded-xl bg-slate-900/60 border border-slate-800/80 shadow-sm">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h3 class="text-sm font-semibold text-slate-100">
            Commit Punchcard {selectedAuthorEmail ? `(${selectedAuthorEmail})` : '(All Contributors)'}
          </h3>
          <p class="text-xs text-slate-500">Bubble radius indicates commit frequency at that specific hour.</p>
        </div>
        {#if selectedAuthorEmail}
          <button
            class="text-xs text-cyan-400 hover:text-cyan-300 transition"
            on:click={() => selectAuthor('')}
          >
            Show All Contributors
          </button>
        {/if}
      </div>
      <div class="w-full h-72">
        <svg bind:this={punchcardSvg} class="w-full h-full"></svg>
      </div>
    </div>

    <!-- Contributor Leaderboard Grid -->
    <div>
      <h3 class="text-xs font-semibold uppercase tracking-wider text-slate-400 mb-3">
        Active Contributors ({contributors.length})
      </h3>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {#each contributors as c}
          {@const isSelected = selectedAuthorEmail === c.email}
          <div
            role="button"
            tabindex="0"
            class="p-4 rounded-xl border transition cursor-pointer flex flex-col justify-between {isSelected
              ? 'bg-cyan-950/30 border-cyan-500/80 shadow-md shadow-cyan-950/40'
              : 'bg-slate-900/60 border-slate-800/80 hover:border-slate-700'}"
            on:click={() => selectAuthor(c.email)}
            on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ') selectAuthor(c.email); }}
          >
            <div>
              <div class="flex items-center justify-between mb-2">
                <span class="font-semibold text-sm text-slate-100 truncate">{c.name}</span>
                <span class="px-2 py-0.5 rounded text-xs font-mono font-bold bg-cyan-950 text-cyan-400 border border-cyan-800/60">
                  {c.totalCommits} commits
                </span>
              </div>
              <span class="text-xs text-slate-500 font-mono block truncate mb-3">{c.email}</span>
            </div>

            <div class="pt-3 border-t border-slate-800/60 flex items-center justify-between text-xs font-mono">
              <span class="text-emerald-400">+{c.insertions.toLocaleString()}</span>
              <span class="text-rose-400">-{c.deletions.toLocaleString()}</span>
              <span class="text-slate-500 text-[10px]">
                {new Date(c.lastCommitDate).toLocaleDateString()}
              </span>
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>
