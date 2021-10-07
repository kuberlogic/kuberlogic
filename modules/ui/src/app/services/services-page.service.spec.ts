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

import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { fakeAsync, TestBed, tick } from '@angular/core/testing';
import { environment } from '@environments/environment.test';
import { ServiceModel, ServiceModelStatus, ServiceModelType } from '@models/service.model';

import { MessagesService } from '@services/messages.service';
import { MockMessageService } from '@testing/mock-messages-service';
import { ServicesPageService } from './services-page.service';

const serviceModels: ServiceModel[] = [
    {
        type: ServiceModelType.MYSQL,
        name: 'MySql prod',
        status: ServiceModelStatus.RUNNING,
        masters: 1,
        replicas: 2,
        created_at: '2021-02-01T10:56:12.115Z',
    },
    {
        type: ServiceModelType.POSTGRES,
        name: 'PostgreSql prod',
        status: ServiceModelStatus.FAILED,
        masters: 2,
        replicas: 5,
        created_at: '2021-02-09T10:56:12.115Z',
    },
];

const serviceModel: ServiceModel = {
    type: ServiceModelType.POSTGRES,
    name: 'PostgreSql prod',
    status: ServiceModelStatus.FAILED,
    masters: 2,
    replicas: 5,
    created_at: '2021-02-09T10:56:12.115Z',
    instances: [{
        name: 'inst1',
        role: 'master',
        status: {
            status: ServiceModelStatus.RUNNING
        },
    }],
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
    maintenanceWindow: {
        day: 'Monday',
        startHour: 0
    },
    limits: {
        cpu: '500m',
        memory: '1Gi',
        volumeSize: '4Gi'
    },
    advancedConf: {
        test_setting: '10'
    }
};

describe('ServicesPageService', () => {
    let service: ServicesPageService;
    let httpMock: HttpTestingController;

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule],
            providers: [
                { provide: MessagesService, useClass: MockMessageService },
            ],
        });
        service = TestBed.inject(ServicesPageService);
        httpMock = TestBed.inject(HttpTestingController);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('should return 2 services', (done) => {
        service.getServicesList().subscribe((services) => {
            if (!services) {
                throw Error('Empty services');
            }
            expect(services.length).toBe(2);
            expect(services[0].type).toBe(ServiceModelType.MYSQL);
            expect(services[1].type).toBe(ServiceModelType.POSTGRES);
            done();
        });
        const req = httpMock.expectOne((request) => request.url === `${environment.apiUrl}/services`);
        req.flush(serviceModels);
    });

    it('should return current service as undefined when no service loaded', (done) => {
        service.getCurrentService().subscribe((result) => {
            expect(result).toEqual(undefined);
            done();
        });
    });

    it('should return service by id', (done) => {
        const serviceId = 'default:postgresql';
        service.getService(serviceId).subscribe((result) => {
            expect(result).toEqual(serviceModel);
            done();
        });

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}`);
        req.flush(serviceModel);
        httpMock.verify();
    });

    it('should return current service id', (done) => {
        const serviceId = 'default:postgresql';
        service.getService(serviceId).subscribe(() => {
            expect(service.getCurrentServiceId()).toEqual(serviceId);
            service.getCurrentServiceInstancesNames().subscribe((data) => {
                expect(data?.size).toBe(1);
                expect(data?.get('inst1')).toBe('inst1 (master)');
            });
            done();
        });

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}`);
        req.flush(serviceModel);
        httpMock.verify();
    });

    it('should not make second request when service by id is already loaded', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        service.getService(serviceId).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}`);
        req.flush(serviceModel);

        tick(10000);
        service.getService(serviceId).subscribe();
        httpMock.expectNone(`${environment.apiUrl}/services/${serviceId}`);

        httpMock.verify();
    }));

    it('should make second request when service by id loads new service', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        const serviceId2 = 'default:postgresql2';
        service.getService(serviceId).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}`);
        req.flush(serviceModel);

        tick(10000);
        service.getService(serviceId2).subscribe();
        httpMock.expectOne(`${environment.apiUrl}/services/${serviceId2}`);

        httpMock.verify();
    }));

    it('should edit service and update current service', (done) => {
        const serviceId = 'default:postgresql';
        service.editService('default:postgresql', serviceModel).subscribe(() => {
            service.getCurrentService().subscribe((result) => {
                expect(result).toEqual(serviceModel);
                done();
            });
        });
        const req = httpMock.expectOne(
            (request) =>
                request.method === 'PUT' &&
                request.url === `${environment.apiUrl}/services/${serviceId}`);
        req.flush(serviceModel);
    });

    it('should create service', (done) => {
        service.createService(serviceModel).subscribe((result) => {
            expect(result).toEqual(serviceModel);
            done();
        });
        const req = httpMock.expectOne(
            (request) =>
                request.method === 'POST' &&
                request.url === `${environment.apiUrl}/services`);
        req.flush(serviceModel);
    });
});
