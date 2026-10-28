const templates = new Map();

export async function loadViews() {
  const embedded = document.getElementById('inline-templates');
  let source;

  if (embedded) {
    source = embedded.textContent;
  } else {
    const response = await fetch('./templates.html');
    if (!response.ok) throw new Error('Не удалось загрузить шаблоны интерфейса.');
    source = await response.text();
  }

  for (const [, name, html] of source.matchAll(/<template id="([\w-]+)">([\s\S]*?)<\/template>/g)) {
    templates.set(name, html.replace(/^\n/, '').replace(/\n$/, ''));
  }
}

export function view(name, values = {}) {
  const html = templates.get(name);
  if (html === undefined) throw new Error(`Не найден HTML-шаблон: ${name}`);
  return html.replace(/\{\{([\w$]+)\}\}/g, (_, field) => String(values[field]));
}
