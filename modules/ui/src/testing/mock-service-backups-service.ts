import { ServiceBackupModel } from '@models/service-backup.model';
import { Observable, of } from 'rxjs';

export const backupItem: ServiceBackupModel = {
    file: 's3://test/postgresql/kuberlogic-kl-pg/logical_backups/1622729871.sql.gz',
    size: 14021,
    lastModified: '2021-06-03T14:17:55.510Z',
};

export class MockServiceBackupsService {
    getList(_serviceId: string): Observable<ServiceBackupModel[] | undefined> {
        return of([backupItem]);
    }
}
