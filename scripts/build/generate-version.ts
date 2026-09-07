#!/usr/bin/env bun
import { resolve } from 'node:path';

import { commandOutput } from './command';
import { normalizeVersion } from './utils';

function main() {
  try {
    const projectDir = resolve(__dirname, '../..');

    const rawVersion =
      process.env.VERSION || commandOutput('git', ['describe', '--tags', '--always', '--dirty'], projectDir) || 'dev';

    console.error(`Raw version from git: ${rawVersion}`);

    const devVersion = normalizeVersion(rawVersion);

    console.error(`Generated dev version: ${devVersion}`);
    console.log(devVersion);
  } catch (error) {
    console.error('Error generating version:', error);
    process.exit(1);
  }
}

main();
