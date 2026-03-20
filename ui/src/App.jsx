import React, { useState, useEffect } from 'react';
import VisionView from './views/VisionView';
import RoadmapView from './views/RoadmapView';
import ScopingView from './views/ScopingView';
import PRDView from './views/PRDView';
import TasksView from './views/TasksView';

const VIEWS = {
  Discovery: VisionView,
  Roadmapping: RoadmapView,
  Scoping: ScopingView,
  PRD: PRDView,
  Tasks: TasksView,
};

export default function App() {
  const [stage, setStage] = useState(null);
  const [content, setContent] = useState('');
  const [status, setStatus] = useState('');
  const [approved, setApproved] = useState(false);

  useEffect(() => {
    fetch('/artifact')
      .then(r => r.json())
      .then(data => {
        setStage(data.stage);
        setContent(data.content);
      })
      .catch(() => setStatus('Error loading artifact.'));
  }, []);

  function handleApprove(updatedContent) {
    setStatus('Sending…');
    fetch('/approve', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content: updatedContent }),
    })
      .then(r => {
        if (!r.ok) throw new Error('Server error');
        setApproved(true);
      })
      .catch(() => setStatus('Error — please try again.'));
  }

  if (approved) {
    return (
      <div style={styles.page}>
        <h1 style={styles.doneHeading}>Approved. You may close this tab.</h1>
      </div>
    );
  }

  if (!stage) {
    return (
      <div style={styles.page}>
        <p>{status || 'Loading…'}</p>
      </div>
    );
  }

  const View = VIEWS[stage];
  if (!View) {
    return (
      <div style={styles.page}>
        <p>Unknown stage: {stage}</p>
      </div>
    );
  }

  return (
    <div style={styles.page}>
      <h1 style={styles.heading}>Review: {stage}</h1>
      <p style={styles.subtitle}>
        Review and edit the artifact below, then click <strong>Approve</strong> to advance the pipeline.
      </p>
      <View content={content} onApprove={handleApprove} status={status} setStatus={setStatus} />
    </div>
  );
}

const styles = {
  page: {
    fontFamily: 'monospace',
    maxWidth: 900,
    margin: '40px auto',
    padding: '0 20px',
  },
  heading: { marginBottom: 4 },
  subtitle: { marginTop: 4, color: '#555' },
  doneHeading: { color: '#080' },
};
