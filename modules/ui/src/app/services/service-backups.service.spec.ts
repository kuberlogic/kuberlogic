import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { fakeAsync, TestBed, tick } from '@angular/core/testing';
import { ServiceBackupModel } from '@models/service-backup.model';

import { environment } from '@environments/environment';
import { MessagesService } from '@services/messages.service';
import { MockMessageService } from '@testing/mock-messages-service';
import { backupItem } from '@testing/mock-service-backups-service';
import { ServiceBackupsService } from './service-backups.service';

const backups: ServiceBackupModel[] = [backupItem];

describe('ServiceBackupsService', () => {
    let service: ServiceBackupsService;
    let httpMock: HttpTestingController;

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule],
            providers: [
                { provide: MessagesService, useClass: MockMessageService },
            ],
        });

        httpMock = TestBed.inject(HttpTestingController);
        service = TestBed.inject(ServiceBackupsService);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('should return backups by service id', (done) => {
        const serviceId = 'default:postgresql';
        service.getList(serviceId).subscribe((result) => {
            expect(result).toEqual(backups);
            done();
        });

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/backups`);
        req.flush(backups);
        httpMock.verify();
    });

    it('should not make second request when backups are already loaded', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        service.getList(serviceId).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/backups`);
        req.flush(backups);

        tick(10000);
        service.getList(serviceId).subscribe();
        httpMock.expectNone(`${environment.apiUrl}/services/${serviceId}/backups`);

        httpMock.verify();
    }));

    it('should make second request when backups are loaded for another service', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        const serviceId2 = 'default:postgresql2';
        service.getList(serviceId).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/backups`);
        req.flush(backups);

        tick(10000);
        service.getList(serviceId2).subscribe();
        httpMock.expectOne(`${environment.apiUrl}/services/${serviceId2}/backups`);

        httpMock.verify();
    }));

});
