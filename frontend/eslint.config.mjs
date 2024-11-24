import love from 'eslint-config-love'
import standardJsx from 'eslint-config-standard-jsx'
import standardReact from 'eslint-config-standard-react'
import react from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'
import eslintPluginPrettierRecommended from 'eslint-plugin-prettier/recommended'

export default [
  {
    ignores: ['.pnp.cjs', '.pnp.loader.mjs', '.yarn'],
  },
  {
    ...love,
    files: [
      'imports/**/*.js',
      'imports/**/*.ts',
      'imports/**/*.tsx',
      'pages/**/*.js',
      'pages/**/*.ts',
      'pages/**/*.tsx',
    ],
  },
  react.configs.flat.recommended,
  { plugins: { 'react-hooks': reactHooks }, rules: reactHooks.configs.recommended.rules },
  { rules: standardJsx.rules },
  { rules: standardReact.rules },
  {
    rules: {
      'react/no-unknown-property': ['error', { ignore: ['css'] }],
      // Make TypeScript ESLint less strict. Perhaps move to ESLint+TS-ESLint+import+promise+n?
      '@typescript-eslint/no-confusing-void-expression': 'off',
      '@typescript-eslint/strict-boolean-expressions': 'off',
      '@typescript-eslint/restrict-plus-operands': 'off',
      '@typescript-eslint/no-dynamic-delete': 'off',
      '@typescript-eslint/no-var-requires': 'off',

      '@typescript-eslint/max-params': 'off',
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-magic-numbers': 'off',
      '@typescript-eslint/no-unsafe-argument': 'off',
      '@typescript-eslint/no-unsafe-assignment': 'off',
      '@typescript-eslint/no-unsafe-member-access': 'off',
      '@typescript-eslint/no-unnecessary-condition': 'off',
      '@typescript-eslint/no-unsafe-type-assertion': 'off',
      '@typescript-eslint/class-methods-use-this': 'off',
      '@typescript-eslint/prefer-destructuring': 'off',
      '@typescript-eslint/use-unknown-in-catch-callback-variable': 'off',
    },
  },
  eslintPluginPrettierRecommended,
]
