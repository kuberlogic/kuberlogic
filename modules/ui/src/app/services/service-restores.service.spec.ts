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
