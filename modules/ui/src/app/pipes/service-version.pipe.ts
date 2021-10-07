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

import { Pipe, PipeTransform } from '@angular/core';
import { ServiceModel, ServiceModelType } from '@models/service.model';

@Pipe({
    name: 'klServiceVersion'
})
export class ServiceVersionPipe implements PipeTransform {

    transform(serviceModel: ServiceModel | undefined): string {
        if (!!serviceModel) {
            // @TODO remove hardcoded version when version support is implemented
            return serviceModel.type === ServiceModelType.POSTGRES ? 'PostgreSQL 12.1.5' : 'MySQL 5.7.31';
        }
        return '';
    }

}
