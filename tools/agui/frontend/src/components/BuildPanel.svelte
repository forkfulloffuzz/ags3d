<script>
  import { project } from "../store.js";
  import { Build } from "../../wailsjs/go/main/App.js";
  import { Hammer } from "lucide-svelte";

  let building = false;

  async function runBuild() {
    if (!$project) { alert("Open a project first."); return; }
    building = true;
    await Build();
    building = false;
  }
</script>

<div class="h-full flex flex-col p-6 gap-6 overflow-y-auto">
  <div>
    <h2 class="text-base font-semibold app-text mb-1">Build</h2>
    <p class="app-dim text-xs">Transpile changed .agscript files to GDScript.</p>
  </div>

  {#if $project}
    <div class="app-card border border-app rounded p-4 space-y-2 text-xs">
      <div class="grid grid-cols-2 gap-x-4 gap-y-1">
        <span class="app-dim">Project</span>
        <span class="app-text font-medium">{$project.name}</span>
        <span class="app-dim">Root</span>
        <span class="app-dim truncate" title={$project.root}>{$project.root}</span>
        {#if $project.startRoom}
          <span class="app-dim">Start room</span>
          <span class="app-dim">{$project.startRoom}</span>
        {/if}
        {#if $project.renderingMode}
          <span class="app-dim">Rendering</span>
          <span class="app-dim">{$project.renderingMode}</span>
        {/if}
      </div>
    </div>
  {:else}
    <div class="app-dim text-xs app-card border border-app rounded p-4">
      No project open. Use <strong class="app-text">Open Project</strong> in the sidebar.
    </div>
  {/if}

  <div class="flex gap-3 flex-wrap">
    <button class="btn-primary" on:click={runBuild} disabled={building || !$project}>
      {#if building}
        <span class="animate-spin inline-block">↻</span> Building…
      {:else}
        <Hammer size={13} /> Build
      {/if}
    </button>
  </div>

  <p class="app-muted text-xs">
    Output streams to the log panel below. Shortcut: <kbd class="app-card border border-app rounded px-1">Ctrl+B</kbd>
  </p>
</div>
