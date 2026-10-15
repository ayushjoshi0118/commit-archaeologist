<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  export let isOpen = false;
  export let title = '';

  const dispatch = createEventDispatcher();

  function close() {
    dispatch('close');
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') close();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if isOpen}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <!-- Backdrop -->
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div
      class="fixed inset-0 bg-black/70 backdrop-blur-sm transition-opacity"
      on:click={close}
    ></div>

    <!-- Modal Content -->
    <div class="relative w-full max-w-lg rounded-xl border border-slate-700/80 bg-slate-900 p-6 shadow-2xl z-10 text-slate-100">
      <div class="flex items-center justify-between pb-4 border-b border-slate-800">
        <h3 class="text-base font-semibold text-slate-100 tracking-wide">{title}</h3>
        <button
          class="rounded p-1 text-slate-400 hover:bg-slate-800 hover:text-slate-200 transition"
          on:click={close}
          aria-label="Close modal"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <div class="mt-4">
        <slot />
      </div>
    </div>
  </div>
{/if}
