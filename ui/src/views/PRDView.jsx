import React, { useState } from 'react';
import ApproveButton from '../ApproveButton';

// PRDView renders a split layout: PRD.md prose on the left, tasks.yaml on the right.
// When no secondaryContent is provided it falls back to a single textarea.
export default function PRDView({ content, secondaryContent, onApprove, status }) {
  const [prdText, setPrdText] = useState(content);
  const [tasksText, setTasksText] = useState(secondaryContent || '');
  const hasTasks = (secondaryContent || '').trim() !== '';

  if (hasTasks) {
    return (
      <div>
        <div style={styles.splitLayout}>
          <div style={styles.pane}>
            <div style={styles.paneLabel}>PRD.md</div>
            <textarea
              style={styles.textarea}
              value={prdText}
              onChange={e => setPrdText(e.target.value)}
              spellCheck={false}
            />
          </div>
          <div style={styles.pane}>
            <div style={styles.paneLabel}>tasks.yaml</div>
            <textarea
              style={{ ...styles.textarea, fontFamily: 'monospace', fontSize: 12 }}
              value={tasksText}
              onChange={e => setTasksText(e.target.value)}
              spellCheck={false}
            />
          </div>
        </div>
        <ApproveButton onApprove={() => onApprove(prdText, tasksText)} status={status} />
      </div>
    );
  }

  return (
    <div>
      <textarea
        style={styles.singleTextarea}
        value={prdText}
        onChange={e => setPrdText(e.target.value)}
        spellCheck={false}
      />
      <ApproveButton onApprove={() => onApprove(prdText, '')} status={status} />
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
    height: 520,
    fontFamily: 'monospace',
    fontSize: 14,
    boxSizing: 'border-box',
    padding: 8,
  },
};
