<script lang="ts">
  import { onMount } from 'svelte';
  import type { Commit, CommitDiffResponse, FileDiff } from '../../types';
  import { selectedRepo, selectedCommit, fileFilter, searchQuery } from '../../stores/app';
  import { api } from '../../api/client';
  import SkeletonLoader from '../common/SkeletonLoader.svelte';

  let commits: Commit[] = [];
  let loadingCommits = true;
  let loadingDiff = false;
  let diffResponse: CommitDiffResponse | null = null;
  let selectedFileDiff: FileDiff | null = null;

  let localSearch = '';
  let localAuthor = '';
  let activeCommit: Commit | null = null;

  onMount(async () => {
    localSearch = $searchQuery || '';
    await loadCommits();
  });

  $: if ($selectedRepo?.id) {
    loadCommits();
  }

  $: if ($selectedCommit) {
    activeCommit = $selectedCommit;
    loadDiff(activeCommit.hash);
  }

  async function loadCommits() {
    if (!$selectedRepo) return;
    loadingCommits = true;
    try {
      commits = await api.getCommits({
        repoId: $selectedRepo.id,
        search: localSearch,
        author: localAuthor,
        file: $fileFilter,
        limit: 100
      });

      if (commits.length > 0 && !activeCommit) {
        selectCommit(commits[0]);
      }
    } catch (err) {
      console.error('Failed to load commits:', err);
    } finally {
      loadingCommits = false;
    }
  }

  async function selectCommit(c: Commit) {
    activeCommit = c;
    selectedCommit.set(c);
    await loadDiff(c.hash);
  }

  async function loadDiff(hash: string) {
    if (!$selectedRepo) return;
    loadingDiff = true;
    try {
      diffResponse = await api.getDiff($selectedRepo.id, hash, $fileFilter);
      if (diffResponse.files.length > 0) {
        selectedFileDiff = diffResponse.files[0];
      } else {
        selectedFileDiff = null;
      }
    } catch (err) {
      console.error('Failed to load diff:', err);
      diffResponse = null;
      selectedFileDiff = null;
    } finally {
      loadingDiff = false;
    }
  }

  function clearFileFilter() {
    fileFilter.set('');
    loadCommits();
  }
</script>

<div class="flex h-full overflow-hidden">
  <!-- Left: Commit Filter & List -->
  <div class="w-80 md:w-96 border-r border-slate-800/80 bg-slate-950/60 flex flex-col h-full overflow-hidden shrink-0">
    <!-- Filter bar -->
    <div class="p-4 border-b border-slate-800/80 space-y-3">
      <div class="flex items-center justify-between">
        <h3 class="text-xs font-semibold uppercase tracking-wider text-slate-300">Commits ({commits.length})</h3>
        {#if $fileFilter}
          <button
            class="text-[10px] text-cyan-400 hover:text-cyan-300 transition underline"
            on:click={clearFileFilter}
          >
            Clear file filter
          </button>
        {/if}
      </div>

      {#if $fileFilter}
        <div class="flex items-center gap-1.5 p-1.5 rounded bg-cyan-950/40 border border-cyan-800/50 text-[11px] text-cyan-300 font-mono truncate">
          <span class="text-cyan-500">File:</span>
          <span class="truncate">{$fileFilter}</span>
        </div>
      {/if}

      <div class="relative">
        <input
          type="text"
          placeholder="Filter by subject or hash..."
          bind:value={localSearch}
          on:input={() => loadCommits()}
          class="w-full px-3 py-1.5 rounded-lg text-xs bg-slate-900 border border-slate-700/80 text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500"
        />
      </div>
    </div>

    <!-- Commits Scroll List -->
    <div class="flex-1 overflow-y-auto divide-y divide-slate-800/60">
      {#if loadingCommits}
        <div class="p-4 space-y-3">
          <SkeletonLoader count={6} height="h-16" />
        </div>
      {:else if commits.length === 0}
        <div class="p-8 text-center text-xs text-slate-500">
          No commits matched the filters.
        </div>
      {:else}
        {#each commits as c}
          {@const isActive = activeCommit?.hash === c.hash}
          <button
            class="w-full text-left p-3.5 transition flex flex-col gap-1.5 {isActive
              ? 'bg-slate-900 border-l-2 border-l-cyan-400'
              : 'hover:bg-slate-900/40 text-slate-400'}"
            on:click={() => selectCommit(c)}
          >
            <div class="flex items-center justify-between gap-2">
              <span class="text-xs font-medium text-slate-100 truncate flex-1">
                {c.subject}
              </span>
              <span class="text-[10px] font-mono text-cyan-400 shrink-0 bg-slate-800/80 px-1.5 py-0.5 rounded">
                {c.shortHash}
              </span>
            </div>

            <div class="flex items-center justify-between text-[11px] text-slate-500">
              <span class="truncate">{c.authorName}</span>
              <span class="shrink-0">{new Date(c.authorDate).toLocaleDateString()}</span>
            </div>

            <div class="flex items-center gap-2 text-[10px] font-mono">
              <span class="text-emerald-400">+{c.insertions}</span>
              <span class="text-rose-400">-{c.deletions}</span>
              <span class="text-slate-600">•</span>
              <span class="text-slate-400">{c.filesChangedCount} files</span>
            </div>
          </button>
        {/each}
      {/if}
    </div>
  </div>

  <!-- Right: Commit Details & Unified Diff Viewer -->
  <div class="flex-1 flex flex-col h-full overflow-hidden bg-slate-900/30">
    {#if !activeCommit}
      <div class="flex-1 flex items-center justify-center text-slate-500 text-sm">
        Select a commit to view its metadata and code diffs.
      </div>
    {:else}
      <!-- Commit Header Summary -->
      <div class="p-5 border-b border-slate-800/80 bg-slate-950/40 shrink-0">
        <div class="flex items-start justify-between gap-4">
          <div class="space-y-1.5 flex-1 min-w-0">
            <h2 class="text-base font-semibold text-slate-100 leading-snug">
              {activeCommit.subject}
            </h2>
            {#if activeCommit.body}
              <p class="text-xs text-slate-400 whitespace-pre-wrap font-mono bg-slate-950/60 p-2 rounded border border-slate-800/60">
                {activeCommit.body}
              </p>
            {/if}
            <div class="flex flex-wrap items-center gap-3 text-xs text-slate-400 pt-1">
              <span>Author: <strong class="text-slate-200">{activeCommit.authorName}</strong> ({activeCommit.authorEmail})</span>
              <span>•</span>
              <span>Committed: <strong class="text-slate-200">{new Date(activeCommit.authorDate).toLocaleString()}</strong></span>
              <span>•</span>
              <span class="font-mono text-cyan-400">Hash: {activeCommit.hash}</span>
            </div>
          </div>
        </div>

        <!-- Files Changed Pill Tabs -->
        {#if diffResponse?.files && diffResponse.files.length > 0}
          <div class="mt-4 flex items-center gap-2 overflow-x-auto pb-1">
            <span class="text-[10px] font-semibold uppercase text-slate-500 shrink-0">Files Changed:</span>
            {#each diffResponse.files as f}
              {@const isSelected = selectedFileDiff?.filePath === f.filePath}
              <button
                class="px-2.5 py-1 rounded text-xs font-mono font-medium shrink-0 transition flex items-center gap-1.5 {isSelected
                  ? 'bg-cyan-500 text-slate-950 font-bold'
                  : 'bg-slate-800 text-slate-300 hover:bg-slate-700/80 border border-slate-700/60'}"
                on:click={() => (selectedFileDiff = f)}
              >
                <span class="text-[9px] uppercase px-1 rounded {f.changeType === 'added' ? 'bg-emerald-950 text-emerald-400' : f.changeType === 'deleted' ? 'bg-rose-950 text-rose-400' : 'bg-slate-700 text-slate-300'}">
                  {f.changeType[0]}
                </span>
                <span>{f.filePath.split('/').pop()}</span>
              </button>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Diff Viewer Area -->
      <div class="flex-1 overflow-y-auto p-5">
        {#if loadingDiff}
          <SkeletonLoader count={8} height="h-6" />
        {:else if !selectedFileDiff || selectedFileDiff.hunks.length === 0}
          <div class="h-64 flex items-center justify-center text-slate-500 text-xs">
            No text diff available for this file (binary file or empty change).
          </div>
        {:else}
          <div class="rounded-xl border border-slate-800 overflow-hidden bg-slate-950/80 shadow-inner font-mono text-xs">
            <!-- Diff file header -->
            <div class="px-4 py-2.5 bg-slate-900 border-b border-slate-800 flex items-center justify-between">
              <span class="text-slate-300 font-semibold">{selectedFileDiff.filePath}</span>
              <span class="text-[10px] uppercase font-bold text-slate-500">{selectedFileDiff.changeType}</span>
            </div>

            <!-- Hunks -->
            <div class="divide-y divide-slate-800/40">
              {#each selectedFileDiff.hunks as hunk}
                <div class="diff-hunk">
                  <div class="px-4 py-1 bg-cyan-950/20 text-cyan-400/80 text-[11px] border-y border-cyan-900/30">
                    {hunk.header}
                  </div>
                  <div class="leading-relaxed">
                    {#each hunk.lines as line}
                      <div
                        class="px-4 py-0.5 flex items-start gap-3 {line.type === 'add'
                          ? 'bg-emerald-950/30 text-emerald-300'
                          : line.type === 'del'
                          ? 'bg-rose-950/30 text-rose-300'
                          : 'text-slate-400 hover:bg-slate-900/40'}"
                      >
                        <span class="w-8 shrink-0 text-right select-none text-[10px] text-slate-600">
                          {line.type === 'add' ? '+' : line.type === 'del' ? '-' : (line.newLine || '')}
                        </span>
                        <pre class="flex-1 font-mono text-xs whitespace-pre-wrap break-all">{line.content}</pre>
                      </div>
                    {/each}
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>
