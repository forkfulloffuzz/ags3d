import { writable } from "svelte/store";

// Active top-level panel: "build" | "viz" | "transpile" | "batch"
export const activePanel = writable("build");

// Currently open project
export const project = writable(null); // ProjectInfo | null

// Source files in the open project
export const sourceFiles = writable([]);

// Log lines: [{kind: "info"|"error"|"warn"|"ok", msg: string}]
export const logLines = writable([]);

// Transpile tabs: [{id, label, path, result}]
export const transpileTabs = writable([]);
export const activeTranspileTab = writable(null);

// Viz state
export const vizFile = writable(null);
export const vizStage = writable("tokens");
export const vizMode = writable("text"); // "text" | "graph"

export function addLog(kind, msg) {
  logLines.update(lines => [...lines.slice(-2000), { kind, msg }]);
}

export function clearLog() {
  logLines.set([]);
}
