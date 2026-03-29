<script>
  import { sourceFiles, project } from "../store.js";
  import {
    ParseRoom, ParseChar,
    GenerateRoomScene, GenerateCharScene,
    ValidateProject,
  } from "../../wailsjs/go/main/App.js";
  import { MapPin, User, CheckCircle, ChevronDown, ChevronRight } from "lucide-svelte";

  // Selected file
  let selectedPath = "";
  let selectedExt  = "";

  // Results
  let parsedRoom   = null;
  let parsedChar   = null;
  let sceneText    = "";
  let validateResult = null;

  // UI state
  let loading      = false;
  let activeTab    = "parsed"; // "parsed" | "scene" | "validate"
  let expandedSections = {};

  $: roomFiles = $sourceFiles.filter(f => f.ext === ".agroom");
  $: charFiles = $sourceFiles.filter(f => f.ext === ".agchar");

  async function selectFile(f) {
    selectedPath = f.path;
    selectedExt  = f.ext;
    parsedRoom   = null;
    parsedChar   = null;
    sceneText    = "";
    loading      = true;
    try {
      if (f.ext === ".agroom") {
        [parsedRoom, sceneText] = await Promise.all([
          ParseRoom(f.path),
          GenerateRoomScene(f.path),
        ]);
      } else if (f.ext === ".agchar") {
        [parsedChar, sceneText] = await Promise.all([
          ParseChar(f.path),
          GenerateCharScene(f.path),
        ]);
      }
    } finally {
      loading = false;
    }
  }

  async function runValidate() {
    validateResult = null;
    loading = true;
    try {
      validateResult = await ValidateProject();
    } finally {
      loading = false;
    }
    activeTab = "validate";
  }

  function toggle(key) {
    expandedSections[key] = !expandedSections[key];
    expandedSections = expandedSections;
  }

  function fmtVec3(v) { return `(${v.x}, ${v.y}, ${v.z})`; }
  function fmtVec2(v) { return `(${v.x}, ${v.z})`; }
</script>

<div class="h-full flex min-h-0">

  <!-- File list sidebar -->
  <div class="w-48 shrink-0 border-r app-border flex flex-col overflow-hidden">
    <div class="px-3 py-2 text-xs font-semibold uppercase tracking-wider app-dim border-b app-border">
      Rooms & Chars
    </div>
    <div class="flex-1 overflow-y-auto">

      {#if roomFiles.length > 0}
        <div class="px-2 pt-2">
          <div class="app-dim text-xs uppercase tracking-wider px-1 pb-1">.agroom</div>
          {#each roomFiles as f}
            <button
              class={selectedPath === f.path ? "sidebar-item-active w-full text-left truncate" : "sidebar-item w-full text-left truncate"}
              on:click={() => selectFile(f)}
              title={f.rel}
            >
              <MapPin size={12} class="shrink-0" />
              <span class="truncate">{f.rel.split("/").pop()}</span>
            </button>
          {/each}
        </div>
      {/if}

      {#if charFiles.length > 0}
        <div class="px-2 pt-2">
          <div class="app-dim text-xs uppercase tracking-wider px-1 pb-1">.agchar</div>
          {#each charFiles as f}
            <button
              class={selectedPath === f.path ? "sidebar-item-active w-full text-left truncate" : "sidebar-item w-full text-left truncate"}
              on:click={() => selectFile(f)}
              title={f.rel}
            >
              <User size={12} class="shrink-0" />
              <span class="truncate">{f.rel.split("/").pop()}</span>
            </button>
          {/each}
        </div>
      {/if}

      {#if roomFiles.length === 0 && charFiles.length === 0}
        <p class="app-dim text-xs px-3 py-3">Open a project to see rooms and characters.</p>
      {/if}

    </div>

    <!-- Validate button -->
    {#if $project}
      <div class="p-2 border-t app-border">
        <button class="btn-primary w-full justify-center text-xs" on:click={runValidate}>
          <CheckCircle size={12} />
          Validate Project
        </button>
      </div>
    {/if}
  </div>

  <!-- Main content -->
  <div class="flex-1 flex flex-col min-w-0 min-h-0">

    {#if !selectedPath && !validateResult}
      <div class="h-full flex items-center justify-center app-muted text-sm">
        Select a room or character file, or run Validate Project.
      </div>

    {:else}
      <!-- Tab bar -->
      <div class="flex items-end border-b app-border app-panel px-2 shrink-0">
        {#if selectedPath}
          <button
            class={activeTab === "parsed" ? "tab-btn-active" : "tab-btn"}
            on:click={() => activeTab = "parsed"}
          >Parsed</button>
          <button
            class={activeTab === "scene" ? "tab-btn-active" : "tab-btn"}
            on:click={() => activeTab = "scene"}
          >Generated .tscn</button>
        {/if}
        {#if validateResult !== null}
          <button
            class={activeTab === "validate" ? "tab-btn-active" : "tab-btn"}
            on:click={() => activeTab = "validate"}
          >Validate
            {#if validateResult.issues?.length}
              <span class="ml-1 text-red-400">({validateResult.issues.length})</span>
            {:else}
              <span class="ml-1 text-green-400">✓</span>
            {/if}
          </button>
        {/if}
      </div>

      <!-- Tab content -->
      <div class="flex-1 min-h-0 overflow-auto p-3 font-mono text-xs">

        {#if loading}
          <div class="h-full flex items-center justify-center app-muted animate-pulse">Loading…</div>

        {:else if activeTab === "parsed"}

          <!-- Room parsed data -->
          {#if parsedRoom}
            {#if parsedRoom.error}
              <div class="text-red-400">{parsedRoom.error}</div>
            {:else}
              <div class="space-y-3">
                <div class="app-card rounded p-2">
                  <span class="app-dim">name</span> <span class="text-violet-400">{parsedRoom.name}</span>
                  {#if parsedRoom.initialCamera}
                    &nbsp;&nbsp;<span class="app-dim">initial_camera</span> <span class="text-amber-400">"{parsedRoom.initialCamera}"</span>
                  {/if}
                </div>

                {#each [
                  ["Cameras", parsedRoom.cameras],
                  ["Points", parsedRoom.points],
                  ["WalkableSurfaces", parsedRoom.walkableSurfaces],
                  ["BlockerVolumes", parsedRoom.blockerVolumes],
                  ["SpawnPoints", parsedRoom.spawnPoints],
                  ["Hotspots", parsedRoom.hotspots],
                ] as [label, items]}
                  {#if items?.length}
                    <div class="app-card rounded overflow-hidden">
                      <button
                        class="w-full flex items-center gap-1 px-2 py-1.5 text-xs font-semibold app-dim hover:app-text transition-colors"
                        on:click={() => toggle(label)}
                      >
                        <svelte:component this={expandedSections[label] === false ? ChevronRight : ChevronDown} size={10} />
                        {label} <span class="ml-1 text-violet-400/60">({items.length})</span>
                      </button>
                      {#if expandedSections[label] !== false}
                        <div class="divide-y app-border">
                          {#each items as item}
                            <div class="px-3 py-1.5 space-x-3">
                              <span class="text-green-400">"{item.name}"</span>
                              {#if item.position}
                                <span class="app-dim">pos</span> <span>{fmtVec3(item.position)}</span>
                              {/if}
                              {#if item.lookAt}
                                <span class="app-dim">look_at</span> <span>{fmtVec3(item.lookAt)}</span>
                              {/if}
                              {#if item.size && item.size.z !== undefined && item.size.y === undefined}
                                <span class="app-dim">size</span> <span>{fmtVec2(item.size)}</span>
                              {:else if item.size}
                                <span class="app-dim">size</span> <span>{fmtVec3(item.size)}</span>
                              {/if}
                              {#if item.character}
                                <span class="app-dim">character</span> <span class="text-amber-400">"{item.character}"</span>
                              {/if}
                            </div>
                          {/each}
                        </div>
                      {/if}
                    </div>
                  {/if}
                {/each}
              </div>
            {/if}
          {/if}

          <!-- Char parsed data -->
          {#if parsedChar}
            {#if parsedChar.error}
              <div class="text-red-400">{parsedChar.error}</div>
            {:else}
              <div class="space-y-2">
                <div class="app-card rounded p-2 space-y-1">
                  {#each [
                    ["name", parsedChar.name],
                    ["display_name", parsedChar.displayName],
                    ["type", parsedChar.type],
                    ["mesh", parsedChar.mesh],
                    ["sprite_sheet", parsedChar.spriteSheet],
                  ] as [key, val]}
                    {#if val}
                      <div><span class="app-dim">{key}</span> <span class="text-violet-400">"{val}"</span></div>
                    {/if}
                  {/each}
                  {#if parsedChar.spriteAngles}
                    <div>
                      <span class="app-dim">sprite_angles</span> <span>{parsedChar.spriteAngles}</span>
                      &nbsp;<span class="app-dim">frames_per_angle</span> <span>{parsedChar.framesPerAngle}</span>
                    </div>
                  {/if}
                </div>
                {#if parsedChar.animations && Object.keys(parsedChar.animations).length}
                  <div class="app-card rounded overflow-hidden">
                    <div class="px-2 py-1 text-xs font-semibold app-dim">animations</div>
                    <div class="divide-y app-border">
                      {#each Object.entries(parsedChar.animations) as [k, v]}
                        <div class="px-3 py-1">
                          <span class="app-dim">{k}</span> = <span class="text-amber-400">"{v}"</span>
                        </div>
                      {/each}
                    </div>
                  </div>
                {/if}
              </div>
            {/if}
          {/if}

        {:else if activeTab === "scene"}
          <pre class="whitespace-pre app-text leading-relaxed">{sceneText}</pre>

        {:else if activeTab === "validate"}
          {#if validateResult?.error}
            <div class="text-red-400">{validateResult.error}</div>
          {:else if !validateResult?.issues?.length}
            <div class="text-green-400 flex items-center gap-2">
              <CheckCircle size={14} /> No issues found — project is clean.
            </div>
          {:else}
            <div class="space-y-1">
              {#each validateResult.issues as iss}
                <div class="app-card rounded px-3 py-1.5 flex items-start gap-2">
                  <span class={iss.severity === "error" ? "text-red-400 shrink-0" : "text-amber-400 shrink-0"}>
                    {iss.severity}
                  </span>
                  <span class="text-violet-400/70 shrink-0">{iss.file}</span>
                  <span class="app-text">{iss.message}</span>
                </div>
              {/each}
            </div>
          {/if}
        {/if}

      </div>
    {/if}
  </div>

</div>
