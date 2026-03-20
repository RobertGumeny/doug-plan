import React, { useState } from 'react';
import ApproveButton from '../ApproveButton';

export default function TasksView({ content, onApprove, status }) {
  const [text, setText] = useState(content);

  return (
    <div>
      <textarea
        style={styles.textarea}
        value={text}
        onChange={e => setText(e.target.value)}
        spellCheck={false}
      />
      <ApproveButton onApprove={() => onApprove(text)} status={status} />
    </div>
  );
}

const styles = {
  textarea: {
    width: '100%',
    height: 520,
    fontFamily: 'monospace',
    fontSize: 14,
    boxSizing: 'border-box',
    padding: 8,
  },
};
