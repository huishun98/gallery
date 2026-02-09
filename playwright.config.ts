import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: 'e2e',
  timeout: 60 * 1000,
  use: {
    baseURL: 'http://127.0.0.1:8025',
    trace: 'on-first-retry',
  },
  webServer: {
    command: 'tsx scripts/e2e-server.ts',
    url: 'http://127.0.0.1:8025/slideshow',
    timeout: 120 * 1000,
    reuseExistingServer: false,
    env: {
      GALLERY_PORT: '8025',
      GALLERY_DANMU_ENABLED: '1',
      GALLERY_APPROVALS_ENABLED: '0',
      GALLERY_UPLOADS_ENABLED: '1',
      GALLERY_WORKDIR: '.e2e-workdir/default',
    },
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
