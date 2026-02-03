import { test, expect } from '@playwright/test';
import { ChildProcess, spawn } from 'node:child_process';
import net from 'node:net';
import path from 'node:path';

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
