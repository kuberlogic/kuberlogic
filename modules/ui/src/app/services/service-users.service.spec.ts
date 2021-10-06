import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { fakeAsync, TestBed, tick } from '@angular/core/testing';
import { environment } from '@environments/environment';
import { ServiceUserModel } from '@models/service-user.model';

import { MessagesService } from '@services/messages.service';
import { MockMessageService } from '@testing/mock-messages-service';
import { ServiceUsersService } from './service-users.service';

const users: ServiceUserModel[] = [
    {name: 'user1', password: 'p@$$w0rd'},
    {name: 'user2', password: 'p@$$w0rd'},
];

describe('ServiceUsersService', () => {
    let service: ServiceUsersService;
    let httpMock: HttpTestingController;

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule],
            providers: [
                { provide: MessagesService, useClass: MockMessageService },
            ]
        });

        httpMock = TestBed.inject(HttpTestingController);
        service = TestBed.inject(ServiceUsersService);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('should return users by service id', (done) => {
        const serviceId = 'default:postgresql';
        service.getUsers(serviceId).subscribe((result) => {
            expect(result).toEqual(users);
            done();
        });

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/users`);
        req.flush(users);
        httpMock.verify();
    });

    it('should not make second request when users are already loaded', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        service.getUsers(serviceId).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/users`);
        req.flush(users);

        tick(10000);
        service.getUsers(serviceId).subscribe();
        httpMock.expectNone(`${environment.apiUrl}/services/${serviceId}/users`);

        httpMock.verify();
    }));

    it('should make second request when users are loaded for another service', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        const serviceId2 = 'default:postgresql2';
        service.getUsers(serviceId).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'GET' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/users`);
        req.flush(users);

        tick(10000);
        service.getUsers(serviceId2).subscribe();
        httpMock.expectOne(`${environment.apiUrl}/services/${serviceId2}/users`);

        httpMock.verify();
    }));

    it('should create user and reload users', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        service.createUser('default:postgresql', users[0]).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'POST' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/users`);
        req.flush(users[0]);

        tick(10000);
        service.getUsers(serviceId).subscribe();
        httpMock.expectOne(`${environment.apiUrl}/services/${serviceId}/users`);

        httpMock.verify();
    }));

    it('should edit user and reload users', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        service.editUser('default:postgresql', users[0]).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'PUT' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/users/${users[0].name}`);
        req.flush(users[0]);

        tick(10000);
        service.getUsers(serviceId).subscribe();
        httpMock.expectOne(`${environment.apiUrl}/services/${serviceId}/users`);

        httpMock.verify();
    }));

    it('should delete user and reload users', fakeAsync(() => {
        const serviceId = 'default:postgresql';
        service.deleteUser('default:postgresql', users[0].name).subscribe();

        const req = httpMock.expectOne(
            (request) =>
                request.method === 'DELETE' &&
                request.url === `${environment.apiUrl}/services/${serviceId}/users/${users[0].name}`);
        req.flush(users[0]);

        tick(10000);
        service.getUsers(serviceId).subscribe();
        httpMock.expectOne(`${environment.apiUrl}/services/${serviceId}/users`);

        httpMock.verify();
    }));
});
