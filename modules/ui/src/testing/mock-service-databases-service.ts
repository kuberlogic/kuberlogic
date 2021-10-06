import { ServiceDatabaseModel } from '@models/service-database.model';
import { Observable, of } from 'rxjs';

export class MockServiceDatabasesService {
    databases: ServiceDatabaseModel[] = [{ name: 'db1'}, { name: 'db2'}];

    getDatabases(_serviceId: string): Observable<ServiceDatabaseModel[] | undefined> {
        return of(this.databases);
    }

    createDatabase(_serviceId: string, db: ServiceDatabaseModel): Observable<ServiceDatabaseModel> {
        return of(db);
    }

    deleteDatabase(_serviceId: string, _name: string): Observable<ServiceDatabaseModel> {
        return of(this.databases[0]);
    }
}
