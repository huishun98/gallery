import { test, expect } from '@playwright/test';
import { ChildProcess, spawn } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';

import { waitForServer } from './helpers/waitForServer';

const clearDir = (dir: string) => {
  if (!fs.existsSync(dir)) {
    return;
  }
  for (const entry of fs.readdirSync(dir)) {
    fs.rmSync(path.join(dir, entry), { recursive: true, force: true });
  }
};

test.describe.serial('approvals disabled', () => {
  test('admin review not available', async ({ request }) => {
    const res = await request.get('/admin/review');
    expect(res.status()).toBe(404);
  });
});

test.describe.serial('approvals enabled', () => {
  const port = 8002;
  const baseURL = `http://127.0.0.1:${port}`;
  const workDir = path.join(process.cwd(), '.e2e-workdir', 'approvals-on');
  const dataDir = path.join(workDir, '.data');
  const pendingDir = path.join(dataDir, 'media', 'pending');
  const mediaDir = path.join(dataDir, 'media', 'media');
  const rejectedDir = path.join(dataDir, 'media', 'rejected');
  let serverProcess: ChildProcess;

  test.beforeAll(async () => {
    fs.mkdirSync(pendingDir, { recursive: true });
    serverProcess = spawn('tsx', ['scripts/e2e-server.ts'], {
      stdio: 'inherit',
      env: {
        ...process.env,
        GALLERY_PORT: String(port),
        GALLERY_APPROVALS_ENABLED: '1',
        GALLERY_DANMU_ENABLED: '1',
        GALLERY_WORKDIR: workDir,
      },
    });
    await waitForServer(`${baseURL}/slideshow`);
  });

  test.afterAll(() => {
    if (serverProcess && !serverProcess.killed) {
      serverProcess.kill('SIGTERM');
    }
  });

  test('admin review requires auth', async ({ request }) => {
    await waitForServer(`${baseURL}/slideshow`);
    const res = await request.get(`${baseURL}/admin/review`);
    expect(res.status()).toBe(401);
  });

  test('admin review renders with auth', async ({ browser }) => {
    await waitForServer(`${baseURL}/slideshow`);
    const context = await browser.newContext({
      baseURL,
      httpCredentials: { username: 'admin', password: 'admin' },
    });
    const page = await context.newPage();
    await page.goto('/admin/review');
    await expect(page).toHaveTitle(/Media Review/);
    await expect(page.locator('#content')).toContainText(/No pending media/);
    await expect(page.locator('#approveBtn')).toBeHidden();
    await expect(page.locator('#rejectBtn')).toBeHidden();
    await context.close();
  });

  test('admin review approve moves pending media to media folder', async ({ browser }) => {
    await waitForServer(`${baseURL}/slideshow`);
    fs.mkdirSync(pendingDir, { recursive: true });
    fs.mkdirSync(mediaDir, { recursive: true });
    clearDir(pendingDir);

    const filename = `approve-${Date.now()}.jpg`;
    const pendingPath = path.join(pendingDir, filename);
    fs.writeFileSync(pendingPath, 'fake');

    const context = await browser.newContext({
      baseURL,
      httpCredentials: { username: 'admin', password: 'admin' },
    });
    const page = await context.newPage();
    await page.goto('/admin/review');
    await expect(page.locator('#approveBtn')).toBeVisible();
    await page.click('#approveBtn');
    await expect(page.locator('#content')).toContainText(/No pending media/);
    expect(fs.existsSync(path.join(mediaDir, filename))).toBeTruthy();
    expect(fs.existsSync(pendingPath)).toBeFalsy();
    await context.close();
  });

  test('admin review reject moves pending media to rejected folder', async ({ browser }) => {
    await waitForServer(`${baseURL}/slideshow`);
    fs.mkdirSync(pendingDir, { recursive: true });
    fs.mkdirSync(rejectedDir, { recursive: true });
    clearDir(pendingDir);

    const filename = `reject-${Date.now()}.jpg`;
    const pendingPath = path.join(pendingDir, filename);
    fs.writeFileSync(pendingPath, 'fake');

    const context = await browser.newContext({
      baseURL,
      httpCredentials: { username: 'admin', password: 'admin' },
    });
    const page = await context.newPage();
    await page.goto('/admin/review');
    await expect(page.locator('#rejectBtn')).toBeVisible();
    await page.click('#rejectBtn');
    await expect(page.locator('#content')).toContainText(/No pending media/);
    expect(fs.existsSync(path.join(rejectedDir, filename))).toBeTruthy();
    expect(fs.existsSync(pendingPath)).toBeFalsy();
    await context.close();
  });
});
