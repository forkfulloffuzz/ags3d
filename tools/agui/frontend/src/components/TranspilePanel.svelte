<script>
  import { transpileTabs, activeTranspileTab } from "../store.js";
  import { TranspileFile } from "../../wailsjs/go/main/App.js";
  import TranspileTab from "./TranspileTab.svelte";
  import { get } from "svelte/store";

  function closeTab(id) {
    transpileTabs.update(ts => ts.filter(t => t.id !== id));
    const remaining = get(transpileTabs);
    if (get(activeTranspileTab) === id) {
      activeTranspileTab.set(remaining.length ? remaining[remaining.length - 1].id : null);
    }
  }

  // Load result when a tab becomes active and has no result yet
  $: {
    const id = $activeTranspileTab;
    if (id) {
      const tab = $transpileTabs.find(t => t.id === id);
      if (tab && !tab.result && !tab.loading) {
        tab.loading = true;
        transpileTabs.update(ts => ts); // force reactivity
        TranspileFile(tab.path).then(result => {
          transpileTabs.update(ts =>
            ts.map(t => t.id === id ? { ...t, result, loading: false } : t)
          );
        });
      }
    }
  }

  $: activeTab = $transpileTabs.find(t => t.id === $activeTranspileTab);
</script>

<div class="h-full flex flex-col">

  {#if $transpileTabs.length === 0}
    <div class="h-full flex items-center justify-center text-gray-600 text-sm">
      Click a source file in the sidebar to open it here.
    </div>
  {:else}
    <!-- Tab bar -->
    <div class="flex items-end border-b border-gray-700 bg-gray-900 px-2 overflow-x-auto shrink-0">
      {#each $transpileTabs as tab}
        <div class="flex items-center">
          <button
            class={$activeTranspileTab === tab.id ? "tab-btn-active" : "tab-btn"}
            on:click={() => activeTranspileTab.set(tab.id)}
          >{tab.label}</button>
          <button
            class="text-gray-600 hover:text-gray-300 text-xs px-1 pb-1"
            on:click|stopPropagation={() => closeTab(tab.id)}
            title="Close"
          >✕</button>
        </div>
      {/each}
    </div>

    <!-- Active tab content -->
    <div class="flex-1 min-h-0 overflow-hidden">
      {#if activeTab}
        {#if activeTab.loading}
          <div class="h-full flex items-center justify-center text-gray-600 text-sm animate-pulse">
            Transpiling…
          </div>
        {:else if activeTab.result}
          <TranspileTab result={activeTab.result} />
        {/if}
      {/if}
    </div>
  {/if}
</div>
