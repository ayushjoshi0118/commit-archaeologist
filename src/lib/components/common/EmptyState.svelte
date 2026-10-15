<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  export let title = 'No Repository Selected';
  export let description = 'Open a local Git repository to explore its commit timeline, code hotspots, and AST dependencies.';

  const dispatch = createEventDispatcher();
</script>

<div class="flex flex-col items-center justify-center h-full min-h-[450px] p-8 text-center">
  <div class="relative flex items-center justify-center w-20 h-20 mb-6 rounded-2xl bg-cyan-950/40 border border-cyan-800/50 shadow-inner">
    <svg class="w-10 h-10 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
    </svg>
    <div class="absolute -top-1 -right-1 w-3 h-3 rounded-full bg-cyan-400 animate-pulse"></div>
  </div>

  <h2 class="text-xl font-semibold text-slate-100 mb-2">{title}</h2>
  <p class="max-w-md text-sm text-slate-400 mb-8 leading-relaxed">
    {description}
  </p>

  <div class="flex flex-wrap items-center justify-center gap-4">
    <button
      class="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-medium text-sm transition shadow-lg shadow-cyan-950/50 hover:shadow-cyan-900/40"
      on:click={() => dispatch('openRepo')}
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
      </svg>
      Select Local Repository
    </button>

    <button
      class="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg bg-slate-800 hover:bg-slate-700/80 text-slate-200 border border-slate-700/80 font-medium text-sm transition"
      on:click={() => dispatch('launchDemo')}
    >
      <svg class="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
      </svg>
      Launch Demo Mode (ChronoLog)
    </button>
  </div>
</div>
