import path from 'node:path';

const WEB_ROOT = 'apps/web';
const quote = (files) => files.map((f) => JSON.stringify(f)).join(' ');

export default {
  // eslint precisa rodar com cwd em apps/web (senão o plugin do Next não acha
  // o diretório `app`). Usamos `next lint --file`, que aplica a config do Next
  // e ignora arquivos gerados como next-env.d.ts. Prettier roda a partir da raiz.
  'apps/web/**/*.{ts,tsx}': (files) => {
    const rel = files
      .map((f) => path.relative(WEB_ROOT, f))
      .filter((f) => f !== 'next-env.d.ts' && !f.startsWith('.next/'));
    const tasks = [`prettier --write ${quote(files)}`];
    if (rel.length > 0) {
      const fileFlags = rel.map((f) => `--file ${JSON.stringify(f)}`).join(' ');
      tasks.unshift(`pnpm --filter web exec next lint --fix ${fileFlags}`);
    }
    return tasks;
  },
  'apps/web/**/*.{json,css,md}': (files) => `prettier --write ${quote(files)}`,
  'apps/api/**/*.go': (files) => `gofmt -w ${quote(files)}`,
};
