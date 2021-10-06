export enum ServiceModelType {
    POSTGRES = 'postgresql',
    MYSQL = 'mysql',
}

export enum ServiceModelStatus {
    PROCESSING = 'Processing',
    RUNNING = 'Running',
    READY = 'Ready',
    FAILED = 'Failed',
    UNKNOWN = 'Unknown',
    NOT_READY = 'NotReady'
}

export interface ServiceModelNode {
    host: string;
    port: number;
    user: string;
    password: string;
    ssl_mode: string;
    cert: string;
}

export interface ServiceModelConnection {
    master: ServiceModelNode;
    replica: ServiceModelNode;
}

export interface ServiceModelMaintenanceWindow {
    day: string;
    startHour: number;
}

export interface ServiceModelLimit {
    cpu: string;
    memory: string;
    volumeSize: string;
}

export interface ServiceModelInstance {
    name: string;
    role: string;
    status: {
        status: ServiceModelStatus
    };
}

export interface ServiceModel {
    type: ServiceModelType;
    name: string;
    ns?: string;
    internalConnection?: ServiceModelConnection;
    externalConnection?: ServiceModelConnection;
    maintenanceWindow?: ServiceModelMaintenanceWindow;
    instances?: ServiceModelInstance[];
    status: ServiceModelStatus;
    masters: number;
    replicas: number;
    created_at: string;
    version?: string;
    limits?: ServiceModelLimit;
    advancedConf?: {[advancedConfProp: string]: string};
}
