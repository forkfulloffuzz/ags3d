<script>
  import { onMount } from "svelte";
  import { EventsOn } from "../wailsjs/runtime/runtime.js";
  import { activePanel, project, addLog, clearLog } from "./store.js";
  import Sidebar from "./components/Sidebar.svelte";
  import BuildPanel from "./components/BuildPanel.svelte";
  import VizPanel from "./components/VizPanel.svelte";
  import TranspilePanel from "./components/TranspilePanel.svelte";
  import BatchPanel from "./components/BatchPanel.svelte";
  import LogPanel from "./components/LogPanel.svelte";

  let logHeight = 160;
  let draggingLog = false;
  let startY = 0;
  let startH = 0;

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
    return () => {
      window.removeEventListener("mousemove", onMove);
      window.removeEventListener("mouseup",  onUp);
    };
  });

  function startDragLog(e) { draggingLog = true; startY = e.clientY; startH = logHeight; }
</script>

<div class="flex flex-col h-full bg-gray-950">

  <!-- Header -->
  <header class="flex items-center gap-3 px-4 h-10 bg-gray-900 border-b border-gray-700 shrink-0 z-10">
    <span class="text-violet-400 font-bold tracking-wide">AG Studio</span>
    <span class="text-gray-600">·</span>
    <span class="text-gray-400 text-xs truncate">{$project?.name ?? "no project open"}</span>
  </header>

  <!-- Body -->
  <div class="flex flex-1 min-h-0">
    <Sidebar />

    <main class="flex-1 flex flex-col min-w-0">
      <div class="flex-1 min-h-0 overflow-hidden">
        {#if $activePanel === "build"}
          <BuildPanel />
        {:else if $activePanel === "viz"}
          <VizPanel />
        {:else if $activePanel === "transpile"}
          <TranspilePanel />
        {:else if $activePanel === "batch"}
          <BatchPanel />
        {/if}
      </div>

      <!-- Log resize handle -->
      <div
        class="h-1 bg-gray-800 hover:bg-violet-600 cursor-row-resize shrink-0 transition-colors"
        on:mousedown={startDragLog}
        role="separator"
      ></div>

      <LogPanel height={logHeight} />
    </main>
  </div>
</div>
