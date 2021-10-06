import { ServiceBackupConfigModel } from '@models/service-backup-config.model';
import { Observable, of } from 'rxjs';

const backupConfig: ServiceBackupConfigModel = {
    enabled: true,
    aws_access_key_id: 'aws_access_key_id',
    aws_secret_access_key: 'aws_secret_access_key',
    bucket: 'bucket',
    endpoint: 'endpoint',
    schedule: '0 8 * * *',
};

export class MockBackupConfigService {
    getBackupConfig(_serviceId: string): Observable<ServiceBackupConfigModel | undefined> {
        return of(backupConfig);
    }

    getCurrentBackupConfig(): Observable<ServiceBackupConfigModel | undefined> {
        return of(backupConfig);
    }

    createBackupConfig(_serviceId: string, config: ServiceBackupConfigModel): Observable<ServiceBackupConfigModel> {
        return of(config);
    }

    editBackupConfig(_serviceId: string, config: ServiceBackupConfigModel): Observable<ServiceBackupConfigModel> {
        return of(config);
    }
}
