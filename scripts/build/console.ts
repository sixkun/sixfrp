export const colors = {
  red: "\x1b[0;31m",
  green: "\x1b[0;32m",
  yellow: "\x1b[1;33m",
  blue: "\x1b[0;34m",
  reset: "\x1b[0m",
};

export function printHeader(text: string) {
  console.log(`${colors.yellow}========================================${colors.reset}`);
  console.log(`${colors.yellow}${text}${colors.reset}`);
  console.log(`${colors.yellow}========================================${colors.reset}`);
}

export function log(message: string) {
  console.log(`${colors.green}${message}${colors.reset}`);
}

export function success(message: string) {
  console.log(`${colors.green}[OK] ${message}${colors.reset}`);
}

export function fail(message: string) {
  console.log(`${colors.red}[FAIL] ${message}${colors.reset}`);
}

export function warn(message: string) {
  console.log(`${colors.yellow}${message}${colors.reset}`);
}

export function info(message: string) {
  console.log(`${colors.blue}${message}${colors.reset}`);
}
