<script lang="ts">
  import { onMount } from 'svelte';
  import type { AppSettings } from '../../types';
  import { api } from '../../api/client';
  import { selectedRepo } from '../../stores/app';

  let settings: AppSettings | null = null;
  let loading = true;

  onMount(async () => {
    try {
      settings = await api.getSettings();
    } catch (err) {
      console.error('Failed to load settings:', err);
    } finally {
      loading = false;
    }
  });
</script>

<div class="h-full overflow-y-auto p-6 max-w-4xl space-y-6">
  <div class="pb-3 border-b border-slate-800/80">
    <h2 class="text-base font-semibold text-slate-100">Repository & Application Settings</h2>
    <p class="text-xs text-slate-400">
      Security postures, local persistence configurations, and language parser diagnostics.
    </p>
  </div>

  <!-- Privacy & Local-First Guarantee -->
  <div class="p-5 rounded-xl bg-slate-900/60 border border-slate-800/80 space-y-4">
    <div class="flex items-center gap-3">
      <div class="p-2 rounded-lg bg-emerald-950/60 text-emerald-400 border border-emerald-800/50">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
        </svg>
      </div>
      <div>
        <h3 class="text-sm font-semibold text-slate-100">Local-Only Security Architecture</h3>
        <p class="text-xs text-slate-400">Your codebase and git history never leave this workstation.</p>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
      <div class="p-3 rounded-lg bg-slate-950/60 border border-slate-800">
        <span class="text-slate-500 font-semibold block uppercase text-[10px]">Cloud Telemetry</span>
        <span class="text-emerald-400 font-medium">Disabled (Air-Gapped Local Execution)</span>
      </div>
      <div class="p-3 rounded-lg bg-slate-950/60 border border-slate-800">
        <span class="text-slate-500 font-semibold block uppercase text-[10px]">AI Integration</span>
        <span class="text-amber-400 font-medium">Disabled by Default (Zero External API Calls)</span>
      </div>
    </div>
  </div>

  <!-- Local SQLite Persistence -->
  <div class="p-5 rounded-xl bg-slate-900/60 border border-slate-800/80 space-y-4">
    <h3 class="text-sm font-semibold text-slate-100">Local SQLite Storage</h3>
    <p class="text-xs text-slate-400">
      All commit indexes, hotspot churn scores, and AST dependency links are cached locally in pure-Go SQLite for instant subsequent loads.
    </p>

    <div class="space-y-2 text-xs font-mono">
      <div class="p-3 rounded-lg bg-slate-950/60 border border-slate-800 flex items-center justify-between">
        <span class="text-slate-400">Database Driver:</span>
        <span class="text-cyan-400 font-semibold">modernc.org/sqlite (Pure-Go CGo-Free)</span>
      </div>
      <div class="p-3 rounded-lg bg-slate-950/60 border border-slate-800 flex items-center justify-between">
        <span class="text-slate-400">Active Repo Directory:</span>
        <span class="text-slate-200 truncate max-w-md">{$selectedRepo?.path || 'None loaded'}</span>
      </div>
    </div>
  </div>

  <!-- Language Parsers Diagnostics -->
  <div class="p-5 rounded-xl bg-slate-900/60 border border-slate-800/80 space-y-4">
    <h3 class="text-sm font-semibold text-slate-100">AST Parser Status</h3>
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
      <div class="p-3 rounded-lg bg-slate-950/60 border border-slate-800 text-center">
        <span class="text-cyan-400 font-semibold block">Go</span>
        <span class="text-[10px] text-slate-500">go/parser & AST</span>
      </div>
      <div class="p-3 rounded-lg bg-slate-950/60 border border-slate-800 text-center">
        <span class="text-indigo-400 font-semibold block">TypeScript</span>
        <span class="text-[10px] text-slate-500">ES6 + Type Imports</span>
      </div>
      <div class="p-3 rounded-lg bg-slate-950/60 border border-slate-800 text-center">
        <span class="text-amber-400 font-semibold block">JavaScript</span>
        <span class="text-[10px] text-slate-500">CJS & ESM Scanner</span>
      </div>
      <div class="p-3 rounded-lg bg-slate-950/60 border border-slate-800 text-center">
        <span class="text-blue-400 font-semibold block">Python</span>
        <span class="text-[10px] text-slate-500">AST Module Resolver</span>
      </div>
    </div>
  </div>
</div>
