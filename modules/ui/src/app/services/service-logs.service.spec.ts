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
