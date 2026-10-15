<script lang="ts">
  export let height = 'h-6';
  export let width = 'w-full';
  export let count = 1;
</script>

<div class="flex flex-col gap-3 {width}">
  {#each Array(count) as _}
    <div class="skeleton {height} rounded-md border border-slate-800/60 bg-slate-800/40"></div>
  {/each}
</div>
