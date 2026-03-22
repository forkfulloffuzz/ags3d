<script>
  import GraphView from "./GraphView.svelte";

  export let result; // TranspileResult

  const stages = [
    { id: "source",    label: "Source" },
    { id: "tokens",    label: "Tokens" },
    { id: "astText",   label: "AST" },
    { id: "symbols",   label: "Symbols" },
    { id: "blocking",  label: "Blocking" },
    { id: "gdscript",  label: "GDScript" },
    { id: "emitView",  label: "Side-by-Side" },
    { id: "sourceMap", label: "Source Map" },
  ];

  let active = "source";
  let showGraph = false;

  $: graphStages = { astText: result?.astDot, symbols: result?.symDot };
  $: canGraph = active === "astText" || active === "symbols";
  $: dotSrc = graphStages[active] ?? "";
  $: if (!canGraph) showGraph = false;

  // Highlighted source line linked from GDScript tab
  let highlightLine = null;
  let gdHighlightLine = null;

  function buildSourceMap(sm) {
    if (!sm) return { gdToAg: {}, agToGd: {} };
    const gdToAg = {}, agToGd = {};
    for (const [gd, , ag] of sm) {
      gdToAg[gd] = ag;
      agToGd[ag] = gd;
    }
    return { gdToAg, agToGd };
  }

  $: smaps = buildSourceMap(result?.sourceMap);

  function onGdLineClick(lineNum) {
    gdHighlightLine = lineNum;
    highlightLine = smaps.gdToAg[lineNum] ?? null;
  }
  function onSrcLineClick(lineNum) {
    highlightLine = lineNum;
    gdHighlightLine = smaps.agToGd[lineNum] ?? null;
  }

  function renderLinked(text, onLineClick, highlight) {
    if (!text) return [];
    return text.split("\n").map((line, i) => ({
      num: i + 1,
      text: line,
      active: highlight === i + 1,
    }));
  }

  $: srcLines = renderLinked(result?.source,  onSrcLineClick, highlightLine);
  $: gdLines  = renderLinked(result?.gdscript, onGdLineClick,  gdHighlightLine);
</script>

<div class="h-full flex flex-col">

  <!-- Stage sub-tabs -->
  <div class="flex items-end border-b border-gray-800 bg-gray-950 px-2 overflow-x-auto shrink-0">
    {#each stages as s}
      <button
        class={active === s.id ? "tab-btn-active" : "tab-btn"}
        on:click={() => active = s.id}
      >{s.label}</button>
    {/each}

    {#if canGraph}
      <div class="ml-auto flex items-center gap-1 pb-1 pr-1">
        <button
          class="px-2 py-0.5 text-xs rounded border transition-colors
                 {showGraph ? 'bg-violet-700 border-violet-600 text-white' : 'bg-gray-800 border-gray-600 text-gray-400 hover:text-gray-200'}"
          on:click={() => showGraph = !showGraph}
        >{showGraph ? "Text" : "Graph"}</button>
      </div>
    {/if}
  </div>

  <!-- Errors banner -->
  {#if result?.errors?.length}
    <div class="px-4 py-2 bg-red-950 border-b border-red-800 text-red-300 text-xs font-mono shrink-0">
      {#each result.errors as e}<div>⚠ {e}</div>{/each}
    </div>
  {/if}

  <!-- Content -->
  <div class="flex-1 min-h-0 overflow-hidden">

    {#if active === "source"}
      <div class="h-full overflow-auto">
        <table class="w-full text-xs font-mono">
          {#each srcLines as line}
            <tr
              class="cursor-pointer hover:bg-gray-800 {line.active ? 'bg-violet-900/40' : ''}"
              on:click={() => onSrcLineClick(line.num)}
            >
              <td class="text-gray-600 select-none w-10 text-right pr-3 pl-2 py-px border-r border-gray-800">{line.num}</td>
              <td class="pl-3 py-px text-gray-300 whitespace-pre">{line.text}</td>
            </tr>
          {/each}
        </table>
      </div>

    {:else if active === "gdscript"}
      <div class="h-full overflow-auto">
        <table class="w-full text-xs font-mono">
          {#each gdLines as line}
            {@const mappedAg = smaps.gdToAg[line.num]}
            <tr
              class="cursor-pointer hover:bg-gray-800 {line.active ? 'bg-violet-900/40' : ''}"
              on:click={() => onGdLineClick(line.num)}
              title={mappedAg ? `from .agscript line ${mappedAg}` : ""}
            >
              <td class="text-gray-600 select-none w-10 text-right pr-3 pl-2 py-px border-r border-gray-800">{line.num}</td>
              <td class="pl-3 py-px text-gray-300 whitespace-pre">{line.text}</td>
              {#if mappedAg}
                <td class="text-gray-600 text-right pr-2 text-xs">← {mappedAg}</td>
              {/if}
            </tr>
          {/each}
        </table>
      </div>

    {:else if active === "sourceMap"}
      <pre class="h-full overflow-auto p-4 text-xs text-gray-300 mono">{JSON.stringify(result?.sourceMap, null, 2)}</pre>

    {:else if (active === "astText" || active === "symbols") && showGraph}
      <GraphView dot={dotSrc} />

    {:else}
      <pre class="h-full overflow-auto p-4 text-xs text-gray-300 mono leading-5">{result?.[active] ?? ""}</pre>
    {/if}

  </div>
</div>
