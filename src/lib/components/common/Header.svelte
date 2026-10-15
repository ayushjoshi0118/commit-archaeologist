<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { selectedRepo, searchQuery, overviewStats } from '../../stores/app';

  const dispatch = createEventDispatcher();
</script>

<header class="flex items-center justify-between h-14 px-5 border-b border-slate-800/80 bg-slate-900/70 backdrop-blur-md z-20">
  <div class="flex items-center gap-4 min-w-0">
    {#if $selectedRepo}
      <div class="flex items-center gap-2.5 min-w-0">
        <span class="font-semibold text-slate-100 text-sm tracking-tight truncate max-w-xs">
          {$selectedRepo.name}
        </span>
        <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-mono font-medium bg-slate-800 text-slate-300 border border-slate-700/60">
          <svg class="w-3 h-3 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
          </svg>
          {$selectedRepo.defaultBranch || 'main'}
        </span>
        {#if $selectedRepo.isDemo}
          <span class="px-2 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wider bg-emerald-950/60 text-emerald-400 border border-emerald-800/60">
            Demo Mode
          </span>
        {/if}
      </div>

      <div class="hidden lg:flex items-center gap-4 pl-4 border-l border-slate-800 text-xs text-slate-400">
        <span title="Total commits analyzed">
          <strong class="text-slate-200 font-mono">{$selectedRepo.totalCommits}</strong> commits
        </span>
        <span>•</span>
        <span title="Active contributors">
          <strong class="text-slate-200 font-mono">{$selectedRepo.totalAuthors}</strong> authors
        </span>
        <span>•</span>
        <span title="Tracked repository files">
          <strong class="text-slate-200 font-mono">{$selectedRepo.totalFiles}</strong> files
        </span>
      </div>
    {:else}
      <div class="flex items-center gap-2 text-slate-400 text-sm">
        <span class="w-2 h-2 rounded-full bg-slate-500"></span>
        <span>No repository loaded</span>
      </div>
    {/if}
  </div>

  <div class="flex items-center gap-3">
    <!-- Global search -->
    <div class="relative w-64 md:w-80">
      <svg class="absolute left-3 top-2.5 w-4 h-4 text-slate-500 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      <input
        type="text"
        placeholder="Search commits, files, authors..."
        bind:value={$searchQuery}
        class="w-full pl-9 pr-3 py-1.5 rounded-lg text-xs bg-slate-950/60 border border-slate-800 text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500/80 focus:ring-1 focus:ring-cyan-500/50 transition font-sans"
      />
    </div>

    <!-- Open Repo Button -->
    <button
      class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-slate-800 hover:bg-slate-700/80 text-slate-200 border border-slate-700 transition"
      on:click={() => dispatch('openModal')}
      title="Analyze a local Git directory"
    >
      <svg class="w-3.5 h-3.5 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
      </svg>
      <span>Open Repo</span>
    </button>
  </div>
</header>
