import { test, expect } from '@playwright/test';
import { spawn } from 'node:child_process';
import net from 'node:net';
import path from 'node:path';

import { waitForServer } from './helpers/waitForServer';

test('admin comments requires auth', async ({ request }) => {
  const res = await request.get('/admin/comments');
  expect(res.status()).toBe(401);
});

test('admin comments renders with auth', async ({ browser }) => {
  const context = await browser.newContext({
    baseURL: 'http://127.0.0.1:8000',
    httpCredentials: { username: 'admin', password: 'admin' },
  });
  const page = await context.newPage();
  await page.goto('/admin/comments');
  await expect(page).toHaveTitle(/Delete Comments/);
  await expect(page.locator('#downloadCsvBtn')).toBeVisible();
  await context.close();
});

test('admin comments download and delete work', async ({ request }) => {
  const filename = `comment-${Date.now()}.jpg`;
  const createRes = await request.post('/comment', {
    form: { filename, comment: 'hello from e2e' },
  });
  expect(createRes.status()).toBe(200);
  const created = await createRes.json();
  const id = String(created.id);

  const authHeader = {
    Authorization: `Basic ${Buffer.from('admin:admin').toString('base64')}`,
  };

  const downloadRes = await request.get('/admin/comments/download', {
    headers: authHeader,
  });
  expect(downloadRes.status()).toBe(200);
  expect(downloadRes.headers()['content-type']).toContain('text/csv');
  const csvBody = await downloadRes.text();
  expect(csvBody).toContain(filename);

  const deleteRes = await request.post('/admin/comments/delete', {
    headers: authHeader,
    form: { id },
  });
  expect(deleteRes.status()).toBe(200);

  const listRes = await request.get(`/comments?filename=${encodeURIComponent(filename)}`);
  expect(listRes.status()).toBe(200);
  const listBody = await listRes.json();
  expect(listBody.total).toBe(0);
});

test('admin comments not available when danmu disabled', async ({ request }) => {
  const port = await new Promise<number>((resolve, reject) => {
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
  const baseURL = `http://127.0.0.1:${port}`;

  const serverProcess = spawn('tsx', ['scripts/e2e-server.ts'], {
    stdio: 'inherit',
      env: {
        ...process.env,
        GALLERY_PORT: String(port),
        GALLERY_DANMU_ENABLED: '0',
        GALLERY_APPROVALS_ENABLED: '0',
        GALLERY_WORKDIR: path.join(process.cwd(), '.e2e-workdir', 'danmu-off-comments'),
      },
    });
  await waitForServer(`${baseURL}/slideshow`);

  const res = await request.get(`${baseURL}/admin/comments`);
  expect(res.status()).toBe(404);

  if (!serverProcess.killed) {
    serverProcess.kill('SIGTERM');
  }
});
