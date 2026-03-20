import React, { useState } from 'react';
import ApproveButton from '../ApproveButton';

// Parse ROADMAP.md into an array of { heading, body } epic objects.
// Lines before the first ## heading are kept as preamble.
function parseEpics(markdown) {
  const lines = markdown.split('\n');
  const epics = [];
  let preamble = [];
  let current = null;

  for (const line of lines) {
    if (line.startsWith('## ')) {
      if (current) epics.push(current);
      else if (preamble.length) {
        // trim trailing blank lines from preamble
      }
      current = { heading: line.slice(3), body: [] };
    } else if (current) {
      current.body.push(line);
    } else {
      preamble.push(line);
    }
  }
  if (current) epics.push(current);

  // Trim trailing blank lines from each body
  for (const epic of epics) {
    while (epic.body.length && epic.body[epic.body.length - 1].trim() === '') {
      epic.body.pop();
    }
  }

  return { preamble, epics };
}

function serializeEpics(preamble, epics) {
  const parts = [];
  if (preamble.length) {
    parts.push(preamble.join('\n'));
  }
  for (const epic of epics) {
    parts.push('## ' + epic.heading + '\n' + epic.body.join('\n'));
  }
  return parts.join('\n\n') + '\n';
}

export default function RoadmapView({ content, onApprove, status }) {
  const parsed = parseEpics(content);
  const [epics, setEpics] = useState(parsed.epics);
  const preamble = parsed.preamble;

  function updateHeading(i, val) {
    setEpics(prev => prev.map((e, idx) => idx === i ? { ...e, heading: val } : e));
  }

  function updateBody(i, val) {
    setEpics(prev => prev.map((e, idx) => idx === i ? { ...e, body: val.split('\n') } : e));
  }

  function moveUp(i) {
    if (i === 0) return;
    setEpics(prev => {
      const next = [...prev];
      [next[i - 1], next[i]] = [next[i], next[i - 1]];
      return next;
    });
  }

  function moveDown(i) {
    setEpics(prev => {
      if (i >= prev.length - 1) return prev;
      const next = [...prev];
      [next[i], next[i + 1]] = [next[i + 1], next[i]];
      return next;
    });
  }

  function handleApprove() {
    onApprove(serializeEpics(preamble, epics));
  }

  if (!epics.length) {
    return (
      <div>
        <p style={{ color: '#888' }}>No epic sections (## headings) found. Edit as raw text:</p>
        <FallbackTextarea content={content} onApprove={onApprove} status={status} />
      </div>
    );
  }

  return (
    <div>
      {epics.map((epic, i) => (
        <div key={i} style={styles.card}>
          <div style={styles.cardHeader}>
            <input
              style={styles.headingInput}
              value={epic.heading}
              onChange={e => updateHeading(i, e.target.value)}
            />
            <div style={styles.reorderButtons}>
              <button
                style={styles.reorderBtn}
                onClick={() => moveUp(i)}
                disabled={i === 0}
                title="Move up"
              >↑</button>
              <button
                style={styles.reorderBtn}
                onClick={() => moveDown(i)}
                disabled={i === epics.length - 1}
                title="Move down"
              >↓</button>
            </div>
          </div>
          <textarea
            style={styles.bodyTextarea}
            value={epic.body.join('\n')}
            onChange={e => updateBody(i, e.target.value)}
            spellCheck={false}
          />
        </div>
      ))}
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
  card: {
    border: '1px solid #ccc',
    borderRadius: 4,
    marginBottom: 12,
    padding: 12,
    background: '#fafafa',
  },
  cardHeader: {
    display: 'flex',
    alignItems: 'center',
    marginBottom: 8,
  },
  headingInput: {
    flex: 1,
    fontFamily: 'monospace',
    fontSize: 15,
    fontWeight: 'bold',
    border: '1px solid #ccc',
    borderRadius: 3,
    padding: '4px 8px',
  },
  reorderButtons: {
    marginLeft: 8,
    display: 'flex',
    gap: 4,
  },
  reorderBtn: {
    padding: '2px 8px',
    fontSize: 14,
    cursor: 'pointer',
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
};
