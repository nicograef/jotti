import playwright from 'eslint-plugin-playwright'
import { defineConfig, globalIgnores } from 'eslint/config'
import tseslint from 'typescript-eslint'

export default defineConfig([
  globalIgnores(['playwright-report', 'test-results', 'blob-report']),
  {
    files: ['**/*.ts'],
    extends: [tseslint.configs.recommendedTypeChecked],
    languageOptions: {
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
  {
    files: ['tests/**/*.ts', 'support/**/*.ts'],
    extends: [playwright.configs['flat/recommended']],
    rules: {
      // The suite names its assertion helpers erwarte… and pruefe….
      'playwright/expect-expect': [
        'error',
        { assertFunctionPatterns: ['^(erwarte|pruefe)[A-Z]'] },
      ],
    },
  },
  {
    files: ['**/*.{js,mjs}'],
    extends: [tseslint.configs.recommended],
  },
])
