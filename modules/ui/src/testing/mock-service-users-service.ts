import { ServiceUserModel } from '@models/service-user.model';
import { Observable, of } from 'rxjs';

export class MockServiceUsersService {
    users: ServiceUserModel[] =  [
        {
            name: 'postgres1',
            password: 'p@$$w0rd',
        },
        {
            name: 'postgres2',
            password: 'p@$$w0rd',
        }
    ];

    getUsers(_serviceId: string): Observable<ServiceUserModel[] | undefined> {
        return of(this.users);
    }

    createUser(_serviceId: string, user: ServiceUserModel): Observable<ServiceUserModel> {
        return of(user);
    }

    editUser(_serviceId: string, user: ServiceUserModel): Observable<ServiceUserModel> {
        return of(user);
    }

    deleteUser(_serviceId: string, _name: string): Observable<ServiceUserModel> {
        return of(this.users[0]);
    }
}
