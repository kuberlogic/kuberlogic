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
