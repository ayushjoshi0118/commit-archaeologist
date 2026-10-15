<script lang="ts">
  import { onMount } from 'svelte';
  import './app.css';

  import Sidebar from './lib/components/common/Sidebar.svelte';
  import Header from './lib/components/common/Header.svelte';
  import EmptyState from './lib/components/common/EmptyState.svelte';
  import Modal from './lib/components/common/Modal.svelte';

  import OverviewView from './lib/components/dashboard/OverviewView.svelte';
  import TimelineView from './lib/components/timeline/TimelineView.svelte';
  import HotspotsView from './lib/components/hotspots/HotspotsView.svelte';
  import DependencyView from './lib/components/dependencies/DependencyView.svelte';
  import CommitExplorerView from './lib/components/explorer/CommitExplorerView.svelte';
  import ContributorsView from './lib/components/contributors/ContributorsView.svelte';
  import SettingsView from './lib/components/settings/SettingsView.svelte';

  import {
    initApp,
    currentView,
    selectedRepo,
    isLoading,
    loadingMessage,
    errorMessage,
    analyzeNewRepository,
    launchDemoMode
  } from './lib/stores/app';

  let isModalOpen = false;
  let repoPathInput = '';
  let isAnalyzing = false;
  let modalError = '';

  onMount(async () => {
    await initApp();
  });

  async function handleOpenRepoSubmit() {
    if (!repoPathInput.trim()) {
      modalError = 'Please enter a valid directory path.';
      return;
    }

    modalError = '';
    isAnalyzing = true;
    try {
      await analyzeNewRepository(repoPathInput.trim());
      isModalOpen = false;
      repoPathInput = '';
    } catch (err: any) {
      modalError = err.message || 'Failed to analyze repository.';
    } finally {
      isAnalyzing = false;
    }
  }

  function handleDemoLaunch() {
    isModalOpen = false;
    launchDemoMode();
  }
</script>

<div class="flex h-screen w-screen overflow-hidden bg-slate-950 text-slate-100 font-sans">
  <!-- Collapsible Sidebar -->
  <Sidebar on:openModal={() => (isModalOpen = true)} />

  <!-- Main Application Content Shell -->
  <div class="flex-1 flex flex-col h-full overflow-hidden min-w-0 bg-slate-950/40">
    <Header
      on:openModal={() => (isModalOpen = true)}
    />

    <!-- Global Loading Status Banner -->
    {#if $isLoading}
      <div class="px-5 py-2 bg-cyan-950/80 border-b border-cyan-800/80 flex items-center justify-between text-xs text-cyan-200 z-10 animate-pulse">
        <div class="flex items-center gap-2">
          <svg class="w-4 h-4 text-cyan-400 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <span class="font-medium">{$loadingMessage || 'Processing repository data...'}</span>
        </div>
      </div>
    {/if}

    <!-- Global Error Message Banner -->
    {#if $errorMessage}
      <div class="px-5 py-2.5 bg-rose-950/90 border-b border-rose-800/80 flex items-center justify-between text-xs text-rose-200 z-10">
        <div class="flex items-center gap-2">
          <svg class="w-4 h-4 text-rose-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>{$errorMessage}</span>
        </div>
        <button
          class="text-rose-400 hover:text-rose-100 underline text-[11px]"
          on:click={() => errorMessage.set(null)}
        >
          Dismiss
        </button>
      </div>
    {/if}

    <!-- Screen Content Router -->
    <main class="flex-1 overflow-hidden relative">
      {#if !$selectedRepo}
        <EmptyState
          on:openRepo={() => (isModalOpen = true)}
          on:launchDemo={handleDemoLaunch}
        />
      {:else}
        {#if $currentView === 'overview'}
          <OverviewView />
        {:else if $currentView === 'timeline'}
          <TimelineView />
        {:else if $currentView === 'hotspots'}
          <HotspotsView />
        {:else if $currentView === 'dependencies'}
          <DependencyView />
        {:else if $currentView === 'commits'}
          <CommitExplorerView />
        {:else if $currentView === 'contributors'}
          <ContributorsView />
        {:else if $currentView === 'settings'}
          <SettingsView />
        {/if}
      {/if}
    </main>
  </div>
</div>

<!-- Modal: Open Repository -->
<Modal
  isOpen={isModalOpen}
  title="Analyze Local Git Repository"
  on:close={() => (isModalOpen = false)}
>
  <form on:submit|preventDefault={handleOpenRepoSubmit} class="space-y-4">
    <div>
      <label for="repo-path" class="block text-xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5">
        Repository Folder Path
      </label>
      <input
        id="repo-path"
        type="text"
        placeholder="e.g. C:/Users/Projects/my-app or /home/user/my-repo"
        bind:value={repoPathInput}
        class="w-full px-3 py-2 text-xs bg-slate-950 border border-slate-700 rounded-lg text-slate-200 placeholder-slate-500 font-mono focus:outline-none focus:border-cyan-500"
      />
      <p class="text-[11px] text-slate-500 mt-1">
        Must be an existing local directory containing a valid <code class="text-cyan-400">.git</code> folder.
      </p>
    </div>

    {#if modalError}
      <div class="p-2.5 rounded bg-rose-950/60 border border-rose-800/60 text-xs text-rose-300">
        {modalError}
      </div>
    {/if}

    <div class="flex items-center justify-between pt-3 border-t border-slate-800">
      <button
        type="button"
        class="text-xs text-emerald-400 hover:text-emerald-300 font-medium transition"
        on:click={handleDemoLaunch}
      >
        Or launch synthetic demo repository →
      </button>

      <div class="flex items-center gap-2">
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg text-xs font-medium text-slate-400 hover:text-slate-200 hover:bg-slate-800 transition"
          on:click={() => (isModalOpen = false)}
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isAnalyzing}
          class="px-4 py-1.5 rounded-lg text-xs font-medium bg-cyan-500 hover:bg-cyan-400 text-slate-950 transition font-semibold disabled:opacity-50"
        >
          {isAnalyzing ? 'Analyzing...' : 'Analyze Repository'}
        </button>
      </div>
    </div>
  </form>
</Modal>
