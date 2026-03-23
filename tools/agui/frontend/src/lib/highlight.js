// Regex-based syntax highlighter — returns an HTML string with <span> tags.
// All non-span text is HTML-escaped, so it is safe to render with {@html}.

function esc(s) {
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

const AG_KW = new Set([
  'function','if','else','elif','while','for','return',
  'true','false','null','and','or','not','break','continue',
]);

const GD_KW = new Set([
  'func','var','const','if','else','elif','while','for','return',
  'true','false','null','pass','extends','class','class_name',
  'enum','static','export','onready','signal','match',
  'and','or','not','break','continue','in','is','as','self','super',
]);

function tokenize(line, kws) {
  const out = [];
  let i = 0;
  while (i < line.length) {
    // single-line comment
    if (line[i] === '/' && line[i+1] === '/') {
      out.push(`<span class="syn-cmt">${esc(line.slice(i))}</span>`);
      break;
    }
    // string
    if (line[i] === '"' || line[i] === "'") {
      const q = line[i]; let j = i+1;
      while (j < line.length && line[j] !== q) { if (line[j]==='\\') j++; j++; }
      out.push(`<span class="syn-str">${esc(line.slice(i, j+1))}</span>`);
      i = j+1; continue;
    }
    // number
    if (/[0-9]/.test(line[i])) {
      let j = i;
      while (j < line.length && /[0-9.]/.test(line[j])) j++;
      out.push(`<span class="syn-num">${esc(line.slice(i, j))}</span>`);
      i = j; continue;
    }
    // identifier / keyword
    if (/[a-zA-Z_]/.test(line[i])) {
      let j = i;
      while (j < line.length && /[a-zA-Z0-9_]/.test(line[j])) j++;
      const w = line.slice(i, j);
      out.push(kws.has(w) ? `<span class="syn-kw">${esc(w)}</span>` : esc(w));
      i = j; continue;
    }
    out.push(esc(line[i++]));
  }
  return out.join('');
}

// Token-kind categories for the Tokens tab colourisation
const KIND_CLASS = {
  // keywords
  FUNCTION:'syn-kw', IF:'syn-kw', ELSE:'syn-kw', WHILE:'syn-kw',
  FOR:'syn-kw', RETURN:'syn-kw', AND:'syn-kw', OR:'syn-kw', NOT:'syn-kw',
  BREAK:'syn-kw', CONTINUE:'syn-kw',
  // literals
  INTEGER:'syn-num', FLOAT:'syn-num', STRING:'syn-str',
  TRUE:'syn-num', FALSE:'syn-num', NULL:'syn-num',
  // identifier
  IDENT:'syn-ident',
  // delimiters / operators  — leave unstyled (just normal text)
};

export function highlightAGScript(line) { return tokenize(line, AG_KW); }
export function highlightGDScript(line) { return tokenize(line, GD_KW); }

// ---------------------------------------------------------------------------
// Rule-based colorizer for multi-line pre-formatted stage outputs.
// Rules are applied in order; earliest match wins (no overlaps).
// ---------------------------------------------------------------------------

function applyRules(text, rules) {
  const matches = [];
  for (const { re, cls } of rules) {
    const r = new RegExp(re.source, re.flags.replace('g','') + 'g');
    let m;
    while ((m = r.exec(text)) !== null) {
      matches.push({ start: m.index, end: m.index + m[0].length, raw: m[0], cls });
    }
  }
  // Earliest start wins; ties go to longer match.
  matches.sort((a, b) => a.start !== b.start ? a.start - b.start : b.end - a.end);
  const used = [];
  let lastEnd = 0;
  for (const m of matches) {
    if (m.start >= lastEnd) { used.push(m); lastEnd = m.end; }
  }
  const parts = [];
  let pos = 0;
  for (const m of used) {
    if (m.start > pos) parts.push(esc(text.slice(pos, m.start)));
    parts.push(`<span class="${m.cls}">${esc(m.raw)}</span>`);
    pos = m.end;
  }
  if (pos < text.length) parts.push(esc(text.slice(pos)));
  return parts.join('');
}

// AST text (viz ast)
const AST_NODES = [
  'FunctionDecl','NamespaceDecl','EnumDecl','EnumMember','TopVarDecl',
  'Block','IfStmt','WhileStmt','ForStmt','ReturnStmt','ExprStmt',
  'VarDecl','AssignStmt','BinaryExpr','UnaryExpr','CallExpr',
  'IndexExpr','MemberExpr','Literal','Ident','SwitchStmt',
  'BreakStmt','ContinueStmt','SwitchCase',
];
const AST_RULES = [
  { re: /^AST — .+$/m,                              cls: 'app-dim'  },
  { re: /\[\d+:\d+\]/g,                             cls: 'syn-num'  },
  { re: /"[^"]*"/g,                                 cls: 'syn-str'  },
  { re: new RegExp(`\\b(${AST_NODES.join('|')})\\b`,'g'), cls: 'syn-kw' },
  { re: /→/g,                                       cls: 'syn-op'   },
  { re: /[└├│─]+/g,                                 cls: 'app-muted'},
];
export function colorizeAST(text) { return applyRules(text || '', AST_RULES); }

// Symbols table (viz symbols)
const SYM_RULES = [
  { re: /^Symbols — .+$/m,              cls: 'app-dim' },
  { re: /\[\d+:\d+\]/g,                 cls: 'syn-num' },
  { re: /\[export\]|\[blocking\]/g,     cls: 'syn-kw'  },
  { re: /→ \w+/g,                       cls: 'syn-num' },
  { re: /\bfunction\b/g,                cls: 'syn-kw'  },
];
export function colorizeSymbols(text) { return applyRules(text || '', SYM_RULES); }

// Blocking annotations (viz blocking)
const BLK_RULES = [
  { re: /^Blocking — .+$/m,             cls: 'app-dim'  },
  { re: /^Functions \(\d+\)$/m,         cls: 'app-dim'  },
  { re: /\[blocking\]/g,                cls: 'syn-kw'   },
  { re: /\[-\]/g,                       cls: 'app-muted'},
  { re: /\d+ blocking call\(s\)/g,      cls: 'syn-num'  },
  { re: /L\d+/g,                        cls: 'syn-num'  },
];
export function colorizeBlocking(text) { return applyRules(text || '', BLK_RULES); }

// Side-by-side emit view (viz emit)
// Format per line:  "  <pad(agText,48)> │  <gdText>"
// We colorize structural elements; code content uses the per-column tokenizers.
export function colorizeEmit(text) {
  if (!text) return '';
  return text.split('\n').map(line => {
    // Header / separator lines
    if (/^Transpile — /.test(line)) return `<span class="app-dim">${esc(line)}</span>`;
    if (/^[ ─]+[┼─]/.test(line))   return `<span class="app-muted">${esc(line)}</span>`;

    // Column header line: "  AGS-spirit ... │  GDScript"
    const headerM = line.match(/^(\s+)(AGS-spirit\s*)(│)(\s+GDScript\s*)$/);
    if (headerM) {
      return esc(headerM[1])
        + `<span class="syn-kw">${esc(headerM[2])}</span>`
        + `<span class="app-muted">${esc(headerM[3])}</span>`
        + `<span class="syn-kw">${esc(headerM[4])}</span>`;
    }

    // Data lines: "  <agCol> │  <gdCol>"
    // Split at the fixed " │  " separator
    const sepIdx = line.indexOf(' │  ');
    if (sepIdx === -1) return esc(line);

    const agPart = line.slice(0, sepIdx);
    const gdPart = line.slice(sepIdx + 4);

    // agPart: "  <lineNum>│ <code>" or "  ~│"
    const agM = agPart.match(/^(\s*)(\d+)(│)(.*)$|^(\s*~)(│)(.*)$/);
    let agHtml;
    if (agM) {
      if (agM[2] !== undefined) {
        // has line number
        agHtml = esc(agM[1])
          + `<span class="app-dim">${esc(agM[2])}</span>`
          + `<span class="app-muted">${esc(agM[3])}</span>`
          + tokenize(agM[4], AG_KW);
      } else {
        // ~│ unmapped
        agHtml = `<span class="app-muted">${esc(agM[5] + agM[6])}</span>`
          + esc(agM[7]);
      }
    } else {
      agHtml = esc(agPart);
    }

    // gdPart: "  <lineNum>│ <code>"  (%3d produces leading spaces)
    const gdM = gdPart.match(/^(\s*)(\d+)(│)(.*)$/);
    let gdHtml;
    if (gdM) {
      gdHtml = esc(gdM[1])
        + `<span class="app-dim">${esc(gdM[2])}</span>`
        + `<span class="app-muted">${esc(gdM[3])}</span>`
        + tokenize(gdM[4], GD_KW);
    } else {
      gdHtml = esc(gdPart);
    }

    return agHtml
      + `<span class="app-muted"> │  </span>`
      + gdHtml;
  }).join('\n');
}

export function colorizeTokens(text) {
  if (!text) return '';
  return text.split('\n').map(line => {
    // fixed-width format:  "     1    1  FUNCTION            "function""
    const m = line.match(/^(\s*\d+\s+\d+\s{2})([A-Z_]+)(.*)$/);
    if (!m) return esc(line);
    const cls = KIND_CLASS[m[2]];
    const kind = cls ? `<span class="${cls}">${esc(m[2])}</span>` : esc(m[2]);
    return esc(m[1]) + kind + esc(m[3]);
  }).join('\n');
}
