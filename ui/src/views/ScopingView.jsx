import React, { useState } from 'react';
import ApproveButton from '../ApproveButton';

// Parse SCOPED.md into sections: each section starts with a ## heading.
// Within each section, look for task-like lines starting with - or *.
function parseSections(markdown) {
  const lines = markdown.split('\n');
  const sections = [];
  let preamble = [];
  let current = null;

  for (const line of lines) {
    if (line.startsWith('## ')) {
      if (current) sections.push(current);
      current = { heading: line.slice(3), lines: [] };
    } else if (current) {
      current.lines.push(line);
    } else {
      preamble.push(line);
    }
  }
  if (current) sections.push(current);

  return { preamble, sections };
}

function serializeSections(preamble, sections) {
  const parts = [];
  if (preamble.some(l => l.trim())) {
    parts.push(preamble.join('\n').trimEnd());
  }
  for (const sec of sections) {
    parts.push('## ' + sec.heading + '\n' + sec.lines.join('\n').trimEnd());
  }
  return parts.join('\n\n') + '\n';
}

export default function ScopingView({ content, onApprove, status }) {
  const parsed = parseSections(content);
  const [sections, setSections] = useState(parsed.sections);
  const preamble = parsed.preamble;

  function updateHeading(i, val) {
    setSections(prev => prev.map((s, idx) => idx === i ? { ...s, heading: val } : s));
  }

  function updateLines(i, val) {
    setSections(prev => prev.map((s, idx) => idx === i ? { ...s, lines: val.split('\n') } : s));
  }

  function addSection() {
    setSections(prev => [...prev, { heading: 'New Section', lines: ['- Task 1'] }]);
  }

  function removeSection(i) {
    setSections(prev => prev.filter((_, idx) => idx !== i));
  }

  if (!sections.length) {
    return <FallbackTextarea content={content} onApprove={onApprove} status={status} />;
  }

  return (
    <div>
      {sections.map((sec, i) => (
        <div key={i} style={styles.card}>
          <div style={styles.cardHeader}>
            <input
              style={styles.headingInput}
              value={sec.heading}
              onChange={e => updateHeading(i, e.target.value)}
            />
            <button
              style={styles.removeBtn}
              onClick={() => removeSection(i)}
              title="Remove section"
            >✕</button>
          </div>
          <textarea
            style={styles.bodyTextarea}
            value={sec.lines.join('\n')}
            onChange={e => updateLines(i, e.target.value)}
            spellCheck={false}
          />
        </div>
      ))}
      <button style={styles.addBtn} onClick={addSection}>+ Add Section</button>
      <ApproveButton onApprove={() => onApprove(serializeSections(preamble, sections))} status={status} />
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
    marginBottom: 10,
    padding: 12,
    background: '#fafafa',
  },
  cardHeader: {
    display: 'flex',
    alignItems: 'center',
    marginBottom: 6,
  },
  headingInput: {
    flex: 1,
    fontFamily: 'monospace',
    fontSize: 14,
    fontWeight: 'bold',
    border: '1px solid #ccc',
    borderRadius: 3,
    padding: '3px 7px',
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
  bodyTextarea: {
    width: '100%',
    minHeight: 80,
    fontFamily: 'monospace',
    fontSize: 13,
    boxSizing: 'border-box',
    padding: 6,
    resize: 'vertical',
  },
  addBtn: {
    marginBottom: 12,
    padding: '6px 14px',
    fontFamily: 'monospace',
    cursor: 'pointer',
  },
};
