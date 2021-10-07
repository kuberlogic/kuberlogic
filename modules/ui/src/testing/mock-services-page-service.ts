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

import { ServiceModel, ServiceModelStatus, ServiceModelType } from '@models/service.model';
import { ServiceInstancesNames } from '@services/services-page.service';
import { Observable, of } from 'rxjs';

const serviceModels: ServiceModel[] = [
    {
        type: ServiceModelType.MYSQL,
        name: 'MySql prod',
        status: ServiceModelStatus.RUNNING,
        masters: 1,
        replicas: 2,
        created_at: '2021-02-01T10:56:12.115Z',
        internalConnection: {
            master: {
                host: 'pg-master-cloudlinux-a892.domain.com',
                port: 20990,
                user: 'cloudmanaged_admin',
                password: '*********',
                ssl_mode: 'require',
                cert: '/cert_link'
            },
            replica: {
                host: 'pg-master-cloudlinux-a892.domain.com',
                port: 20990,
                user: 'cloudmanaged_admin',
                password: '*********',
                ssl_mode: 'require',
                cert: '/cert_link'
            },
        },
        externalConnection: {
            master: {
                host: 'external-pg-master-cloudlinux-a892.domain.com',
                port: 20990,
                user: 'cloudmanaged_admin',
                password: '*********',
                ssl_mode: 'require',
                cert: '/cert_link'
            },
            replica: {
                host: 'external-pg-master-cloudlinux-a892.domain.com',
                port: 20990,
                user: 'cloudmanaged_admin',
                password: '*********',
                ssl_mode: 'require',
                cert: '/cert_link'
            },
        },
        advancedConf: {}
    },
    {
        type: ServiceModelType.POSTGRES,
        name: 'PostgreSql prod',
        status: ServiceModelStatus.FAILED,
        masters: 2,
        replicas: 5,
        created_at: '2021-02-09T10:56:12.115Z',
        internalConnection: {
            master: {
                host: 'pg-master-cloudlinux-a892.domain.com',
                port: 20990,
                user: 'cloudmanaged_admin',
                password: '*********',
                ssl_mode: 'require',
                cert: '/cert_link'
            },
            replica: {
                host: 'pg-master-cloudlinux-a892.domain.com',
                port: 20990,
                user: 'cloudmanaged_admin',
                password: '*********',
                ssl_mode: 'require',
                cert: '/cert_link'
            },
        },
        externalConnection: {
            master: {
                host: 'external-pg-master-cloudlinux-a892.domain.com',
                port: 20990,
                user: 'cloudmanaged_admin',
                password: '*********',
                ssl_mode: 'require',
                cert: '/cert_link'
            },
            replica: {
                host: 'external-pg-master-cloudlinux-a892.domain.com',
                port: 20990,
                user: 'cloudmanaged_admin',
                password: '*********',
                ssl_mode: 'require',
                cert: '/cert_link'
            },
        },
        advancedConf: {}
    },

];

export class MockServicesPageService {
    currentServiceId = '123';
    getServicesList(): Observable<ServiceModel[]> {
        return of(serviceModels);
    }

    createService(serviceModel: Partial<ServiceModel>): Observable<ServiceModel> {
        return of({...serviceModel} as ServiceModel);
    }

    editService(serviceModel: Partial<ServiceModel>): Observable<ServiceModel> {
        return of({...serviceModel} as ServiceModel);
    }

    getCurrentServiceId(): string {
        return this.currentServiceId;
    }

    deleteService(_serviceModel: ServiceModel): Observable<void> {
        return (of(undefined));
    }

    getService(serviceId: string): Observable<ServiceModel | undefined> {
        return of(serviceModels[0]);
    }

    getCurrentService(): Observable<ServiceModel | undefined> {
        return of(serviceModels[0]);
    }

    getCurrentServiceInstancesNames(): Observable<ServiceInstancesNames | undefined> {
        const instances = new Map<string, string>();
        instances.set('key_instance', `key_instance (some_role)`);
        return of(instances);
    }
}
