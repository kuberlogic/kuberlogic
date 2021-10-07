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
import { TestBed } from '@angular/core/testing';
import { environment } from '@environments/environment';
import { ServiceLogModel } from '@models/service-log.model';
import { ServiceLogsService } from '@services/service-logs.service';

const logs: ServiceLogModel = {lines: 100, body: 'sdfgsadf'};

describe('ServiceLogsService', () => {
    let service: ServiceLogsService;
    let httpMock: HttpTestingController;

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule]
        });

        httpMock = TestBed.inject(HttpTestingController);
        service = TestBed.inject(ServiceLogsService);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('should return logs by service id and instance', (done) => {
        const serviceId = 'default:postgresql';
        const serviceInstance = 'some_instance';
        service.get(serviceId, serviceInstance).subscribe((result) => {
            expect(result).toEqual(logs);
            done();
        });

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/logs`);
        req.flush(logs);
        httpMock.verify();
    });
});
