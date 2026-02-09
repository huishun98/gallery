import { test, expect } from '@playwright/test';
import { ChildProcess, spawn } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';

import { getFreePort } from './helpers/getFreePort';
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
    port = await getFreePort();
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

test.describe('uploads disabled', () => {
  let port: number;
  let baseURL: string;
  let serverProcess: ChildProcess;

  test.beforeAll(async () => {
    port = await getFreePort();
    baseURL = `http://127.0.0.1:${port}`;

    serverProcess = spawn('tsx', ['scripts/e2e-server.ts'], {
      stdio: 'inherit',
      env: {
        ...process.env,
        GALLERY_PORT: String(port),
        GALLERY_DANMU_ENABLED: '1',
        GALLERY_APPROVALS_ENABLED: '0',
        GALLERY_UPLOADS_ENABLED: '0',
        GALLERY_WORKDIR: path.join(process.cwd(), '.e2e-workdir', 'uploads-off-upload'),
      },
    });
    await waitForServer(`${baseURL}/slideshow`);
  });

  test.afterAll(() => {
    if (serverProcess && !serverProcess.killed) {
      serverProcess.kill('SIGTERM');
    }
  });

  test('upload form hidden, comments visible', async ({ page }) => {
    await page.goto(`${baseURL}/`);
    await expect(page).toHaveTitle(/Upload Media/);
    await expect(page.locator('#uploadForm')).toHaveCount(0);
    await expect(page.locator('#progressWrapper')).toHaveCount(0);
    await expect(page.locator('#successState')).toHaveCount(0);
    await expect(page.locator('#commentForm')).toBeVisible();
  });

  test('upload endpoint is not available', async ({ request }) => {
    const buffer = fs.readFileSync(fixturePath);
    const uploadRes = await request.post(`${baseURL}/upload`, {
      multipart: {
        picture: {
          name: 'sample.png',
          mimeType: 'image/png',
          buffer,
        },
      },
    });
    expect(uploadRes.status()).toBe(404);
  });
});
