import { readFile, writeFile } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = dirname(fileURLToPath(import.meta.url));
const read = name => readFile(resolve(root, 'dist', name), 'utf8');
const files = ['icons.js', 'model.js', 'api.js', 'templates.js', 'map.js', 'mobile.js', 'app.js'];
const geography = JSON.parse(await read('assets/countries.geojson'));

const modules = [];
for (const file of files) {
  let source = await read(file);
  source = source.replace(/^import .*?;\r?\n/gm, '').replace(/^export /gm, '');

  if (file === 'map.js') {
    source = source.replace(
      'let countries = null;',
      `let countries = ${JSON.stringify(geography.features).replace(/</g, '\\u003c')};`
    );
    source = source.replace(
      /^const geographyReady = fetch[\s\S]*?\.catch\(\(\) => \{\s*\}\);/m,
      'const geographyReady = Promise.resolve();'
    );
    if (source.includes("fetch(new URL('./assets/countries.geojson'")) {
      throw new Error('Could not inline map geography');
    }
  }

  modules.push(`// ${file}\n${source}`);
}

let html = await read('index.html');
const css = (await read('styles.css')).replace(/^@import[^\n]+\n/, '');
const templates = await read('templates.html');

html = html
  .replace('<link rel="stylesheet" href="./styles.css">', `<style>\n${css}\n</style>`)
  .replace('<script src="./config.js"></script>', `<script type="text/plain" id="inline-templates">${templates}</script>\n  <script>${await read('config.js')}</script>`)
  .replace('<script type="module" src="./app.js"></script>', `<script type="module">\n${modules.join('\n\n')}\n</script>`);

await writeFile(resolve(root, 'dist', 'standalone.html'), html);
await writeFile(resolve(root, '..', 'travel-demo.html'), html);
console.log('Portable demo: dist/standalone.html and ../travel-demo.html');
