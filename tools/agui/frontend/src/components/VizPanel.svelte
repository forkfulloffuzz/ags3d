<script>
  import { vizFile, vizStage, vizMode, sourceFiles } from "../store.js";
  import {
    VizTokens, VizAST, VizASTDot, VizSymbols, VizSymbolsDot, VizBlocking, VizEmit
  } from "../../wailsjs/go/main/App.js";
  import GraphView from "./GraphView.svelte";

  const stages = [
    { id: "tokens",      label: "Tokens",    hasGraph: false },
    { id: "ast",         label: "AST",        hasGraph: true  },
    { id: "symbols",     label: "Symbols",    hasGraph: true  },
    { id: "blocking",    label: "Blocking",   hasGraph: false },
    { id: "emit",        label: "Emit View",  hasGraph: false },
  ];

  const vizFns = {
    tokens:   VizTokens,
    ast:      VizAST,
    "ast-dot": VizASTDot,
    symbols:  VizSymbols,
    "sym-dot": VizSymbolsDot,
    blocking: VizBlocking,
    emit:     VizEmit,
  };

  let output = "";
  let dotSrc = "";
  let loading = false;

  async function run() {
    if (!$vizFile) return;
    loading = true;
    const stage = $vizStage;
    const mode = $vizMode;
    if (mode === "graph") {
      const dotKey = stage === "ast" ? "ast-dot" : stage === "symbols" ? "sym-dot" : null;
      if (dotKey) dotSrc = await vizFns[dotKey]($vizFile);
    } else {
      output = await (vizFns[stage] ?? VizTokens)($vizFile);
    }
    loading = false;
  }

  $: currentStage = stages.find(s => s.id === $vizStage);
  $: graphAvailable = currentStage?.hasGraph ?? false;
  $: if (!graphAvailable && $vizMode === "graph") vizMode.set("text");
  $: $vizFile, $vizStage, $vizMode, run();
</script>

<div class="h-full flex flex-col">

  <!-- Toolbar -->
  <div class="flex items-center gap-2 px-4 py-2 bg-gray-900 border-b border-gray-700 shrink-0 flex-wrap">

    <!-- File picker -->
    <select
      class="bg-gray-800 border border-gray-600 text-gray-200 text-xs rounded px-2 py-1 max-w-xs"
      bind:value={$vizFile}
    >
      <option value={null}>— select file —</option>
      {#each $sourceFiles as f}
        <option value={f.path}>{f.rel}</option>
      {/each}
    </select>

    <!-- Stage tabs -->
    <div class="flex border border-gray-700 rounded overflow-hidden">
      {#each stages as s}
        <button
          class="px-3 py-1 text-xs border-r border-gray-700 last:border-r-0 transition-colors
                 {$vizStage === s.id ? 'bg-violet-700 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700 hover:text-gray-200'}"
          on:click={() => vizStage.set(s.id)}
        >{s.label}</button>
      {/each}
    </div>

    <!-- Text / Graph toggle -->
    <div class="flex border border-gray-700 rounded overflow-hidden ml-auto">
      <button
        class="px-3 py-1 text-xs border-r border-gray-700 transition-colors
               {$vizMode === 'text' ? 'bg-violet-700 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'}"
        on:click={() => vizMode.set("text")}
      >Text</button>
      <button
        class="px-3 py-1 text-xs transition-colors
               {$vizMode === 'graph' ? 'bg-violet-700 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'}
               {!graphAvailable ? 'opacity-30 cursor-not-allowed' : ''}"
        on:click={() => graphAvailable && vizMode.set("graph")}
        disabled={!graphAvailable}
      >Graph</button>
    </div>
  </div>

  <!-- Output -->
  <div class="flex-1 overflow-hidden">
    {#if !$vizFile}
      <div class="h-full flex items-center justify-center text-gray-600 text-sm">
        Select a file and stage above.
      </div>
    {:else if loading}
      <div class="h-full flex items-center justify-center text-gray-600 text-sm animate-pulse">
        Running…
      </div>
    {:else if $vizMode === "graph" && graphAvailable}
      <GraphView dot={dotSrc} />
    {:else}
      <pre class="h-full overflow-auto p-4 text-xs text-gray-300 mono leading-5">{output}</pre>
    {/if}
  </div>
</div>
