<script>
  import { activePanel, project, sourceFiles } from "../store.js";
  import { OpenProject, LoadProject, ListSourceFiles } from "../../wailsjs/go/main/App.js";
  import { transpileTabs, activeTranspileTab } from "../store.js";
  import { get } from "svelte/store";

  const navItems = [
    { id: "build",     label: "Build",     icon: "⚙" },
    { id: "transpile", label: "Transpile",  icon: "⇄" },
    { id: "viz",       label: "Visualize",  icon: "◈" },
    { id: "batch",     label: "Batch Viz",  icon: "⊞" },
  ];

  async function openProject() {
    const info = await OpenProject();
    if (info.error && info.error !== "cancelled") {
      alert("Error: " + info.error);
      return;
    }
    if (!info.error) {
      project.set(info);
      const files = await ListSourceFiles();
      sourceFiles.set(files);
    }
  }

  function openTranspile(f) {
    activePanel.set("transpile");
    const tabs = get(transpileTabs);
    const existing = tabs.find(t => t.path === f.path);
    if (existing) { activeTranspileTab.set(existing.id); return; }
    const id = "t-" + Date.now();
    transpileTabs.update(ts => [...ts, { id, label: f.rel.split("/").pop(), path: f.path, result: null }]);
    activeTranspileTab.set(id);
  }

  // Group files by extension
  function groupFiles(files) {
    const groups = {};
    for (const f of files) {
      const g = f.ext || "other";
      (groups[g] = groups[g] || []).push(f);
    }
    return Object.entries(groups).sort();
  }

  $: groups = groupFiles($sourceFiles);
</script>

<aside class="w-52 shrink-0 flex flex-col bg-gray-900 border-r border-gray-700 overflow-hidden">

  <!-- Nav -->
  <nav class="p-2 space-y-0.5 border-b border-gray-700">
    {#each navItems as item}
      <button
        class={$activePanel === item.id ? "sidebar-item-active w-full text-left" : "sidebar-item w-full text-left"}
        on:click={() => activePanel.set(item.id)}
      >
        <span class="text-base leading-none">{item.icon}</span>
        <span>{item.label}</span>
      </button>
    {/each}
  </nav>

  <!-- Project -->
  <div class="p-2 border-b border-gray-700">
    <button class="btn-primary w-full justify-center" on:click={openProject}>
      Open Project
    </button>
    {#if $project}
      <div class="mt-2 px-2 py-1.5 bg-gray-800 rounded text-xs">
        <div class="text-gray-200 font-medium truncate">{$project.name}</div>
        <div class="text-gray-500 truncate">{$project.root.split("/").pop()}</div>
      </div>
    {/if}
  </div>

  <!-- File tree -->
  <div class="flex-1 overflow-y-auto p-2 space-y-2">
    {#if $sourceFiles.length === 0}
      <p class="text-gray-600 text-xs px-2 pt-2">Open a project to see source files.</p>
    {:else}
      {#each groups as [ext, files]}
        <div>
          <div class="text-gray-600 text-xs uppercase tracking-wider px-2 py-1">{ext}</div>
          {#each files as f}
            <button
              class="sidebar-item w-full text-left truncate"
              on:click={() => openTranspile(f)}
              title={f.rel}
            >
              <span class="text-gray-500">›</span>
              <span class="truncate">{f.rel.split("/").pop()}</span>
            </button>
          {/each}
        </div>
      {/each}
    {/if}
  </div>
</aside>
