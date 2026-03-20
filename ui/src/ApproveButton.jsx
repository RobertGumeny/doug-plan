import React from 'react';

export default function ApproveButton({ onApprove, status }) {
  return (
    <div style={styles.row}>
      <button style={styles.button} type="button" onClick={onApprove}>
        Approve
      </button>
      {status && <span style={styles.status}>{status}</span>}
    </div>
  );
}

const styles = {
  row: { marginTop: 10 },
  button: {
    padding: '8px 28px',
    fontSize: 16,
    cursor: 'pointer',
  },
  status: {
    marginLeft: 12,
    color: '#080',
    fontFamily: 'monospace',
  },
};
