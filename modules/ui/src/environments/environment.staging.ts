import { EnvShape } from './env-shape';

declare const configApiUrl: string;
declare const monitoringConsoleUrl: string;

export const environment: EnvShape = {
    production: true,
    name: 'staging',
    apiUrl: configApiUrl,
    monitoringConsoleUrl,
};
