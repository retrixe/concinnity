import js from '@eslint/js'
import tseslint from 'typescript-eslint'
import sveltePlugin from 'eslint-plugin-svelte'
import * as svelteParser from 'svelte-eslint-parser'
import importPlugin from 'eslint-plugin-import'
import pluginPromise from 'eslint-plugin-promise'
import nodePlugin from 'eslint-plugin-n'
import eslintPluginPrettierRecommended from 'eslint-plugin-prettier/recommended'
import globals from 'globals'
import svelteConfig from './svelte.config.js'

export default tseslint.config(
  {
    ignores: [
      '.pnp.cjs',
      '.pnp.loader.mjs',
      '.yarn',
      '.svelte-kit',
      '.prettierrc.cjs',
      '*.config.{mjs,js}',
    ],
  },
  js.configs.recommended,
  tseslint.configs.strictTypeChecked,
  tseslint.configs.stylisticTypeChecked,
  pluginPromise.configs['flat/recommended'],
  importPlugin.flatConfigs.recommended, // Could use TypeScript resolver
  nodePlugin.configs['flat/recommended-module'],
  ...sveltePlugin.configs.recommended,
  ...sveltePlugin.configs.prettier,
  {
    files: ['**/*.svelte', '**/*.svelte.js', '**/*.svelte.ts'],
    languageOptions: {
      parser: svelteParser,
      parserOptions: {
        projectService: true,
        parser: tseslint.parser,
        extraFileExtensions: ['.svelte'],
        svelteConfig,
      },
    },
  },
  eslintPluginPrettierRecommended,
  {
    languageOptions: {
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
      globals: {
        ...globals.browser,
      },
    },
    rules: {
      '@typescript-eslint/no-confusing-void-expression': 'off',
      '@typescript-eslint/no-import-type-side-effects': ['error'],
      '@typescript-eslint/consistent-type-imports': [
        'error',
        {
          prefer: 'type-imports',
          disallowTypeAnnotations: true,
          fixStyle: 'inline-type-imports',
        },
      ],
      '@typescript-eslint/restrict-template-expressions': [
        'error',
        {
          allowAny: false,
          allowBoolean: false,
          allowNumber: true,
          allowRegExp: false,
          allowNever: false,
        },
      ],
      'promise/always-return': ['error', { ignoreLastCallback: true }],
      'n/no-missing-import': 'off',
      'n/no-unsupported-features/node-builtins': 'off',
      'n/no-unsupported-features/es-syntax': 'off',
      'import/no-unresolved': 'off',
    },
  },
)
