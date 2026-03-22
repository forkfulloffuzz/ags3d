<script>
  import { logLines, clearLog } from "../store.js";
  import { afterUpdate } from "svelte";

  export let height = 160;

  let el;
  let autoScroll = true;

  afterUpdate(() => {
    if (autoScroll && el) el.scrollTop = el.scrollHeight;
  });

  function onScroll() {
    if (!el) return;
    autoScroll = el.scrollHeight - el.scrollTop - el.clientHeight < 20;
  }
</script>

<section class="flex flex-col bg-gray-950 border-t border-gray-800 shrink-0" style="height:{height}px">
  <div class="flex items-center gap-2 px-3 py-1 bg-gray-900 border-b border-gray-800 shrink-0">
    <span class="text-gray-500 text-xs font-medium uppercase tracking-wider">Log</span>
    <span class="flex-1"></span>
    <button class="btn-ghost text-xs py-0.5 px-2" on:click={clearLog}>Clear</button>
  </div>
  <div
    class="flex-1 overflow-y-auto p-2 font-mono text-xs space-y-px"
    bind:this={el}
    on:scroll={onScroll}
  >
    {#each $logLines as line}
      <div class="log-{line.kind} leading-5">{line.msg}</div>
    {:else}
      <div class="text-gray-700 italic">No output yet.</div>
    {/each}
  </div>
</section>
