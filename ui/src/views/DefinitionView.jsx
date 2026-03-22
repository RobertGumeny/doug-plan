import React, { useState } from 'react';
import ApproveButton from '../ApproveButton';

// Parse DEFINITION.md into preamble + structured task list.
// Tasks begin after the "## Tasks" heading and are delimited by "### " lines.
function parseTasks(markdown) {
  const lines = markdown.split('\n');
  const preambleLines = [];
  const tasks = [];
  let inTasksSection = false;
  let currentTask = null;
  let inCriteria = false;

  for (const line of lines) {
    if (!inTasksSection) {
      preambleLines.push(line);
      if (line === '## Tasks') {
        inTasksSection = true;
      }
      continue;
    }

    if (line.startsWith('### ')) {
      if (currentTask) tasks.push(currentTask);
      currentTask = { heading: line.slice(4), type: '', description: '', criteria: [] };
      inCriteria = false;
      continue;
    }

    if (!currentTask) continue;

    const typeMatch = line.match(/^\*\*Type\*\*:\s*(.*)/);
    if (typeMatch) {
      currentTask.type = typeMatch[1].trim();
      inCriteria = false;
      continue;
    }

    const descMatch = line.match(/^\*\*Description\*\*:\s*(.*)/);
    if (descMatch) {
      currentTask.description = descMatch[1].trim();
      inCriteria = false;
      continue;
    }

    if (line === '**Acceptance Criteria**:') {
      inCriteria = true;
      continue;
    }

    if (inCriteria && line.startsWith('- ')) {
      currentTask.criteria.push(line.slice(2));
      continue;
    }

    if (inCriteria && line.trim() !== '') {
      inCriteria = false;
    }
  }

  if (currentTask) tasks.push(currentTask);
  return { preamble: preambleLines, tasks };
}

function serializeTasks(preamble, tasks) {
  const preambleStr = preamble.join('\n');
  if (tasks.length === 0) return preambleStr + '\n';

  const taskParts = tasks.map(task => {
    const lines = [`### ${task.heading}`, ''];
    if (task.type) lines.push(`**Type**: ${task.type}`);
    if (task.description) lines.push(`**Description**: ${task.description}`);
    if (task.criteria.length > 0) {
      lines.push('', '**Acceptance Criteria**:');
      for (const c of task.criteria) {
        lines.push(`- ${c}`);
      }
    }
    return lines.join('\n');
  });

  return preambleStr + '\n\n' + taskParts.join('\n\n') + '\n';
}

export default function DefinitionView({ content, onApprove, status }) {
  const parsed = parseTasks(content);
  const [tasks, setTasks] = useState(parsed.tasks);
  const preamble = parsed.preamble;

  if (tasks.length === 0) {
    return <FallbackTextarea content={content} onApprove={onApprove} status={status} />;
  }

  function updateTask(i, field, value) {
    setTasks(prev => prev.map((t, idx) => idx === i ? { ...t, [field]: value } : t));
  }

  function updateCriterion(taskIdx, critIdx, value) {
    setTasks(prev => prev.map((t, i) => {
      if (i !== taskIdx) return t;
      const criteria = [...t.criteria];
      criteria[critIdx] = value;
      return { ...t, criteria };
    }));
  }

  function addCriterion(taskIdx) {
    setTasks(prev => prev.map((t, i) =>
      i === taskIdx ? { ...t, criteria: [...t.criteria, ''] } : t
    ));
  }

  function removeCriterion(taskIdx, critIdx) {
    setTasks(prev => prev.map((t, i) => {
      if (i !== taskIdx) return t;
      return { ...t, criteria: t.criteria.filter((_, ci) => ci !== critIdx) };
    }));
  }

  function addTask() {
    setTasks(prev => [...prev, { heading: 'NEW-001: New Task', type: 'feature', description: '', criteria: [''] }]);
  }

  function removeTask(i) {
    setTasks(prev => prev.filter((_, idx) => idx !== i));
  }

  return (
    <div>
      {tasks.map((task, ti) => (
        <div key={ti} style={styles.card}>
          <div style={styles.cardHeader}>
            <input
              style={styles.headingInput}
              value={task.heading}
              onChange={e => updateTask(ti, 'heading', e.target.value)}
              placeholder="EPIC-ID-001: Task Name"
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
      <ApproveButton onApprove={() => onApprove(serializeTasks(preamble, tasks))} status={status} />
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
  card: {
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
