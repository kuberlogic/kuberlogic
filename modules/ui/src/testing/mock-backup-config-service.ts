/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
