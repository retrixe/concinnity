const { join } = require('path')
const standard = ['standard-with-typescript', 'standard-react', 'standard-jsx']
module.exports = {
  env: {
    es6: true,
    node: true,
    browser: true
  },
  extends: ['plugin:react-hooks/recommended', 'plugin:react/recommended', ...standard],
  plugins: ['react', 'react-hooks', '@typescript-eslint'],
  parser: '@typescript-eslint/parser',
  parserOptions: {
    project: join(__dirname, 'tsconfig.json')
  },
  ignorePatterns: ['.eslintrc.js'],
  rules: {
    '@typescript-eslint/triple-slash-reference': 'off',
    '@typescript-eslint/consistent-type-imports': 'off',
    '@typescript-eslint/strict-boolean-expressions': 'off',
    '@typescript-eslint/no-confusing-void-expression': 'off',
    '@typescript-eslint/explicit-function-return-type': 'off',
    'react/no-unknown-property': ['error', { ignore: ['css'] }],
    // TODO: Remove these rules being switched off.
    '@typescript-eslint/promise-function-async': 'off',
    '@typescript-eslint/no-misused-promises': 'off',
  }
}
