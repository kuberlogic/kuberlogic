import { EnvShape } from './env-shape';

declare const configApiUrl: string;
declare const monitoringConsoleUrl: string;

export const environment: EnvShape = {
    production: true,
    name: 'prod',
    apiUrl: configApiUrl,
    monitoringConsoleUrl,
};
