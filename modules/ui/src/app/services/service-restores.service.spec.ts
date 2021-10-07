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
import { ServiceRestoreModel } from '@models/service-restore.model';

import { environment } from '@environments/environment';
import { MessagesService } from '@services/messages.service';
import { MockMessageService } from '@testing/mock-messages-service';
import { restoreItem } from '@testing/mock-service-restores-service';
import { ServiceRestoresService } from './service-restores.service';

const restores: ServiceRestoreModel[] = [restoreItem];

describe('ServiceRestoresService', () => {
    let service: ServiceRestoresService;
    let httpMock: HttpTestingController;

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule],
            providers: [
                { provide: MessagesService, useClass: MockMessageService },
            ],
        });

        httpMock = TestBed.inject(HttpTestingController);
        service = TestBed.inject(ServiceRestoresService);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('should return restores by service id', (done) => {
        const serviceId = 'default:postgresql';
        service.getList(serviceId).subscribe((result) => {
            expect(result).toEqual(restores);
            done();
        });

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/restores`);
        req.flush(restores);
        httpMock.verify();
    });

    it('should not make second request when restores are already loaded', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        service.getList(serviceId).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/restores`);
        req.flush(restores);

        tick(10000);
        service.getList(serviceId).subscribe();
        httpMock.expectNone(`${environment.apiUrl}/services/${serviceId}/restores`);

        httpMock.verify();
    }));

    it('should make second request when restores are loaded for another service', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        const serviceId2 = 'default:postgresql2';
        service.getList(serviceId).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/restores`);
        req.flush(restores);

        tick(10000);
        service.getList(serviceId2).subscribe();
        httpMock.expectOne(`${environment.apiUrl}/services/${serviceId2}/restores`);

        httpMock.verify();
    }));

    it('should restore databases and reload restores', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        service.restore('default:postgresql', 'dfg', restoreItem.database).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'POST' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/restores`);
        req.flush(restoreItem);

        tick(10000);
        service.getList(serviceId).subscribe();
        httpMock.expectOne(`${environment.apiUrl}/services/${serviceId}/restores`);

        httpMock.verify();
    }));
});
