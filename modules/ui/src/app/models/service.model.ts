/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
