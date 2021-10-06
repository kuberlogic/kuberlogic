import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { fakeAsync, TestBed, tick } from '@angular/core/testing';

import { environment } from '@environments/environment';
import { ServiceBackupConfigModel } from '@models/service-backup-config.model';
import { MessagesService } from '@services/messages.service';
import { MockMessageService } from '@testing/mock-messages-service';
import { BackupConfigService } from './backup-config.service';

const backupConfig: ServiceBackupConfigModel = {
    enabled: true,
    aws_access_key_id: 'aws_access_key_id',
    aws_secret_access_key: 'aws_secret_access_key',
    bucket: 'bucket',
    endpoint: 'endpoint',
    schedule: '0 12 * * *',
};

describe('BackupConfigService', () => {
    let service: BackupConfigService;
    let httpMock: HttpTestingController;

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule],
            providers: [
                { provide: MessagesService, useClass: MockMessageService },
            ],
        });

        httpMock = TestBed.inject(HttpTestingController);
        service = TestBed.inject(BackupConfigService);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('should return current backup as undefined when no backup loaded', (done) => {
        service.getCurrentBackupConfig().subscribe((result) => {
            expect(result).toEqual(undefined);
            done();
        });
    });

    it('should return backup by service id', (done) => {
        const serviceId = 'default:postgresql';
        service.getBackupConfig(serviceId).subscribe((result) => {
            expect(result).toEqual(backupConfig);
            done();
        });

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/backup-config`);
        req.flush(backupConfig);
        httpMock.verify();
    });

    it('should not make second request when backup by service id is already loaded', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        service.getBackupConfig(serviceId).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/backup-config`);
        req.flush(backupConfig);

        tick(10000);
        service.getBackupConfig(serviceId).subscribe();
        httpMock.expectNone(`${environment.apiUrl}/services/${serviceId}/backup-config`);

        httpMock.verify();
    }));

    it('should make second request when backup by service id loads new backup', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        const serviceId2 = 'default:postgresql2';
        service.getBackupConfig(serviceId).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/backup-config`);
        req.flush(backupConfig);

        tick(10000);
        service.getBackupConfig(serviceId2).subscribe();
        httpMock.expectOne(`${environment.apiUrl}/services/${serviceId2}/backup-config`);

        httpMock.verify();
    }));

    it('should edit backup and update current backup', (done) => {
        const serviceId = 'default:postgresql';
        service.editBackupConfig('default:postgresql', backupConfig).subscribe(() => {
            service.getCurrentBackupConfig().subscribe((result) => {
                expect(result).toEqual(backupConfig);
                done();
            });
        });
        const req = httpMock.expectOne(
            (request) =>
                request.method === 'PUT' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/backup-config`);
        req.flush(backupConfig);
    });

    it('should create backup', (done) => {
        const serviceId = 'default:postgresql';
        service.createBackupConfig('default:postgresql', backupConfig).subscribe((result) => {
            expect(result).toEqual(backupConfig);
            done();
        });
        const req = httpMock.expectOne(
            (request) =>
                request.method === 'POST' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/backup-config`);
        req.flush(backupConfig);
    });
});
