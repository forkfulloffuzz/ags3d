import { writable } from "svelte/store";

// Persisted store — value is kept in localStorage
function local(key, init) {
  let stored;
  try { stored = JSON.parse(localStorage.getItem(key)); } catch {}
  const s = writable(stored ?? init);
  s.subscribe(v => { try { localStorage.setItem(key, JSON.stringify(v)); } catch {} });
  return s;
}

// Active top-level panel: "build" | "viz" | "transpile" | "batch" | "roomchar"
export const activePanel = writable("build");

// Currently open project
export const project = writable(null); // ProjectInfo | null

// Source files in the open project
export const sourceFiles = writable([]);

// Reference folders: [{root, name, files: SourceFile[], open: bool}]
export const refFolders = writable([]);

// Log lines: [{kind: "info"|"error"|"warn"|"ok", msg: string}]
export const logLines = writable([]);

// Transpile tabs: [{id, label, path, result}]
export const transpileTabs = writable([]);
export const activeTranspileTab = writable(null);

// Viz state
export const vizFile = writable(null);
export const vizStage = writable("tokens");
export const vizMode = writable("text"); // "text" | "graph"

// Theme: "dark" | "light" — persisted
export const theme = local("ag-theme", "dark");

// Recent projects: [{root, name}] — last 5, persisted
export const recentProjects = local("ag-recent", []);

export function addLog(kind, msg) {
  logLines.update(lines => [...lines.slice(-2000), { kind, msg }]);
}

export function clearLog() {
  logLines.set([]);
}

export function addRecentProject({ root, name }) {
  recentProjects.update(list =>
    [{ root, name }, ...list.filter(p => p.root !== root)].slice(0, 5)
  );
}

export function closeTab(id) {
  let next = null;
  transpileTabs.update(tabs => {
    const idx = tabs.findIndex(t => t.id === id);
    const rest = tabs.filter(t => t.id !== id);
    if (rest.length > 0) next = rest[Math.max(0, idx - 1)].id;
    return rest;
  });
  activeTranspileTab.set(next);
}
