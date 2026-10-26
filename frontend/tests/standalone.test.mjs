import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import vm from 'node:vm';
test('Portable HTML bundles valid JavaScript and geography without external script or stylesheet files',async()=>{
  const html=await readFile(new URL('../dist/standalone.html',import.meta.url),'utf8');
  assert.ok(!/<script[^>]+src=/.test(html));
  assert.ok(!/<link[^>]+rel="stylesheet"/.test(html));
  assert.ok(!html.includes('const geographyReady = fetch'));
  const script=html.match(/<script type="module">([\s\S]*?)<\/script>/)?.[1];
  assert.ok(script);
  assert.doesNotThrow(()=>new vm.Script(script));
  assert.ok(script.includes('MultiPolygon'));
});
