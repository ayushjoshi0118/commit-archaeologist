<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import {
    currentView,
    type ViewType,
    repositories,
    selectedRepo,
    loadRepoStats,
    launchDemoMode
  } from '../../stores/app';

  export let isCollapsed = false;
  const dispatch = createEventDispatcher();

  const navItems: { id: ViewType; label: string; icon: string; countKey?: string }[] = [
    {
      id: 'overview',
      label: 'Overview',
      icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6'
    },
    {
      id: 'timeline',
      label: 'Commit Timeline',
      icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z'
    },
    {
      id: 'hotspots',
      label: 'Code Hotspots',
      icon: 'M17.657 18.657A8 8 0 016.343 7.343S7 9 9 10c0-2 .5-5 2.986-7C14 5 16.09 5.777 17.656 7.343a7.975 7.975 0 012.343 5.657c0 2.12-.837 4.148-2.342 5.657z'
    },
    {
      id: 'dependencies',
      label: 'Dependency Graph',
      icon: 'M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1'
    },
    {
      id: 'commits',
      label: 'Commit Explorer',
      icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2'
    },
    {
      id: 'contributors',
      label: 'Contributor Analytics',
      icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z'
    },
    {
      id: 'settings',
      label: 'Settings',
      icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z'
    }
  ];

  function handleSelectRepo(e: Event) {
    const target = e.target as HTMLSelectElement;
    if (target.value) {
      loadRepoStats(target.value);
    }
  }
</script>

<aside
  class="flex flex-col justify-between h-full border-r border-slate-800/80 bg-slate-950/90 backdrop-blur-md transition-all duration-200 z-30 select-none {isCollapsed ? 'w-16' : 'w-64'}"
>
  <div>
    <!-- Brand Header -->
    <div class="flex items-center justify-between h-14 px-4 border-b border-slate-800/80">
      <div class="flex items-center gap-3 overflow-hidden">
        <div class="flex items-center justify-center w-8 h-8 rounded-lg bg-cyan-950 border border-cyan-800/60 text-cyan-400 shrink-0 shadow-inner">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"/>
          </svg>
        </div>
        {#if !isCollapsed}
          <div class="flex flex-col">
            <span class="text-xs font-bold uppercase tracking-wider text-slate-100">Archaeologist</span>
            <span class="text-[10px] text-slate-500 font-mono">v1.0.0 • Git Observability</span>
          </div>
        {/if}
      </div>

      <button
        class="p-1 rounded text-slate-400 hover:text-slate-200 hover:bg-slate-900 transition"
        on:click={() => (isCollapsed = !isCollapsed)}
        title={isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          {#if isCollapsed}
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 5l7 7-7 7M5 5l7 7-7 7"/>
          {:else}
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7"/>
          {/if}
        </svg>
      </button>
    </div>

    <!-- Repository Switcher -->
    {#if !isCollapsed}
      <div class="p-3 border-b border-slate-800/60">
        <label for="repo-select" class="block text-[10px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">
          Active Repository
        </label>
        <select
          id="repo-select"
          class="w-full px-2.5 py-1.5 text-xs bg-slate-900 border border-slate-700/80 rounded-md text-slate-200 focus:outline-none focus:border-cyan-500 font-sans cursor-pointer truncate"
          on:change={handleSelectRepo}
          value={$selectedRepo?.id || ''}
        >
          {#if $repositories.length === 0}
            <option value="" disabled>No repositories analyzed</option>
          {:else}
            {#each $repositories as r}
              <option value={r.id}>
                {r.name} {r.isDemo ? '(Demo)' : ''}
              </option>
            {/each}
          {/if}
        </select>
      </div>
    {/if}

    <!-- Navigation Items -->
    <nav class="p-2 space-y-1">
      {#each navItems as item}
        {@const active = $currentView === item.id}
        <button
          class="flex items-center gap-3 w-full px-3 py-2 rounded-lg text-xs font-medium transition text-left relative {active
            ? 'bg-slate-900 text-cyan-400 border border-slate-800 shadow-sm'
            : 'text-slate-400 hover:text-slate-200 hover:bg-slate-900/60'}"
          on:click={() => currentView.set(item.id)}
          title={isCollapsed ? item.label : undefined}
        >
          {#if active}
            <span class="absolute left-0 top-2 bottom-2 w-1 rounded-r bg-cyan-400"></span>
          {/if}
          <svg class="w-4 h-4 shrink-0 {active ? 'text-cyan-400' : 'text-slate-400'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d={item.icon}/>
          </svg>
          {#if !isCollapsed}
            <span class="truncate">{item.label}</span>
          {/if}
        </button>
      {/each}
    </nav>
  </div>

  <!-- Bottom Quick Actions -->
  <div class="p-3 border-t border-slate-800/80 space-y-2">
    {#if !isCollapsed}
      <button
        class="w-full flex items-center justify-center gap-2 px-3 py-2 rounded-lg text-xs font-medium bg-emerald-950/40 hover:bg-emerald-900/40 text-emerald-400 border border-emerald-800/50 transition shadow-inner"
        on:click={launchDemoMode}
        title="Load bundled synthetic sample repo"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
        <span>Demo (ChronoLog)</span>
      </button>

      <button
        class="w-full flex items-center justify-center gap-2 px-3 py-2 rounded-lg text-xs font-medium bg-cyan-950/40 hover:bg-cyan-900/40 text-cyan-400 border border-cyan-800/50 transition"
        on:click={() => dispatch('openModal')}
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
        </svg>
        <span>Add Repository</span>
      </button>
    {:else}
      <button
        class="w-full flex items-center justify-center p-2 rounded-lg text-emerald-400 hover:bg-emerald-950/50 transition"
        on:click={launchDemoMode}
        title="Launch Demo Mode"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
        </svg>
      </button>
    {/if}
  </div>
</aside>
