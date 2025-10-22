export default {
    input: '../api/openapi.yml',
    output: 'src/api/generated',
    plugins: [
     {
        name: '@hey-api/client-fetch',
        runtimeConfigPath: './src/api/hey-api.ts', 
      },
    ], 
};