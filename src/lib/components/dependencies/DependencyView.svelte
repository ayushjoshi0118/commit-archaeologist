<script lang="ts">
  import { onMount } from 'svelte';
  import * as d3 from 'd3';
  import type { DependencyGraph, DependencyNode, DependencyLink } from '../../types';
  import { selectedRepo, currentView, fileFilter } from '../../stores/app';
  import { api } from '../../api/client';
  import SkeletonLoader from '../common/SkeletonLoader.svelte';

  let svgElement: SVGSVGElement;
  let graphData: DependencyGraph | null = null;
  let loading = true;
  let selectedNode: DependencyNode | null = null;
  let searchFilter = '';

  const langColorMap: Record<string, string> = {
    go: '#00add8',
    typescript: '#3178c6',
    javascript: '#f7df1e',
    python: '#3572a5',
    unknown: '#94a3b8'
  };

  onMount(async () => {
    await fetchGraph();
  });

  $: if ($selectedRepo?.id) {
    fetchGraph();
  }

  async function fetchGraph() {
    if (!$selectedRepo) return;
    loading = true;
    try {
      graphData = await api.getDependencies($selectedRepo.id);
      renderForceGraph();
    } catch (err) {
      console.error('Failed to load dependency graph:', err);
    } finally {
      loading = false;
    }
  }

  function renderForceGraph() {
    if (!svgElement || !graphData || graphData.nodes.length === 0) return;

    const width = svgElement.clientWidth || 900;
    const height = 550;

    const svg = d3.select(svgElement);
    svg.selectAll('*').remove();
    svg.attr('viewBox', `0 0 ${width} ${height}`);

    const g = svg.append('g');

    // Zoom & Pan
    const zoom = d3.zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.2, 5])
      .on('zoom', (event) => {
        g.attr('transform', event.transform);
      });
    svg.call(zoom);

    // Arrow markers for directed edges
    const defs = svg.append('defs');
    defs.append('marker')
      .attr('id', 'arrow')
      .attr('viewBox', '0 -5 10 10')
      .attr('refX', 18)
      .attr('refY', 0)
      .attr('markerWidth', 6)
      .attr('markerHeight', 6)
      .attr('orient', 'auto')
      .append('path')
      .attr('d', 'M0,-5L10,0L0,5')
      .attr('fill', '#475569');

    defs.append('marker')
      .attr('id', 'arrow-active')
      .attr('viewBox', '0 -5 10 10')
      .attr('refX', 18)
      .attr('refY', 0)
      .attr('markerWidth', 6)
      .attr('markerHeight', 6)
      .attr('orient', 'auto')
      .append('path')
      .attr('d', 'M0,-5L10,0L0,5')
      .attr('fill', '#06b6d4');

    // Clone data for simulation
    const nodes: DependencyNode[] = graphData.nodes.map(d => ({ ...d }));
    const links: any[] = graphData.links.map(d => ({ ...d }));

    const simulation = d3.forceSimulation(nodes)
      .force('link', d3.forceLink(links).id((d: any) => d.id).distance(90))
      .force('charge', d3.forceManyBody().strength(-280))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .force('collision', d3.forceCollide().radius(25));

    // Links
    const link = g.append('g')
      .selectAll('line')
      .data(links)
      .enter()
      .append('line')
      .attr('stroke', '#334155')
      .attr('stroke-width', 1.5)
      .attr('marker-end', 'url(#arrow)');

    // Nodes
    const node = g.append('g')
      .selectAll('.node')
      .data(nodes)
      .enter()
      .append('g')
      .attr('class', 'node')
      .style('cursor', 'grab')
      .call(
        d3.drag<any, DependencyNode>()
          .on('start', (event, d: any) => {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
          })
          .on('drag', (event, d: any) => {
            d.fx = event.x;
            d.fy = event.y;
          })
          .on('end', (event, d: any) => {
            if (!event.active) simulation.alphaTarget(0);
            d.fx = null;
            d.fy = null;
          })
      );

    // Node circles
    node.append('circle')
      .attr('r', d => 6 + Math.min((d.inDegree + d.outDegree) * 2, 16))
      .attr('fill', d => langColorMap[d.language] || '#94a3b8')
      .attr('fill-opacity', 0.8)
      .attr('stroke', '#0f172a')
      .attr('stroke-width', 2);

    // Node labels
    node.append('text')
      .attr('dx', 12)
      .attr('dy', 4)
      .attr('font-size', '10px')
      .attr('font-family', 'monospace')
      .attr('fill', '#cbd5e1')
      .text(d => d.name.split('/').pop() || d.name);

    // Click handler
    node.on('click', (_, d) => {
      selectedNode = d;
      highlightNeighbors(d);
    });

    function highlightNeighbors(target: DependencyNode) {
      const neighborIds = new Set<string>();
      neighborIds.add(target.id);

      links.forEach(l => {
        const sId = typeof l.source === 'object' ? l.source.id : l.source;
        const tId = typeof l.target === 'object' ? l.target.id : l.target;
        if (sId === target.id) neighborIds.add(tId);
        if (tId === target.id) neighborIds.add(sId);
      });

      node.selectAll('circle')
        .attr('stroke', (d: any) => (neighborIds.has(d.id) ? '#38bdf8' : '#0f172a'))
        .attr('stroke-width', (d: any) => (d.id === target.id ? 3 : neighborIds.has(d.id) ? 2 : 1))
        .attr('fill-opacity', (d: any) => (neighborIds.has(d.id) ? 1.0 : 0.25));

      link
        .attr('stroke', (l: any) => {
          const sId = typeof l.source === 'object' ? l.source.id : l.source;
          const tId = typeof l.target === 'object' ? l.target.id : l.target;
          return sId === target.id || tId === target.id ? '#06b6d4' : '#1e293b';
        })
        .attr('stroke-width', (l: any) => {
          const sId = typeof l.source === 'object' ? l.source.id : l.source;
          const tId = typeof l.target === 'object' ? l.target.id : l.target;
          return sId === target.id || tId === target.id ? 2.5 : 1;
        })
        .attr('marker-end', (l: any) => {
          const sId = typeof l.source === 'object' ? l.source.id : l.source;
          const tId = typeof l.target === 'object' ? l.target.id : l.target;
          return sId === target.id || tId === target.id ? 'url(#arrow-active)' : 'url(#arrow)';
        });
    }

    simulation.on('tick', () => {
      link
        .attr('x1', (d: any) => d.source.x)
        .attr('y1', (d: any) => d.source.y)
        .attr('x2', (d: any) => d.target.x)
        .attr('y2', (d: any) => d.target.y);

      node.attr('transform', (d: any) => `translate(${d.x},${d.y})`);
    });
  }

  function viewFileCommits(path: string) {
    fileFilter.set(path);
    currentView.set('commits');
  }
</script>

<div class="flex h-full overflow-hidden">
  <!-- Force Graph Area -->
  <div class="flex-1 flex flex-col h-full overflow-hidden p-6">
    <div class="flex items-center justify-between pb-3 border-b border-slate-800/80 mb-3">
      <div>
        <h2 class="text-base font-semibold text-slate-100">AST Source Dependency Graph</h2>
        <p class="text-xs text-slate-400">
          Parsed cross-file imports & module dependencies for Go, TypeScript, JavaScript, and Python.
        </p>
      </div>

      <div class="flex items-center gap-3 text-xs">
        <div class="flex items-center gap-2">
          <span class="inline-flex items-center gap-1 text-[11px] text-slate-400">
            <span class="w-2.5 h-2.5 rounded-full bg-[#00add8]"></span> Go
          </span>
          <span class="inline-flex items-center gap-1 text-[11px] text-slate-400">
            <span class="w-2.5 h-2.5 rounded-full bg-[#3178c6]"></span> TypeScript
          </span>
          <span class="inline-flex items-center gap-1 text-[11px] text-slate-400">
            <span class="w-2.5 h-2.5 rounded-full bg-[#3572a5]"></span> Python
          </span>
        </div>
      </div>
    </div>

    {#if loading}
      <div class="flex-1 flex items-center justify-center">
        <SkeletonLoader count={4} height="h-28" width="w-3/4" />
      </div>
    {:else if !graphData || graphData.nodes.length === 0}
      <div class="flex-1 flex flex-col items-center justify-center text-slate-500 text-sm">
        <svg class="w-10 h-10 mb-2 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101" />
        </svg>
        <p>No supported language dependencies found in this repository.</p>
        <p class="text-xs mt-1 text-slate-600">Supported: Go, TypeScript, JavaScript, Python</p>
      </div>
    {:else}
      <div class="flex-1 w-full bg-slate-950/60 rounded-xl border border-slate-800/80 relative overflow-hidden shadow-inner">
        <svg bind:this={svgElement} class="w-full h-full"></svg>
      </div>
    {/if}
  </div>

  <!-- Detail Inspector Drawer -->
  {#if selectedNode}
    <div class="w-80 md:w-96 border-l border-slate-800/80 bg-slate-900/90 backdrop-blur-md p-5 flex flex-col justify-between overflow-y-auto">
      <div class="space-y-4">
        <div class="flex items-center justify-between pb-3 border-b border-slate-800">
          <span class="text-xs font-semibold uppercase tracking-wider text-slate-400">Selected Node</span>
          <span class="px-2 py-0.5 rounded text-[11px] font-mono uppercase font-bold text-slate-200 border border-slate-700">
            {selectedNode.language}
          </span>
        </div>

        <div>
          <span class="text-[10px] text-slate-500 uppercase font-semibold">File Path</span>
          <p class="font-mono text-xs text-slate-200 break-all bg-slate-950/60 p-2 rounded border border-slate-800 mt-1">
            {selectedNode.id}
          </p>
        </div>

        <div class="grid grid-cols-2 gap-3 text-center text-xs">
          <div class="p-3 rounded-lg bg-slate-800/40 border border-slate-800">
            <span class="block text-slate-500 text-[10px] uppercase font-semibold">In-Degree (Dependents)</span>
            <strong class="text-slate-200 font-mono text-lg">{selectedNode.inDegree}</strong>
          </div>
          <div class="p-3 rounded-lg bg-slate-800/40 border border-slate-800">
            <span class="block text-slate-500 text-[10px] uppercase font-semibold">Out-Degree (Imports)</span>
            <strong class="text-slate-200 font-mono text-lg">{selectedNode.outDegree}</strong>
          </div>
        </div>

        <!-- Links list -->
        {#if graphData}
          {@const outgoing = graphData.links.filter(l => (typeof l.source === 'object' ? (l.source as any).id : l.source) === selectedNode?.id)}
          {@const incoming = graphData.links.filter(l => (typeof l.target === 'object' ? (l.target as any).id : l.target) === selectedNode?.id)}

          {#if outgoing.length > 0}
            <div>
              <span class="text-[11px] font-semibold text-slate-400 block mb-1">Imports ({outgoing.length}):</span>
              <div class="space-y-1 max-h-32 overflow-y-auto">
                {#each outgoing as link}
                  {@const targetId = typeof link.target === 'object' ? (link.target as any).id : link.target}
                  <div class="text-[10px] font-mono text-slate-300 truncate bg-slate-950/40 p-1 rounded">
                    → {targetId}
                  </div>
                {/each}
              </div>
            </div>
          {/if}

          {#if incoming.length > 0}
            <div>
              <span class="text-[11px] font-semibold text-slate-400 block mb-1">Imported by ({incoming.length}):</span>
              <div class="space-y-1 max-h-32 overflow-y-auto">
                {#each incoming as link}
                  {@const sourceId = typeof link.source === 'object' ? (link.source as any).id : link.source}
                  <div class="text-[10px] font-mono text-slate-300 truncate bg-slate-950/40 p-1 rounded">
                    ← {sourceId}
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        {/if}
      </div>

      <div class="pt-4 mt-6 border-t border-slate-800">
        <button
          class="w-full flex items-center justify-center gap-2 py-2 px-4 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-medium text-xs transition"
          on:click={() => viewFileCommits(selectedNode?.id || '')}
        >
          <span>View Commit History for File</span>
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3"/>
          </svg>
        </button>
      </div>
    </div>
  {/if}
</div>
