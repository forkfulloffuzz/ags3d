<script>
  import { sourceFiles } from "../store.js";
  import { BatchViz, ListSourceFiles } from "../../wailsjs/go/main/App.js";
  import { EventsOn } from "../../wailsjs/runtime/runtime.js";
  import { onMount } from "svelte";
  import { addLog } from "../store.js";

  const stages = ["all", "tokens", "ast", "ast-dot", "symbols", "symbols-dot", "blocking", "emit"];
  let stage = "all";
  let running = false;
  let results = [];   // BatchVizResult[]
  let selected = null; // selected result
  let activeStage = "tokens";

  onMount(() => {
    EventsOn("batchviz:result", r => {
      results = [...results, r];
      addLog("info", `  ${r.rel} — ${Object.keys(r.stages).length} stage(s)`);
    });
    EventsOn("batchviz:done", count => {
      addLog("ok", `Batch viz done: ${count} file(s)`);
      running = false;
    });
  });

  async function run() {
    if (!$sourceFiles.length) { alert("Open a project first."); return; }
    results = [];
    selected = null;
    running = true;
    addLog("info", `Batch viz (${stage}) — ${$sourceFiles.length} files…`);
    await BatchViz($sourceFiles.map(f => f.path), stage);
  }

  $: selectedStages = selected ? Object.keys(selected.stages) : [];
  $: if (selected && !selectedStages.includes(activeStage)) activeStage = selectedStages[0] ?? "";
</script>

<div class="h-full flex flex-col">

  <!-- Toolbar -->
  <div class="flex items-center gap-3 px-4 py-2 bg-gray-900 border-b border-gray-700 shrink-0">
    <select
      class="bg-gray-800 border border-gray-600 text-gray-200 text-xs rounded px-2 py-1"
      bind:value={stage}
    >
      {#each stages as s}
        <option value={s}>{s}</option>
      {/each}
    </select>
    <button class="btn-primary" on:click={run} disabled={running}>
      {#if running}<span class="animate-spin">↻</span> Running…{:else}▶ Run{/if}
    </button>
    <span class="text-gray-600 text-xs ml-auto">{results.length} / {$sourceFiles.length} files</span>
  </div>

  <div class="flex flex-1 min-h-0">

    <!-- File list -->
    <div class="w-56 shrink-0 border-r border-gray-700 overflow-y-auto bg-gray-900">
      {#each results as r}
        <button
          class="w-full text-left px-3 py-2 text-xs border-b border-gray-800 hover:bg-gray-800 transition-colors
                 {selected?.file === r.file ? 'bg-gray-800 text-violet-300' : 'text-gray-400'}"
          on:click={() => selected = r}
        >
          <div class="truncate">{r.rel}</div>
          {#if r.error}
            <div class="text-red-400 truncate">{r.error}</div>
          {:else}
            <div class="text-gray-600">{Object.keys(r.stages).join(", ")}</div>
          {/if}
        </button>
      {:else}
        <p class="text-gray-600 text-xs p-4">No results yet.</p>
      {/each}
    </div>

    <!-- Stage output -->
    <div class="flex-1 flex flex-col min-w-0">
      {#if selected}
        <!-- Stage tabs -->
        <div class="flex border-b border-gray-700 bg-gray-900 px-2 overflow-x-auto shrink-0">
          {#each selectedStages as s}
            <button
              class={activeStage === s ? "tab-btn-active" : "tab-btn"}
              on:click={() => activeStage = s}
            >{s}</button>
          {/each}
        </div>
        <pre class="flex-1 overflow-auto p-4 text-xs text-gray-300 mono leading-5">{selected.stages[activeStage] ?? ""}</pre>
      {:else}
        <div class="h-full flex items-center justify-center text-gray-600 text-sm">
          Run batch viz and select a file.
        </div>
      {/if}
    </div>
  </div>
</div>
