#!/usr/bin/env node
// Builds the React app to a single self-contained HTML file with no CDN dependencies.
// Output: ../internal/ui/bundle.html

const esbuild = require('esbuild');
const path = require('path');
const fs = require('fs');

const outFile = path.resolve(__dirname, '../internal/ui/bundle.html');

async function build() {
  const result = await esbuild.build({
    entryPoints: [path.resolve(__dirname, 'src/index.jsx')],
    bundle: true,
    minify: true,
    format: 'iife',
    platform: 'browser',
    jsx: 'automatic',
    write: false,
  });

  const js = result.outputFiles[0].text;

  const html = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Doug Plan — Review</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; }
    body { margin: 0; background: #fff; color: #222; }
    textarea { border: 1px solid #ccc; border-radius: 3px; resize: vertical; outline-offset: 2px; }
    input[type="text"], input:not([type]) { outline-offset: 2px; }
    button { border-radius: 3px; }
  </style>
</head>
<body>
  <div id="root"></div>
  <script>${js}</script>
</body>
</html>
`;

  fs.writeFileSync(outFile, html, 'utf8');
  console.log('Built:', outFile, `(${(html.length / 1024).toFixed(1)} KB)`);
}

build().catch(err => {
  console.error(err);
  process.exit(1);
});
