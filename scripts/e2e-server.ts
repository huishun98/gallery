import { spawn } from 'node:child_process';
import { existsSync, mkdirSync, writeFileSync } from 'node:fs';
import { join, resolve } from 'node:path';

const port = process.env.GALLERY_PORT ?? '8000';
const danmuEnabled = process.env.GALLERY_DANMU_ENABLED !== '0';
const approvalsEnabled = process.env.GALLERY_APPROVALS_ENABLED === '1';
const uploadsEnabled = process.env.GALLERY_UPLOADS_ENABLED !== '0';
const repoRoot = resolve(process.cwd());
const workDir = process.env.GALLERY_WORKDIR ?? join(repoRoot, '.e2e-workdir', 'default');
const dataDir = join(workDir, '.data');
if (!existsSync(workDir)) {
  mkdirSync(workDir, { recursive: true });
}
if (!existsSync(dataDir)) {
  mkdirSync(dataDir, { recursive: true });
}
const mediaDir = join(dataDir, 'media');
const mediaFilesDir = join(mediaDir, 'media');
const pendingDir = join(mediaDir, 'pending');
if (!existsSync(mediaFilesDir)) {
  mkdirSync(mediaFilesDir, { recursive: true });
}
if (!existsSync(pendingDir)) {
  mkdirSync(pendingDir, { recursive: true });
}

const settingsPath = join(dataDir, 'settings.json');
const settings = {
  data_directory: dataDir,
  admin: { admin: 'admin' },
  port,
  danmu_enabled: danmuEnabled,
  approvals_enabled: approvalsEnabled,
  uploads_enabled: uploadsEnabled,
};
writeFileSync(settingsPath, JSON.stringify(settings, null, 2));

const child = spawn('go', ['run', repoRoot], {
  stdio: 'inherit',
  env: { ...process.env, DISABLE_TUNNEL: '1' },
  cwd: workDir,
});

const shutdown = () => {
  if (!child.killed) {
    child.kill('SIGTERM');
  }
};

process.on('SIGTERM', shutdown);
process.on('SIGINT', shutdown);

child.on('exit', (code) => {
  process.exit(code ?? 0);
});
