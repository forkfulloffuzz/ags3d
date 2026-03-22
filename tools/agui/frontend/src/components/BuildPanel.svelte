<script>
  import { project } from "../store.js";
  import { Build } from "../../wailsjs/go/main/App.js";

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
    <h2 class="text-base font-semibold text-gray-100 mb-1">Build</h2>
    <p class="text-gray-500 text-xs">Transpile changed .agscript files to GDScript.</p>
  </div>

  {#if $project}
    <div class="bg-gray-900 border border-gray-700 rounded p-4 space-y-2 text-xs">
      <div class="grid grid-cols-2 gap-x-4 gap-y-1">
        <span class="text-gray-500">Project</span>
        <span class="text-gray-200 font-medium">{$project.name}</span>
        <span class="text-gray-500">Root</span>
        <span class="text-gray-400 truncate" title={$project.root}>{$project.root}</span>
        {#if $project.startRoom}
          <span class="text-gray-500">Start room</span>
          <span class="text-gray-400">{$project.startRoom}</span>
        {/if}
        {#if $project.renderingMode}
          <span class="text-gray-500">Rendering</span>
          <span class="text-gray-400">{$project.renderingMode}</span>
        {/if}
      </div>
    </div>
  {:else}
    <div class="text-gray-600 text-xs bg-gray-900 border border-gray-700 rounded p-4">
      No project open. Use <strong class="text-gray-400">Open Project</strong> in the sidebar.
    </div>
  {/if}

  <div class="flex gap-3 flex-wrap">
    <button class="btn-primary" on:click={runBuild} disabled={building || !$project}>
      {#if building}
        <span class="animate-spin">↻</span> Building…
      {:else}
        ⚙ Build
      {/if}
    </button>
  </div>

  <p class="text-gray-700 text-xs">Output is streamed to the log panel below.</p>
</div>
