import { test, expect } from '@playwright/test';
import { ChildProcess, spawn } from 'node:child_process';
import fs from 'node:fs';
import net from 'node:net';
import path from 'node:path';

import { waitForServer } from './helpers/waitForServer';

const fixturePath = path.join(__dirname, 'fixtures', 'sample.png');

test.describe('danmu enabled', () => {
  test('upload page renders', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/Upload Media/);
    await expect(page.locator('#fileInput')).toBeVisible();
    await expect(page.locator('#uploadBtn')).toBeVisible();
    await expect(page.locator('#commentForm')).toBeVisible();
  });

  test('upload works', async ({ page }) => {
    await page.goto('/');
    await page.setInputFiles('#fileInput', fixturePath);
    await page.click('#uploadBtn');
    await expect(page.locator('#successState')).toBeVisible({ timeout: 10000 });
  });

  test('upload works and comments can be submitted', async ({ page, request }) => {
    const buffer = fs.readFileSync(fixturePath);
    const uploadRes = await request.post('/upload', {
      multipart: {
        picture: {
          name: 'sample.png',
          mimeType: 'image/png',
          buffer,
        },
      },
    });
    expect(uploadRes.ok()).toBeTruthy();

    await page.goto('/');
    await page.waitForSelector('#mediaGrid button', { timeout: 10000 });
    await page.locator('#mediaGrid button').first().click();
    await page.fill('#commentInput', 'Great photo!');
    await page.click('#commentForm button[type="submit"]');
    await expect(page.locator('#commentStatus')).toHaveText(/Comment saved\./);
  });
});

test.describe('danmu disabled', () => {
  let port: number;
  let baseURL: string;
  let serverProcess: ChildProcess;

  test.beforeAll(async () => {
    port = await new Promise((resolve, reject) => {
      const server = net.createServer();
      server.listen(0, '127.0.0.1', () => {
        const address = server.address();
        server.close(() => {
          if (address && typeof address === 'object') {
            resolve(address.port);
          } else {
            reject(new Error('failed to resolve free port'));
          }
        });
      });
      server.on('error', reject);
    });
    baseURL = `http://127.0.0.1:${port}`;

    serverProcess = spawn('tsx', ['scripts/e2e-server.ts'], {
      stdio: 'inherit',
      env: {
        ...process.env,
        GALLERY_PORT: String(port),
        GALLERY_DANMU_ENABLED: '0',
        GALLERY_WORKDIR: path.join(process.cwd(), '.e2e-workdir', 'danmu-off-upload'),
        GALLERY_APPROVALS_ENABLED: '0',
      },
    });
    await waitForServer(`${baseURL}/slideshow`);
  });

  test.afterAll(() => {
    if (serverProcess && !serverProcess.killed) {
      serverProcess.kill('SIGTERM');
    }
  });

  test('comment section is hidden', async ({ page }) => {
    await page.goto(`${baseURL}/`);
    await expect(page).toHaveTitle(/Upload Media/);
    await expect(page.locator('#commentForm')).toHaveCount(0);
  });

  test('upload works', async ({ page }) => {
    await page.goto(`${baseURL}/`);
    await page.setInputFiles('#fileInput', fixturePath);
    await page.click('#uploadBtn');
    await expect(page.locator('#successState')).toBeVisible({ timeout: 10000 });
  });
});
