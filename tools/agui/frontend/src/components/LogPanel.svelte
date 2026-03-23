<script>
  import { logLines, clearLog } from "../store.js";
  import { afterUpdate } from "svelte";
  import { Trash2 } from "lucide-svelte";

  export let height = 160;

  let container;
  let autoScroll = true;

  afterUpdate(() => {
    if (autoScroll && container) container.scrollTop = container.scrollHeight;
  });
</script>

<div class="flex flex-col border-t app-border shrink-0" style="height:{height}px">
  <div class="flex items-center px-3 py-1 app-header border-b app-border gap-2 shrink-0">
    <span class="text-xs font-semibold uppercase tracking-wider app-dim">Log</span>
    <label class="ml-2 flex items-center gap-1 text-xs app-dim cursor-pointer">
      <input type="checkbox" bind:checked={autoScroll} class="accent-violet-500" />
      Auto-scroll
    </label>
    <button
      class="ml-auto btn-ghost py-0.5 px-1.5 text-xs"
      on:click={clearLog}
      title="Clear log"
    >
      <Trash2 size={11} />
    </button>
  </div>
  <div bind:this={container} class="flex-1 overflow-y-auto font-mono text-xs p-2 space-y-0.5 app-code">
    {#each $logLines as line}
      <div class="log-{line.kind} leading-5">{line.msg}</div>
    {:else}
      <div class="app-muted">No output yet.</div>
    {/each}
  </div>
</div>
