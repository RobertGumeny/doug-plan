import React, { useState, useEffect, useMemo } from 'react';
import ApproveButton from '../ApproveButton';

// Parse a YAML frontmatter block from the start of content.
// Returns an object of top-level keys, or null if no frontmatter block is found.
// Only handles flat key-value pairs and simple list values (- item).
function parseFrontmatter(content) {
  const match = content.match(/^---\n([\s\S]*?)\n---\n/);
  if (!match) return null;
  const result = {};
  let currentKey = null;
  let inList = false;
  for (const line of match[1].split('\n')) {
    const listItem = line.match(/^\s+-\s+(.*)/);
    if (inList && listItem) {
      result[currentKey].push(listItem[1].trim());
      continue;
    }
    inList = false;
    const kv = line.match(/^([a-zA-Z_][a-zA-Z0-9_]*):\s*(.*)/);
    if (kv) {
      currentKey = kv[1];
      const val = kv[2].trim();
      if (val === '') {
        result[currentKey] = [];
        inList = true;
      } else {
        result[currentKey] = val;
      }
    }
  }
  return result;
}

// Extract the project name from the "## Project Name" section of VISION.md.
function extractProjectName(content) {
  const withoutFm = content.replace(/^---\n[\s\S]*?\n---\n/, '');
  const match = withoutFm.match(/##\s+Project Name\s*\n+([^\n#][^\n]*)/);
  return match ? match[1].trim() : '';
}

// Validate required greenfield frontmatter fields.
// Returns an array of error strings; empty means valid.
function validateFrontmatter(fm) {
  if (!fm) return ['No YAML frontmatter block found. Add a frontmatter block with project_mode: greenfield and required scaffold fields.'];
  const missing = [];
  if (!fm.project_mode || !fm.project_mode.trim()) missing.push('project_mode');
  if (!fm.language || !fm.language.trim()) missing.push('language');
  if (!fm.runtime || !fm.runtime.trim()) missing.push('runtime');
  if (missing.length > 0) return [`Missing required fields: ${missing.join(', ')}`];
  return [];
}

// Derive a manifest.yaml string from validated frontmatter and project name.
function deriveManifest(fm, projectName) {
  const lines = [
    'schema_version: 1',
    'project:',
    `  name: "${projectName}"`,
    `  mode: "${fm.project_mode}"`,
    'scaffold:',
    `  language: "${fm.language}"`,
    `  runtime: "${fm.runtime}"`,
  ];
  if (fm.framework) lines.push(`  framework: "${fm.framework}"`);
  if (fm.package_manager) lines.push(`  package_manager: "${fm.package_manager}"`);
  if (fm.build_system) lines.push(`  build_system: "${fm.build_system}"`);

  const rtDeps = Array.isArray(fm.runtime_dependencies) ? fm.runtime_dependencies : [];
  const devDeps = Array.isArray(fm.dev_dependencies) ? fm.dev_dependencies : [];
  if (rtDeps.length > 0 || devDeps.length > 0) {
    lines.push('dependencies:');
    if (rtDeps.length > 0) {
      lines.push('  runtime:');
      rtDeps.forEach(d => lines.push(`    - "${d}"`));
    }
    if (devDeps.length > 0) {
      lines.push('  development:');
      devDeps.forEach(d => lines.push(`    - "${d}"`));
    }
  }

  const constraints = Array.isArray(fm.bootstrap_constraints) ? fm.bootstrap_constraints : [];
  if (constraints.length > 0) {
    lines.push('constraints:');
    constraints.forEach(c => lines.push(`  - "${c}"`));
  }

  return lines.join('\n') + '\n';
}

// VisionView renders the Discovery approval UI.
//
// Greenfield projects (project_mode: greenfield in VISION.md frontmatter) show a
// split view: VISION.md on the left and manifest.yaml on the right, matching the
// PRD split-view pattern. Both panes are editable and persisted on approve.
//
// If VISION.md frontmatter is invalid on load, an error is shown inline above the
// manifest pane and the manifest is re-derived client-side as the user edits.
// Non-greenfield projects show a single VISION.md pane.
export default function VisionView({ content, secondaryContent, onApprove, status }) {
  const [visionText, setVisionText] = useState(content);
  const [manifestText, setManifestText] = useState(secondaryContent || '');

  // Track whether the manifest pane has been populated (server pre-render or auto-derived).
  // Once seeded we stop auto-updating so user edits in the manifest pane are preserved.
  const [manifestSeeded, setManifestSeeded] = useState((secondaryContent || '').trim() !== '');

  const fm = useMemo(() => parseFrontmatter(visionText), [visionText]);
  const isGreenfield = fm != null && fm.project_mode === 'greenfield';
  const errors = useMemo(() => isGreenfield ? validateFrontmatter(fm) : [], [fm, isGreenfield]);

  const derivedManifest = useMemo(() => {
    if (!isGreenfield || errors.length > 0) return null;
    return deriveManifest(fm, extractProjectName(visionText));
  }, [fm, isGreenfield, errors, visionText]);

  // Auto-seed the manifest pane the first time a valid derivedManifest is available
  // and the pane has not yet been populated (server did not pre-render a draft).
  useEffect(() => {
    if (!manifestSeeded && derivedManifest) {
      setManifestText(derivedManifest);
      setManifestSeeded(true);
    }
  }, [manifestSeeded, derivedManifest]);

  // Non-greenfield: single pane, same as original behaviour.
  if (!isGreenfield) {
    return (
      <div>
        <textarea
          style={styles.singleTextarea}
          value={visionText}
          onChange={e => setVisionText(e.target.value)}
          spellCheck={false}
        />
        <ApproveButton onApprove={() => onApprove(visionText, '')} status={status} />
      </div>
    );
  }

  // Greenfield: split view.
  return (
    <div>
      <div style={styles.splitLayout}>
        <div style={styles.pane}>
          <div style={styles.paneLabel}>VISION.md</div>
          <textarea
            style={styles.textarea}
            value={visionText}
            onChange={e => setVisionText(e.target.value)}
            spellCheck={false}
          />
        </div>
        <div style={styles.pane}>
          <div style={styles.paneLabel}>manifest.yaml</div>
          {errors.length > 0 && (
            <div style={styles.errorBanner}>
              {errors.map((msg, i) => <div key={i}>{msg}</div>)}
            </div>
          )}
          <textarea
            style={{ ...styles.textarea, fontSize: 12 }}
            value={manifestText}
            onChange={e => setManifestText(e.target.value)}
            spellCheck={false}
          />
        </div>
      </div>
      <ApproveButton onApprove={() => onApprove(visionText, manifestText)} status={status} />
    </div>
  );
}

const styles = {
  splitLayout: {
    display: 'flex',
    gap: 12,
    marginBottom: 10,
  },
  pane: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
  },
  paneLabel: {
    fontFamily: 'monospace',
    fontSize: 12,
    color: '#555',
    marginBottom: 4,
    fontWeight: 'bold',
  },
  errorBanner: {
    background: '#fff3cd',
    border: '1px solid #ffc107',
    borderRadius: 3,
    padding: '6px 10px',
    marginBottom: 6,
    fontFamily: 'monospace',
    fontSize: 12,
    color: '#856404',
  },
  textarea: {
    flex: 1,
    width: '100%',
    height: 520,
    fontFamily: 'monospace',
    fontSize: 13,
    boxSizing: 'border-box',
    padding: 8,
    resize: 'vertical',
  },
  singleTextarea: {
    width: '100%',
    height: 480,
    fontFamily: 'monospace',
    fontSize: 14,
    boxSizing: 'border-box',
    padding: 8,
  },
};
