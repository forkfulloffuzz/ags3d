<script>
  import GraphView from "./GraphView.svelte";
  import {
    highlightAGScript, highlightGDScript,
    colorizeTokens, colorizeAST, colorizeSymbols, colorizeBlocking, colorizeEmit,
  } from "../lib/highlight.js";
  import {
    FileText, Hash, GitBranch, List, Pause, Code2, Columns2, Map, X,
  } from "lucide-svelte";

  export let result; // TranspileResult

  const stages = [
    { id: "source",    label: "Source",     Icon: FileText  },
    { id: "tokens",    label: "Tokens",     Icon: Hash      },
    { id: "astText",   label: "AST",        Icon: GitBranch },
    { id: "symbols",   label: "Symbols",    Icon: List      },
    { id: "blocking",  label: "Blocking",   Icon: Pause     },
    { id: "gdscript",  label: "GDScript",   Icon: Code2     },
    { id: "emitView",  label: "Side-by-Side",Icon: Columns2 },
    { id: "sourceMap", label: "Source Map", Icon: Map       },
  ];

  let active = "source";
  let showGraph = false;

  $: graphStages = { astText: result?.astDot, symbols: result?.symDot };
  $: canGraph = active === "astText" || active === "symbols";
  $: dotSrc = graphStages[active] ?? "";
  $: if (!canGraph) showGraph = false;

  let highlightLine   = null;
  let gdHighlightLine = null;

  function buildSourceMap(sm) {
    if (!sm) return { gdToAg: {}, agToGd: {} };
    const gdToAg = {}, agToGd = {};
    for (const [gd, , ag] of sm) { gdToAg[gd] = ag; agToGd[ag] = gd; }
    return { gdToAg, agToGd };
  }

  $: smaps = buildSourceMap(result?.sourceMap);

  function onGdLineClick(n)  { gdHighlightLine = n; highlightLine   = smaps.gdToAg[n] ?? null; }
  function onSrcLineClick(n) { highlightLine   = n; gdHighlightLine = smaps.agToGd[n] ?? null; }

  function makeLines(text) {
    if (!text) return [];
    return text.split("\n").map((t, i) => ({ num: i+1, text: t }));
  }

  $: srcLines = makeLines(result?.source);
  $: gdLines  = makeLines(result?.gdscript);
  $: coloredTokens   = colorizeTokens(result?.tokens   ?? '');
  $: coloredAST      = colorizeAST(result?.astText     ?? '');
  $: coloredSymbols  = colorizeSymbols(result?.symbols ?? '');
  $: coloredBlocking = colorizeBlocking(result?.blocking ?? '');
  $: coloredEmit     = colorizeEmit(result?.emitView   ?? '');
</script>

<div class="h-full flex flex-col">

  <!-- Stage sub-tabs -->
  <div class="flex items-end border-b app-border app-bg px-2 overflow-x-auto shrink-0">
    {#each stages as s}
      <button
        class={active === s.id ? "tab-btn-active" : "tab-btn"}
        on:click={() => active = s.id}
      >
        <svelte:component this={s.Icon} size={11} />
        {s.label}
      </button>
    {/each}

    {#if canGraph}
      <div class="ml-auto flex items-center gap-1 pb-1 pr-1">
        <button
          class="px-2 py-0.5 text-xs rounded border transition-colors
                 {showGraph ? 'bg-violet-700 border-violet-600 text-white' : 'app-card border-app text-app-dim hover:app-text'}"
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
              class="cursor-pointer hover:bg-violet-900/20 {highlightLine === line.num ? 'bg-violet-900/40' : ''}"
              on:click={() => onSrcLineClick(line.num)}
            >
              <td class="app-dim select-none w-10 text-right pr-3 pl-2 py-px border-r app-border">{line.num}</td>
              <td class="pl-3 py-px app-code whitespace-pre">{@html highlightAGScript(line.text)}</td>
            </tr>
          {/each}
        </table>
      </div>

    {:else if active === "tokens"}
      <pre class="h-full overflow-auto p-4 text-xs app-code leading-5">{@html coloredTokens}</pre>

    {:else if active === "gdscript"}
      <div class="h-full overflow-auto">
        <table class="w-full text-xs font-mono">
          {#each gdLines as line}
            {@const mappedAg = smaps.gdToAg[line.num]}
            <tr
              class="cursor-pointer hover:bg-violet-900/20 {gdHighlightLine === line.num ? 'bg-violet-900/40' : ''}"
              on:click={() => onGdLineClick(line.num)}
              title={mappedAg ? `from .agscript line ${mappedAg}` : ""}
            >
              <td class="app-dim select-none w-10 text-right pr-3 pl-2 py-px border-r app-border">{line.num}</td>
              <td class="pl-3 py-px app-code whitespace-pre">{@html highlightGDScript(line.text)}</td>
              {#if mappedAg}
                <td class="app-dim text-right pr-2 text-xs">← {mappedAg}</td>
              {/if}
            </tr>
          {/each}
        </table>
      </div>

    {:else if active === "sourceMap"}
      <pre class="h-full overflow-auto p-4 text-xs app-code">{JSON.stringify(result?.sourceMap, null, 2)}</pre>

    {:else if (active === "astText" || active === "symbols") && showGraph}
      <GraphView dot={dotSrc} />

    {:else if active === "astText"}
      <pre class="h-full overflow-auto p-4 text-xs app-code leading-5">{@html coloredAST}</pre>

    {:else if active === "symbols"}
      <pre class="h-full overflow-auto p-4 text-xs app-code leading-5">{@html coloredSymbols}</pre>

    {:else if active === "blocking"}
      <pre class="h-full overflow-auto p-4 text-xs app-code leading-5">{@html coloredBlocking}</pre>

    {:else if active === "emitView"}
      <pre class="h-full overflow-auto p-4 text-xs app-code leading-5">{@html coloredEmit}</pre>

    {:else}
      <pre class="h-full overflow-auto p-4 text-xs app-code leading-5">{result?.[active] ?? ""}</pre>
    {/if}

  </div>
</div>
