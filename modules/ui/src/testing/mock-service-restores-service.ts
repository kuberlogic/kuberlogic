import { ServiceRestoreModel } from '@models/service-restore.model';
import { Observable, of } from 'rxjs';

export const restoreItem: ServiceRestoreModel = {
    file: 's3://test/postgresql/kuberlogic-kl-pg/logical_backups/1622729871.sql.gz',
    database: 'db1',
    time: '2021-06-04T08:00:14.000Z',
    status: 'Failed',
};

export class MockServiceRestoresService {
    getList(_serviceId: string): Observable<ServiceRestoreModel[] | undefined> {
        return of([restoreItem]);
    }

    restore(serviceId: string, key: string, database: string): Observable<ServiceRestoreModel> {
        return of(restoreItem);
    }
}
