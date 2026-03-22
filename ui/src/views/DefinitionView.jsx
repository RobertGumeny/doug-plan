import React, { useState } from 'react';
import ApproveButton from '../ApproveButton';

// Parse YAML frontmatter block (---\nkey: "val"\n---\n) from the top of content.
// Returns { frontmatter: {id, name}, rest: string }.
function parseFrontmatter(markdown) {
  const fm = { id: '', name: '' };
  const match = markdown.match(/^---\n([\s\S]*?)\n---\n/);
  if (!match) return { frontmatter: fm, rest: markdown };
  for (const line of match[1].split('\n')) {
    const m = line.match(/^(\w+):\s*"?(.*?)"?\s*$/);
    if (m) fm[m[1]] = m[2];
  }
  return { frontmatter: fm, rest: markdown.slice(match[0].length) };
}

// Parse lines of a tasks block (everything after "## Tasks") into task objects.
function parseTaskLines(lines) {
  const tasks = [];
  let current = null;
  let inCriteria = false;

  for (const line of lines) {
    if (line.startsWith('### ')) {
      if (current) tasks.push(current);
      current = { heading: line.slice(4), type: '', description: '', criteria: [] };
      inCriteria = false;
      continue;
    }
    if (!current) continue;

    const typeMatch = line.match(/^\*\*Type\*\*:\s*(.*)/);
    if (typeMatch) { current.type = typeMatch[1].trim(); inCriteria = false; continue; }

    const descMatch = line.match(/^\*\*Description\*\*:\s*(.*)/);
    if (descMatch) { current.description = descMatch[1].trim(); inCriteria = false; continue; }

    if (line === '**Acceptance Criteria**:') { inCriteria = true; continue; }
    if (inCriteria && line.startsWith('- ')) { current.criteria.push(line.slice(2)); continue; }
    if (inCriteria && line.trim() !== '') inCriteria = false;
  }
  if (current) tasks.push(current);
  return tasks;
}

// Parse content (after frontmatter) into sections and tasks.
// Returns { sectionOrder: string[], sections: {name: string}, tasks: [], hasStructure: bool }
function parseContent(content) {
  const sectionOrder = [];
  const sections = {};
  const taskLines = [];
  let currentSection = null;
  let inTasks = false;

  for (const line of content.split('\n')) {
    if (line.startsWith('## Tasks')) {
      inTasks = true;
      currentSection = null;
      continue;
    }
    if (line.startsWith('## ')) {
      currentSection = line.slice(3).trim();
      sectionOrder.push(currentSection);
      sections[currentSection] = [];
      inTasks = false;
      continue;
    }
    if (inTasks) { taskLines.push(line); continue; }
    if (currentSection !== null) sections[currentSection].push(line);
    // lines before first ## heading (e.g. "# Definition") are silently skipped
  }

  // Trim leading/trailing blank lines from each section's content
  for (const key of Object.keys(sections)) {
    let lines = sections[key];
    while (lines.length > 0 && lines[0].trim() === '') lines.shift();
    while (lines.length > 0 && lines[lines.length - 1].trim() === '') lines.pop();
    sections[key] = lines.join('\n');
  }

  const tasks = parseTaskLines(taskLines);
  const hasStructure = sectionOrder.length > 0 || tasks.length > 0;
  return { sectionOrder, sections, tasks, hasStructure };
}

function serializeTask(task) {
  const lines = [`### ${task.heading}`, ''];
  if (task.type) lines.push(`**Type**: ${task.type}`);
  if (task.description) lines.push(`**Description**: ${task.description}`);
  if (task.criteria.length > 0) {
    lines.push('', '**Acceptance Criteria**:');
    for (const c of task.criteria) lines.push(`- ${c}`);
  }
  return lines.join('\n');
}

function serialize(frontmatter, sectionOrder, sections, tasks) {
  const parts = [
    '---',
    `id: "${frontmatter.id}"`,
    `name: "${frontmatter.name}"`,
    '---',
    '',
    `# Definition`,
    '',
  ];

  for (const key of sectionOrder) {
    parts.push(`## ${key}`, '');
    if (sections[key]) parts.push(sections[key], '');
  }

  if (tasks.length > 0) {
    parts.push('## Tasks', '');
    parts.push(...tasks.map(serializeTask).join('\n\n').split('\n'));
    parts.push('');
  }

  return parts.join('\n');
}

export default function DefinitionView({ content, onApprove, status }) {
  const { frontmatter: initFm, rest } = parseFrontmatter(content);
  const { sectionOrder: initOrder, sections: initSections, tasks: initTasks, hasStructure } = parseContent(rest);

  const [frontmatter, setFrontmatter] = useState(initFm);
  const [sectionOrder] = useState(initOrder);
  const [sections, setSections] = useState(initSections);
  const [tasks, setTasks] = useState(initTasks);

  // If no recognizable structure at all, fall back to raw textarea
  const hasFm = initFm.id || initFm.name;
  if (!hasFm && !hasStructure) {
    return <FallbackTextarea content={content} onApprove={onApprove} status={status} />;
  }

  function updateSection(key, value) {
    setSections(prev => ({ ...prev, [key]: value }));
  }

  function updateTask(i, field, value) {
    setTasks(prev => prev.map((t, idx) => idx === i ? { ...t, [field]: value } : t));
  }

  function updateCriterion(ti, ci, value) {
    setTasks(prev => prev.map((t, i) => {
      if (i !== ti) return t;
      const criteria = [...t.criteria];
      criteria[ci] = value;
      return { ...t, criteria };
    }));
  }

  function addCriterion(ti) {
    setTasks(prev => prev.map((t, i) =>
      i === ti ? { ...t, criteria: [...t.criteria, ''] } : t
    ));
  }

  function removeCriterion(ti, ci) {
    setTasks(prev => prev.map((t, i) => {
      if (i !== ti) return t;
      return { ...t, criteria: t.criteria.filter((_, k) => k !== ci) };
    }));
  }

  function addTask() {
    setTasks(prev => [...prev, { heading: 'NEW-001: New Task', type: 'feature', description: '', criteria: [''] }]);
  }

  function removeTask(i) {
    setTasks(prev => prev.filter((_, idx) => idx !== i));
  }

  function handleApprove() {
    onApprove(serialize(frontmatter, sectionOrder, sections, tasks));
  }

  return (
    <div>
      {/* Epic metadata */}
      <div style={styles.metaCard}>
        <div style={styles.metaRow}>
          <label style={styles.metaLabel}>Epic ID</label>
          <input
            style={styles.metaInput}
            value={frontmatter.id}
            onChange={e => setFrontmatter(prev => ({ ...prev, id: e.target.value }))}
            placeholder="EPIC-N"
          />
        </div>
        <div style={styles.metaRow}>
          <label style={styles.metaLabel}>Epic Name</label>
          <input
            style={styles.metaInput}
            value={frontmatter.name}
            onChange={e => setFrontmatter(prev => ({ ...prev, name: e.target.value }))}
            placeholder="Epic Name"
          />
        </div>
      </div>

      {/* Prose sections */}
      {sectionOrder.map(key => (
        <div key={key} style={styles.sectionCard}>
          <div style={styles.sectionHeading}>{key}</div>
          <textarea
            style={styles.sectionTextarea}
            value={sections[key] || ''}
            onChange={e => updateSection(key, e.target.value)}
            spellCheck={false}
            placeholder={`${key} content…`}
          />
        </div>
      ))}

      {/* Tasks */}
      {tasks.map((task, ti) => (
        <div key={ti} style={styles.taskCard}>
          <div style={styles.cardHeader}>
            <input
              style={styles.headingInput}
              value={task.heading}
              onChange={e => updateTask(ti, 'heading', e.target.value)}
              placeholder="EPIC-N-001: Task Name"
            />
            <button style={styles.removeBtn} onClick={() => removeTask(ti)} title="Remove task">✕</button>
          </div>

          <div style={styles.fieldRow}>
            <label style={styles.label}>Type</label>
            <input
              style={styles.typeInput}
              value={task.type}
              onChange={e => updateTask(ti, 'type', e.target.value)}
              placeholder="feature"
            />
          </div>

          <div style={styles.fieldRow}>
            <label style={styles.label}>Description</label>
            <textarea
              style={styles.descTextarea}
              value={task.description}
              onChange={e => updateTask(ti, 'description', e.target.value)}
              spellCheck={false}
              placeholder="Concrete description of what to implement."
            />
          </div>

          <div style={styles.fieldRow}>
            <label style={styles.label}>Acceptance Criteria</label>
            <div style={styles.criteriaList}>
              {task.criteria.map((c, ci) => (
                <div key={ci} style={styles.criterionRow}>
                  <input
                    style={styles.criterionInput}
                    value={c}
                    onChange={e => updateCriterion(ti, ci, e.target.value)}
                    placeholder="Criterion"
                  />
                  <button
                    style={styles.removeCritBtn}
                    onClick={() => removeCriterion(ti, ci)}
                    title="Remove criterion"
                  >✕</button>
                </div>
              ))}
              <button style={styles.addCritBtn} onClick={() => addCriterion(ti)}>+ Add criterion</button>
            </div>
          </div>
        </div>
      ))}

      <button style={styles.addTaskBtn} onClick={addTask}>+ Add task</button>
      <ApproveButton onApprove={handleApprove} status={status} />
    </div>
  );
}

function FallbackTextarea({ content, onApprove, status }) {
  const [text, setText] = useState(content);
  return (
    <div>
      <textarea
        style={{ width: '100%', height: 480, fontFamily: 'monospace', fontSize: 14, boxSizing: 'border-box', padding: 8 }}
        value={text}
        onChange={e => setText(e.target.value)}
        spellCheck={false}
      />
      <ApproveButton onApprove={() => onApprove(text)} status={status} />
    </div>
  );
}

const styles = {
  metaCard: {
    border: '1px solid #ccc',
    borderRadius: 4,
    marginBottom: 14,
    padding: 14,
    background: '#f0f4ff',
  },
  metaRow: {
    display: 'flex',
    alignItems: 'center',
    marginBottom: 8,
    gap: 8,
  },
  metaLabel: {
    fontFamily: 'monospace',
    fontSize: 12,
    color: '#555',
    width: 80,
    flexShrink: 0,
  },
  metaInput: {
    flex: 1,
    fontFamily: 'monospace',
    fontSize: 13,
    border: '1px solid #ccc',
    borderRadius: 3,
    padding: '3px 7px',
  },
  sectionCard: {
    border: '1px solid #ddd',
    borderRadius: 4,
    marginBottom: 10,
    padding: 12,
    background: '#fafafa',
  },
  sectionHeading: {
    fontFamily: 'monospace',
    fontSize: 13,
    fontWeight: 'bold',
    color: '#333',
    marginBottom: 6,
  },
  sectionTextarea: {
    width: '100%',
    minHeight: 72,
    fontFamily: 'monospace',
    fontSize: 13,
    boxSizing: 'border-box',
    padding: 6,
    resize: 'vertical',
    border: '1px solid #ccc',
    borderRadius: 3,
  },
  taskCard: {
    border: '1px solid #ccc',
    borderRadius: 4,
    marginBottom: 14,
    padding: 14,
    background: '#fafafa',
  },
  cardHeader: {
    display: 'flex',
    alignItems: 'center',
    marginBottom: 10,
  },
  headingInput: {
    flex: 1,
    fontFamily: 'monospace',
    fontSize: 14,
    fontWeight: 'bold',
    border: '1px solid #ccc',
    borderRadius: 3,
    padding: '4px 8px',
  },
  removeBtn: {
    marginLeft: 8,
    background: 'none',
    border: '1px solid #ccc',
    borderRadius: 3,
    cursor: 'pointer',
    padding: '2px 6px',
    color: '#c00',
  },
  fieldRow: {
    display: 'flex',
    alignItems: 'flex-start',
    marginBottom: 8,
    gap: 8,
  },
  label: {
    fontFamily: 'monospace',
    fontSize: 12,
    color: '#555',
    width: 130,
    flexShrink: 0,
    paddingTop: 4,
  },
  typeInput: {
    fontFamily: 'monospace',
    fontSize: 13,
    border: '1px solid #ccc',
    borderRadius: 3,
    padding: '3px 7px',
    width: 140,
  },
  descTextarea: {
    flex: 1,
    minHeight: 60,
    fontFamily: 'monospace',
    fontSize: 13,
    boxSizing: 'border-box',
    padding: 6,
    resize: 'vertical',
  },
  criteriaList: {
    flex: 1,
  },
  criterionRow: {
    display: 'flex',
    alignItems: 'center',
    marginBottom: 4,
    gap: 4,
  },
  criterionInput: {
    flex: 1,
    fontFamily: 'monospace',
    fontSize: 13,
    border: '1px solid #ccc',
    borderRadius: 3,
    padding: '3px 7px',
  },
  removeCritBtn: {
    background: 'none',
    border: '1px solid #ddd',
    borderRadius: 3,
    cursor: 'pointer',
    padding: '2px 5px',
    color: '#c00',
    fontSize: 11,
  },
  addCritBtn: {
    fontFamily: 'monospace',
    fontSize: 12,
    cursor: 'pointer',
    padding: '3px 8px',
    marginTop: 2,
  },
  addTaskBtn: {
    marginBottom: 12,
    padding: '6px 14px',
    fontFamily: 'monospace',
    cursor: 'pointer',
  },
};
