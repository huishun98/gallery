import http from 'node:http';

export const waitForServer = (url: string, timeoutMs = 15000): Promise<void> =>
  new Promise((resolve, reject) => {
    const start = Date.now();
    const poll = () => {
      http
        .get(url, (res) => {
          res.resume();
          if (res.statusCode && res.statusCode >= 200 && res.statusCode < 500) {
            resolve();
            return;
          }
          if (Date.now() - start > timeoutMs) {
            reject(new Error(`timeout waiting for ${url}`));
            return;
          }
          setTimeout(poll, 250);
        })
        .on('error', () => {
          if (Date.now() - start > timeoutMs) {
            reject(new Error(`timeout waiting for ${url}`));
            return;
          }
          setTimeout(poll, 250);
        });
    };
    poll();
  });
