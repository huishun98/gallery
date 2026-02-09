import { test, expect } from '@playwright/test';
import { ChildProcess, spawn } from 'node:child_process';
import path from 'node:path';

import { getFreePort } from './helpers/getFreePort';
import { waitForServer } from './helpers/waitForServer';

test.describe('danmu enabled', () => {
  test('slideshow page renders with danmu and QR', async ({ page, request }) => {
    await page.goto('/slideshow');
    await expect(page).toHaveTitle(/Dynamic Slideshow/);
    await expect(page.locator('#danmu')).toBeVisible();
    await expect(page.locator('#qr')).toBeVisible();

    const qrRes = await request.get('/qr');
    expect(qrRes.status()).toBe(200);
    expect(qrRes.headers()['content-type']).toContain('image/png');

    const qrBox = await page.locator('#qr').boundingBox();
    const viewport = page.viewportSize();
    expect(qrBox).not.toBeNull();
    expect(viewport).not.toBeNull();

    if (qrBox && viewport) {
      expect(qrBox.x + qrBox.width).toBeGreaterThan(viewport.width - 40);
      expect(qrBox.y + qrBox.height).toBeGreaterThan(viewport.height - 40);
    }
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
        GALLERY_APPROVALS_ENABLED: '0',
        GALLERY_WORKDIR: path.join(process.cwd(), '.e2e-workdir', 'danmu-off-slideshow'),
      },
    });
    await waitForServer(`${baseURL}/slideshow`);
  });

  test.afterAll(() => {
    if (serverProcess && !serverProcess.killed) {
      serverProcess.kill('SIGTERM');
    }
  });

  test('slideshow hides danmu overlay', async ({ page }) => {
    await page.goto(`${baseURL}/slideshow`);
    await expect(page).toHaveTitle(/Dynamic Slideshow/);
    await expect(page.locator('#danmu')).toHaveCount(0);
    await expect(page.locator('#qr')).toBeVisible();
  });
});

test.describe('empty media banner', () => {
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
        GALLERY_WORKDIR: path.join(process.cwd(), '.e2e-workdir', 'empty-media-banner'),
      },
    });
    await waitForServer(`${baseURL}/slideshow`);
  });

  test.afterAll(() => {
    if (serverProcess && !serverProcess.killed) {
      serverProcess.kill('SIGTERM');
    }
  });

  test('shows media folder path and copy button when empty', async ({ page }) => {
    await page.goto(`${baseURL}/slideshow`);
    await expect(page).toHaveTitle(/Dynamic Slideshow/);
    await expect(page.locator('#emptyBanner')).toBeVisible();
    await expect(page.locator('#mediaPath')).toContainText('.e2e-workdir');
    await expect(page.locator('#copyMediaPath')).toBeVisible();
  });

  test('copy button writes folder path to clipboard', async ({ page, context }) => {
    await context.grantPermissions(['clipboard-read', 'clipboard-write']);
    await page.goto(`${baseURL}/slideshow`);
    await expect(page.locator('#emptyBanner')).toBeVisible();

    const expected = await page.locator('#mediaPath').innerText();
    await page.click('#copyMediaPath');

    const copied = await page.evaluate(() => navigator.clipboard.readText());
    expect(copied).toBe(expected);
  });
});
