const { defineConfig } = require('playwright/test');
const executablePath = process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE;
const chromium = { browserName: 'chromium', ...(executablePath ? { launchOptions: { executablePath } } : {}) };
module.exports = defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  workers: 2,
  timeout: 45000,
  expect: { timeout: 7000 },
  reporter: [['list'], ['html', { outputFolder: 'playwright-report', open: 'never' }]],
  use: { baseURL: 'http://127.0.0.1:4173', screenshot: 'only-on-failure', trace: 'retain-on-failure' },
  webServer: { command: 'node serve.mjs', url: 'http://127.0.0.1:4173', reuseExistingServer: !process.env.CI },
  projects: [
    ...[[320, 568], [390, 844], [430, 932], [844, 390]].map(([width, height]) => ({ name: `chromium-phone-${width}`, use: { ...chromium, viewport: { width, height }, isMobile: true, hasTouch: true } })),
    { name: 'chromium-tablet', use: { ...chromium, viewport: { width: 768, height: 1024 }, isMobile: true, hasTouch: true } },
    { name: 'chromium-desktop', use: { ...chromium, viewport: { width: 1440, height: 900 } } },
    { name: 'webkit-phone', use: { browserName: 'webkit', viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true } }
  ]
});
