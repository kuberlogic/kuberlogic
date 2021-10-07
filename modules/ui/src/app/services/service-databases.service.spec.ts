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
import { environment } from '@environments/environment';
import { ServiceDatabaseModel } from '@models/service-database.model';

import { MessagesService } from '@services/messages.service';
import { MockMessageService } from '@testing/mock-messages-service';
import { ServiceDatabasesService } from './service-databases.service';

const databases: ServiceDatabaseModel[] = [
    {name: 'db1'},
    {name: 'db2'},
];

describe('ServiceDatabasesService', () => {
    let service: ServiceDatabasesService;
    let httpMock: HttpTestingController;

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule],
            providers: [
                { provide: MessagesService, useClass: MockMessageService },
            ],
        });

        httpMock = TestBed.inject(HttpTestingController);
        service = TestBed.inject(ServiceDatabasesService);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('should return databases by service id', (done) => {
        const serviceId = 'default:postgresql';
        service.getDatabases(serviceId).subscribe((result) => {
            expect(result).toEqual(databases);
            done();
        });

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/databases`);
        req.flush(databases);
        httpMock.verify();
    });

    it('should not make second request when databases are already loaded', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        service.getDatabases(serviceId).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/databases`);
        req.flush(databases);

        tick(10000);
        service.getDatabases(serviceId).subscribe();
        httpMock.expectNone(`${environment.apiUrl}/services/${serviceId}/databases`);

        httpMock.verify();
    }));

    it('should make second request when databases are loaded for another service', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        const serviceId2 = 'default:postgresql2';
        service.getDatabases(serviceId).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/databases`);
        req.flush(databases);

        tick(10000);
        service.getDatabases(serviceId2).subscribe();
        httpMock.expectOne(`${environment.apiUrl}/services/${serviceId2}/databases`);

        httpMock.verify();
    }));

    it('should create databases and reload databases', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        service.createDatabase('default:postgresql', databases[0]).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'POST' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/databases`);
        req.flush(databases[0]);

        tick(10000);
        service.getDatabases(serviceId).subscribe();
        httpMock.expectOne(`${environment.apiUrl}/services/${serviceId}/databases`);

        httpMock.verify();
    }));

    it('should delete database and reload databases', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        service.deleteDatabase('default:postgresql', databases[0].name).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'DELETE' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/databases/${databases[0].name}`);
        req.flush(databases[0]);

        tick(10000);
        service.getDatabases(serviceId).subscribe();
        httpMock.expectOne(`${environment.apiUrl}/services/${serviceId}/databases`);

        httpMock.verify();
    }));
});
