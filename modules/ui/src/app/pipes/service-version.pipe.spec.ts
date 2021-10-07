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

import { ServiceModel, ServiceModelStatus, ServiceModelType } from '@models/service.model';
import { ServiceVersionPipe } from './service-version.pipe';

const serviceModel: ServiceModel = {
    type: ServiceModelType.POSTGRES,
    name: 'PostgreSql prod',
    status: ServiceModelStatus.FAILED,
    masters: 2,
    replicas: 5,
    created_at: '2021-02-09T10:56:12.115Z',
};

describe('ServiceVersionPipe', () => {
    const pipe = new ServiceVersionPipe();

    it('create an instance', () => {
        expect(pipe).toBeTruthy();
    });

    it('should return "PostgreSQL 12.1.5" when type is postgre', () => {
        const result = pipe.transform(serviceModel);
        expect(result).toEqual('PostgreSQL 12.1.5');
    });

    it('should return "MySQL 5.7.31" when type is mysql', () => {
        const result = pipe.transform({
            ...serviceModel,
            type: ServiceModelType.MYSQL
        });
        expect(result).toEqual('MySQL 5.7.31');
    });
});
