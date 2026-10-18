import tseslint from 'typescript-eslint';
import hooks from 'eslint-plugin-react-hooks';
export default [{ignores:['dist/**','node_modules/**']},{files:['src/**/*.ts','src/**/*.tsx'],languageOptions:{parser:tseslint.parser,parserOptions:{ecmaVersion:2022,sourceType:'module',ecmaFeatures:{jsx:true}}},plugins:{'react-hooks':hooks},rules:{'react-hooks/rules-of-hooks':'error'}}];
