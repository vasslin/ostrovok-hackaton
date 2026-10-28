import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { spawnSync } from 'node:child_process';
test('Portable HTML bundles valid JavaScript and geography without external script or stylesheet files',async()=>{
  const html=await readFile(new URL('../dist/standalone.html',import.meta.url),'utf8');
  assert.ok(!/<script[^>]+src=/.test(html));
  assert.ok(!/<link[^>]+rel="stylesheet"/.test(html));
  assert.ok(!html.includes('const geographyReady = fetch'));
  const script=html.match(/<script type="module">([\s\S]*?)<\/script>/)?.[1];
  assert.ok(script);
  const syntax = spawnSync(process.execPath, ['--input-type=module', '--check'], { input: script, encoding: 'utf8' });
  assert.equal(syntax.status, 0, syntax.stderr || syntax.error?.message);
  assert.ok(script.includes('MultiPolygon'));
});
