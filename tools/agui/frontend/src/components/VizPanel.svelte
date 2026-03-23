<script>
  import { vizFile, vizStage, vizMode, sourceFiles, refFolders } from "../store.js";
  import {
    VizTokens, VizAST, VizASTDot, VizSymbols, VizSymbolsDot, VizBlocking, VizEmit
  } from "../../wailsjs/go/main/App.js";
  import GraphView from "./GraphView.svelte";
  import { colorizeTokens, colorizeAST, colorizeSymbols, colorizeBlocking, colorizeEmit } from "../lib/highlight.js";

  const stages = [
    { id: "tokens",   label: "Tokens",   hasGraph: false },
    { id: "ast",      label: "AST",      hasGraph: true  },
    { id: "symbols",  label: "Symbols",  hasGraph: true  },
    { id: "blocking", label: "Blocking", hasGraph: false },
    { id: "emit",     label: "Emit View",hasGraph: false },
  ];

  const vizFns = {
    tokens:      VizTokens,
    ast:         VizAST,
    "ast-dot":   VizASTDot,
    symbols:     VizSymbols,
    "sym-dot":   VizSymbolsDot,
    blocking:    VizBlocking,
    emit:        VizEmit,
  };

  let output  = "";
  let dotSrc  = "";
  let loading = false;

  // Combine project files and ref folder files for the file picker
  $: allFiles = [
    ...$sourceFiles.map(f => ({ ...f, group: "Project" })),
    ...$refFolders.flatMap(rf => rf.files.map(f => ({ ...f, group: rf.name }))),
  ];

  async function run() {
    if (!$vizFile) return;
    loading = true;
    const stage = $vizStage;
    const mode  = $vizMode;
    if (mode === "graph") {
      const dotKey = stage === "ast" ? "ast-dot" : stage === "symbols" ? "sym-dot" : null;
      if (dotKey) dotSrc = await vizFns[dotKey]($vizFile);
    } else {
      output = await (vizFns[stage] ?? VizTokens)($vizFile);
    }
    loading = false;
  }

  $: currentStage   = stages.find(s => s.id === $vizStage);
  $: graphAvailable = currentStage?.hasGraph ?? false;
  $: if (!graphAvailable && $vizMode === "graph") vizMode.set("text");
  $: $vizFile, $vizStage, $vizMode, run();

  const colorizers = { tokens: colorizeTokens, ast: colorizeAST, symbols: colorizeSymbols, blocking: colorizeBlocking, emit: colorizeEmit };
  $: coloredOutput = $vizMode === "text" && colorizers[$vizStage] ? colorizers[$vizStage](output) : null;
</script>

<div class="h-full flex flex-col">

  <!-- Toolbar -->
  <div class="flex items-center gap-2 px-4 py-2 app-header border-b app-border shrink-0 flex-wrap">

    <select class="app-select text-xs rounded px-2 py-1 max-w-xs" bind:value={$vizFile}>
      <option value={null}>— select file —</option>
      {#each allFiles as f}
        <option value={f.path}>[{f.group}] {f.rel}</option>
      {/each}
    </select>

    <div class="flex border app-border rounded overflow-hidden">
      {#each stages as s}
        <button
          class="px-3 py-1 text-xs border-r app-border last:border-r-0 transition-colors
                 {$vizStage === s.id ? 'bg-violet-700 text-white' : 'app-card app-dim hover:app-text'}"
          on:click={() => vizStage.set(s.id)}
        >{s.label}</button>
      {/each}
    </div>

    <div class="flex border app-border rounded overflow-hidden ml-auto">
      <button
        class="px-3 py-1 text-xs border-r app-border transition-colors
               {$vizMode === 'text' ? 'bg-violet-700 text-white' : 'app-card app-dim hover:app-text'}"
        on:click={() => vizMode.set("text")}
      >Text</button>
      <button
        class="px-3 py-1 text-xs transition-colors
               {$vizMode === 'graph' ? 'bg-violet-700 text-white' : 'app-card app-dim hover:app-text'}
               {!graphAvailable ? 'opacity-30 cursor-not-allowed' : ''}"
        on:click={() => graphAvailable && vizMode.set("graph")}
        disabled={!graphAvailable}
      >Graph</button>
    </div>
  </div>

  <!-- Output -->
  <div class="flex-1 overflow-hidden">
    {#if !$vizFile}
      <div class="h-full flex items-center justify-center app-dim text-sm">
        Select a file and stage above.
      </div>
    {:else if loading}
      <div class="h-full flex items-center justify-center app-dim text-sm animate-pulse">
        Running…
      </div>
    {:else if $vizMode === "graph" && graphAvailable}
      <GraphView dot={dotSrc} />
    {:else if coloredOutput !== null}
      <pre class="h-full overflow-auto p-4 text-xs app-code leading-5">{@html coloredOutput}</pre>
    {:else}
      <pre class="h-full overflow-auto p-4 text-xs app-code leading-5">{output}</pre>
    {/if}
  </div>
</div>
