<script>
  import { onMount } from "svelte";
  import { Graphviz } from "@hpcc-js/wasm-graphviz";

  export let dot = "";

  let container;
  let svg = "";
  let error = "";

  // Module-level singleton — loaded once, never released.
  // Releasing the WASM worker on every unmount freezes subsequent mounts.
  let _gvReady = null;
  function getGv() {
    if (!_gvReady) _gvReady = Graphviz.load();
    return _gvReady;
  }

  let gv;

  onMount(async () => {
    gv = await getGv();
    render();
  });

  function render() {
    if (!gv || !dot) return;
    try {
      svg = gv.dot(dot);
      error = "";
    } catch (e) {
      error = String(e);
      svg = "";
    }
  }

  $: dot, gv && render();
</script>

<div class="h-full overflow-auto app-bg p-4" bind:this={container}>
  {#if error}
    <pre class="text-red-400 text-xs p-4">{error}</pre>
  {:else if svg}
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div class="svg-wrap flex justify-center">
      {@html svg}
    </div>
  {:else}
    <div class="text-gray-600 text-sm flex items-center justify-center h-full">
      No graph output.
    </div>
  {/if}
</div>

<style>
  .svg-wrap :global(svg) {
    max-width: 100%;
    height: auto;
    filter: invert(0.85) hue-rotate(180deg);
  }
</style>
