import http from 'node:http';
import { readFile, stat } from 'node:fs/promises';
import { dirname, resolve, extname, sep } from 'node:path';
import { fileURLToPath } from 'node:url';
const root = resolve(dirname(fileURLToPath(import.meta.url)), 'dist');
const port = Number(process.env.PORT || 5173);
const types = { '.html': 'text/html; charset=utf-8', '.js': 'text/javascript; charset=utf-8', '.css': 'text/css; charset=utf-8', '.json': 'application/json; charset=utf-8', '.svg': 'image/svg+xml' };
const server = http.createServer(async (req, res) => {
    try {
        const url = new URL(req.url, 'http://localhost');
        if (url.pathname.startsWith('/api/')) {
            const target = new URL(req.url, process.env.GATEWAY_URL || 'http://127.0.0.1:8080');
            const upstream = http.request(target, { method: req.method, headers: { ...req.headers, host: target.host } }, response => {
                res.writeHead(response.statusCode || 502, response.headers);
                response.pipe(res);
            });
            upstream.setTimeout(400000, () => upstream.destroy(new Error('Gateway timeout')));
            upstream.on('error', () => { if (!res.headersSent) res.writeHead(502); res.end('Gateway unavailable'); });
            req.on('aborted', () => upstream.destroy());
            req.pipe(upstream);
            return;
        }
        const name = decodeURIComponent(url.pathname);
        const path = resolve(root, `.${name === '/' ? '/index.html' : name}`);
        if (path !== root && !path.startsWith(root + sep)) {
            res.writeHead(403);
            res.end('Forbidden');
            return;
        }
        const fileStat = await stat(path);
        if (!fileStat.isFile()) {
            res.writeHead(404);
            res.end('Not found');
            return;
        }
        const data = await readFile(path);
        res.writeHead(200, { 'Content-Type': types[extname(path)] || 'application/octet-stream', 'Cache-Control': 'no-cache', 'X-Content-Type-Options': 'nosniff' });
        res.end(data);
    }
    catch {
        res.writeHead(404, { 'Content-Type': 'text/plain; charset=utf-8' });
        res.end('Not found');
    }
});
server.listen(port, '127.0.0.1', () => console.log(`Local: http://127.0.0.1:${port}`));
for (const signal of ['SIGINT', 'SIGTERM'])
    process.on(signal, () => server.close(() => process.exit(0)));
