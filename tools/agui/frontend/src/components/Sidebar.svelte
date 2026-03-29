<script>
  import {
    activePanel, project, sourceFiles, refFolders,
    recentProjects, addRecentProject,
  } from "../store.js";
  import {
    OpenProject, LoadProject, ListSourceFiles,
    OpenRefFolder, ListRefFiles,
  } from "../../wailsjs/go/main/App.js";
  import { transpileTabs, activeTranspileTab } from "../store.js";
  import { get } from "svelte/store";
  import {
    Hammer, Repeat2, Workflow, LayoutGrid, MapPin,
    FolderOpen, ChevronDown, ChevronRight,
    FileCode, LayoutTemplate, User, Package, MessageSquare, File,
    Clock, X,
  } from "lucide-svelte";

  const navItems = [
    { id: "build",     label: "Build",      Icon: Hammer },
    { id: "transpile", label: "Transpile",   Icon: Repeat2 },
    { id: "viz",       label: "Visualize",   Icon: Workflow },
    { id: "batch",     label: "Batch Viz",   Icon: LayoutGrid },
    { id: "roomchar",  label: "Rooms/Chars", Icon: MapPin },
  ];

  const fileIcons = {
    ".agscript": FileCode,
    ".agroom":   LayoutTemplate,
    ".agchar":   User,
    ".agitem":   Package,
    ".agdlg":    MessageSquare,
  };
  function fileIcon(ext) { return fileIcons[ext] ?? File; }

  // Collapsible section state
  let projectOpen = true;
  let refsOpen    = true;
  let recentsOpen = false;

  async function openProject() {
    const info = await OpenProject();
    if (info.error && info.error !== "cancelled") { alert("Error: " + info.error); return; }
    if (!info.error) {
      project.set(info);
      const files = await ListSourceFiles();
      sourceFiles.set(files);
      addRecentProject({ root: info.root, name: info.name });
    }
  }

  async function loadRecent(root) {
    const info = await LoadProject(root);
    if (info.error) { alert("Error: " + info.error); return; }
    project.set(info);
    const files = await ListSourceFiles();
    sourceFiles.set(files);
    addRecentProject({ root: info.root, name: info.name });
  }

  async function addRefFolder() {
    const info = await OpenRefFolder();
    if (info.error === "cancelled" || info.error) return;
    const files = await ListRefFiles(info.root);
    refFolders.update(fs => {
      if (fs.find(f => f.root === info.root)) return fs;
      return [...fs, { root: info.root, name: info.name, files: files || [], open: true }];
    });
  }

  function removeRefFolder(root) { refFolders.update(fs => fs.filter(f => f.root !== root)); }
  function toggleRefFolder(root) {
    refFolders.update(fs => fs.map(f => f.root === root ? {...f, open: !f.open} : f));
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

  function groupFiles(files) {
    const groups = {};
    for (const f of files) { const g = f.ext || "other"; (groups[g] = groups[g] || []).push(f); }
    return Object.entries(groups).sort();
  }

  $: projectGroups = groupFiles($sourceFiles);
</script>

<aside class="w-52 shrink-0 flex flex-col app-sidebar border-r overflow-hidden">

  <!-- Nav -->
  <nav class="p-2 space-y-0.5 border-b app-border">
    {#each navItems as { id, label, Icon }}
      <button
        class={$activePanel === id ? "sidebar-item-active w-full text-left" : "sidebar-item w-full text-left"}
        on:click={() => activePanel.set(id)}
      >
        <svelte:component this={Icon} size={14} />
        <span>{label}</span>
      </button>
    {/each}
  </nav>

  <!-- Scrollable file area -->
  <div class="flex-1 overflow-y-auto">

    <!-- PROJECT section -->
    <div class="border-b app-border">
      <button
        class="w-full flex items-center gap-1 px-3 py-1.5 text-xs font-semibold uppercase tracking-wider app-dim hover:app-text transition-colors"
        on:click={() => projectOpen = !projectOpen}
      >
        <svelte:component this={projectOpen ? ChevronDown : ChevronRight} size={10} />
        <span>Project</span>
      </button>

      {#if projectOpen}
        <div class="px-2 pb-2 space-y-1">
          <button class="btn-primary w-full justify-center text-xs" on:click={openProject}>
            <svelte:component this={FolderOpen} size={12} />
            Open Project
          </button>
          {#if $project}
            <div class="px-2 py-1.5 app-card rounded text-xs">
              <div class="app-text font-medium truncate">{$project.name}</div>
              <div class="app-dim truncate">{$project.root.split("/").pop()}</div>
            </div>
          {/if}
        </div>

        {#if $sourceFiles.length > 0}
          <div class="pb-2 px-2 space-y-1">
            {#each projectGroups as [ext, files]}
              <div>
                <div class="app-dim text-xs uppercase tracking-wider px-2 py-0.5">{ext}</div>
                {#each files as f}
                  <button
                    class="sidebar-item w-full text-left truncate"
                    on:click={() => openTranspile(f)}
                    title={f.rel}
                  >
                    <svelte:component this={fileIcon(f.ext)} size={12} class="shrink-0" />
                    <span class="truncate">{f.rel.split("/").pop()}</span>
                  </button>
                {/each}
              </div>
            {/each}
          </div>
        {:else}
          <p class="app-dim text-xs px-4 pb-2">Open a project to see files.</p>
        {/if}

        <!-- Recent projects -->
        {#if $recentProjects.length > 0}
          <div class="border-t app-border">
            <button
              class="w-full flex items-center gap-1 px-3 py-1 text-xs app-dim hover:app-text transition-colors"
              on:click={() => recentsOpen = !recentsOpen}
            >
              <svelte:component this={recentsOpen ? ChevronDown : ChevronRight} size={9} />
              <svelte:component this={Clock} size={10} />
              <span>Recent</span>
            </button>
            {#if recentsOpen}
              {#each $recentProjects as r}
                <button
                  class="sidebar-item w-full text-left truncate"
                  on:click={() => loadRecent(r.root)}
                  title={r.root}
                >
                  <svelte:component this={FolderOpen} size={11} class="shrink-0 app-dim" />
                  <span class="truncate">{r.name}</span>
                </button>
              {/each}
            {/if}
          </div>
        {/if}
      {/if}
    </div>

    <!-- REFERENCES section -->
    <div>
      <button
        class="w-full flex items-center gap-1 px-3 py-1.5 text-xs font-semibold uppercase tracking-wider app-dim hover:app-text transition-colors"
        on:click={() => refsOpen = !refsOpen}
      >
        <svelte:component this={refsOpen ? ChevronDown : ChevronRight} size={10} />
        <span>References</span>
      </button>

      {#if refsOpen}
        <div class="px-2 pb-2 space-y-1">
          <button class="btn w-full justify-center text-xs" on:click={addRefFolder}>
            + Add Folder
          </button>
        </div>

        {#each $refFolders as folder}
          <div class="border-t app-border">
            <div class="flex items-center px-2 py-0.5">
              <button
                class="flex items-center gap-1 flex-1 min-w-0 text-left text-xs app-dim hover:app-text transition-colors py-1"
                on:click={() => toggleRefFolder(folder.root)}
                title={folder.root}
              >
                <svelte:component this={folder.open ? ChevronDown : ChevronRight} size={9} class="shrink-0" />
                <span class="truncate font-medium">{folder.name}</span>
                <span class="app-muted shrink-0">({folder.files.length})</span>
              </button>
              <button
                class="app-muted hover:text-red-400 transition-colors text-xs px-1 shrink-0"
                on:click={() => removeRefFolder(folder.root)}
                title="Remove folder"
              ><X size={11} /></button>
            </div>

            {#if folder.open}
              {#if folder.files.length === 0}
                <p class="app-dim text-xs px-4 pb-2">No AG source files found.</p>
              {:else}
                <div class="pb-1 px-2 space-y-1">
                  {#each groupFiles(folder.files) as [ext, files]}
                    <div>
                      <div class="app-dim text-xs uppercase tracking-wider px-2 py-0.5">{ext}</div>
                      {#each files as f}
                        <button
                          class="sidebar-item w-full text-left truncate"
                          on:click={() => openTranspile(f)}
                          title={f.path}
                        >
                          <svelte:component this={fileIcon(f.ext)} size={12} class="shrink-0" />
                          <span class="truncate">{f.rel.split("/").pop()}</span>
                        </button>
                      {/each}
                    </div>
                  {/each}
                </div>
              {/if}
            {/if}
          </div>
        {/each}

        {#if $refFolders.length === 0}
          <p class="app-dim text-xs px-4 pb-2">Add a folder to browse test files.</p>
        {/if}
      {/if}
    </div>

  </div>
</aside>
