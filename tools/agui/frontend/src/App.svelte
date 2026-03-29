<script>
  import { onMount } from "svelte";
  import { EventsOn } from "../wailsjs/runtime/runtime.js";
  import { Build } from "../wailsjs/go/main/App.js";
  import {
    activePanel, project, addLog, clearLog,
    theme, transpileTabs, activeTranspileTab, closeTab,
  } from "./store.js";
  import { get } from "svelte/store";
  import Sidebar from "./components/Sidebar.svelte";
  import BuildPanel from "./components/BuildPanel.svelte";
  import VizPanel from "./components/VizPanel.svelte";
  import TranspilePanel from "./components/TranspilePanel.svelte";
  import BatchPanel from "./components/BatchPanel.svelte";
  import RoomCharPanel from "./components/RoomCharPanel.svelte";
  import LogPanel from "./components/LogPanel.svelte";
  import { Sun, Moon } from "lucide-svelte";

  let logHeight = 160;
  let draggingLog = false;
  let startY = 0;
  let startH = 0;

  function toggleTheme() {
    theme.update(t => t === "dark" ? "light" : "dark");
  }

  onMount(() => {
    EventsOn("log:info",  msg => addLog("info",  msg));
    EventsOn("log:error", msg => addLog("error", msg));
    EventsOn("log:warn",  msg => addLog("warn",  msg));
    EventsOn("log:ok",    msg => addLog("ok",    msg));
    EventsOn("log:clear", ()  => clearLog());

    const onMove = e => {
      if (!draggingLog) return;
      logHeight = Math.max(60, Math.min(600, startH + (startY - e.clientY)));
    };
    const onUp = () => { draggingLog = false; };
    window.addEventListener("mousemove", onMove);
    window.addEventListener("mouseup",  onUp);

    // Keyboard shortcuts
    const onKey = e => {
      if (e.ctrlKey && e.key === 'b') {
        e.preventDefault();
        if (get(project)) { addLog("info", "Building…"); Build(); }
      }
      if (e.ctrlKey && e.key === 'w') {
        e.preventDefault();
        const id = get(activeTranspileTab);
        if (id && get(activePanel) === 'transpile') closeTab(id);
      }
      if (e.ctrlKey && e.key === 't') {
        e.preventDefault();
        activePanel.set('transpile');
      }
      if (e.key === 'F5') {
        e.preventDefault();
        // Trigger a reactive refresh by toggling a dummy store value
        activePanel.update(p => p);
      }
    };
    window.addEventListener("keydown", onKey);

    return () => {
      window.removeEventListener("mousemove", onMove);
      window.removeEventListener("mouseup",  onUp);
      window.removeEventListener("keydown",  onKey);
    };
  });

  function startDragLog(e) { draggingLog = true; startY = e.clientY; startH = logHeight; }

  $: isDark = $theme === "dark";
</script>

<div class="flex flex-col h-full app-bg {isDark ? 'theme-dark' : 'theme-light'}">

  <!-- Header -->
  <header class="flex items-center gap-3 px-4 h-10 app-header border-b shrink-0 z-10">
    <span class="text-violet-400 font-bold tracking-wide text-sm">AG Studio</span>
    <span class="app-muted">·</span>
    <span class="app-dim text-xs truncate">{$project?.name ?? "no project open"}</span>
    <div class="ml-auto flex items-center gap-2">
      <span class="app-dim text-xs hidden sm:block">
        {#if $project}<span class="text-violet-400/60">{$project.root.split("/").pop()}</span>{/if}
      </span>
      <button
        class="p-1.5 rounded app-dim hover:app-text transition-colors"
        on:click={toggleTheme}
        title={isDark ? "Switch to light theme" : "Switch to dark theme"}
      >
        {#if isDark}
          <Sun size={14} />
        {:else}
          <Moon size={14} />
        {/if}
      </button>
    </div>
  </header>

  <!-- Body -->
  <div class="flex flex-1 min-h-0">
    <Sidebar />

    <main class="flex-1 flex flex-col min-w-0 app-bg">
      <div class="flex-1 min-h-0 overflow-hidden">
        {#if $activePanel === "build"}
          <BuildPanel />
        {:else if $activePanel === "viz"}
          <VizPanel />
        {:else if $activePanel === "transpile"}
          <TranspilePanel />
        {:else if $activePanel === "batch"}
          <BatchPanel />
        {:else if $activePanel === "roomchar"}
          <RoomCharPanel />
        {/if}
      </div>

      <!-- Log resize handle -->
      <div
        class="h-1 app-resize-handle cursor-row-resize shrink-0 transition-colors"
        on:mousedown={startDragLog}
        role="separator"
      ></div>

      <LogPanel height={logHeight} />
    </main>
  </div>
</div>
