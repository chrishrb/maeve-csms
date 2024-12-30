import { defaultPlugins } from '@hey-api/openapi-ts';

export default {
  client: '@hey-api/client-axios',
  input: '../manager/api/api-spec.yaml',
  output: 'src/api',
  plugins: [
    ...defaultPlugins,
    '@tanstack/react-query', 
  ],
};
