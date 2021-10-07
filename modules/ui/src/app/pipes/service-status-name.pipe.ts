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
import { ServiceModelStatus } from '@models/service.model';

const statusName = {
    [ServiceModelStatus.RUNNING]: 'Running',
    [ServiceModelStatus.READY]: 'Ready',
    [ServiceModelStatus.FAILED]: 'Failed',
    [ServiceModelStatus.PROCESSING]: 'Processing',
    [ServiceModelStatus.UNKNOWN]: 'Unknown',
    [ServiceModelStatus.NOT_READY]: 'Not Ready',
};

@Pipe({
    name: 'klServiceStatusName'
})
export class ServiceStatusNamePipe implements PipeTransform {

    transform(status: ServiceModelStatus | undefined): string {
        return !!status ? statusName[status] : statusName[ServiceModelStatus.UNKNOWN];
    }

}
