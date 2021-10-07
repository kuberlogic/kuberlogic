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

import { Component, OnInit } from '@angular/core';
import { ServiceLogModel } from '@models/service-log.model';
import { ServiceModel } from '@models/service.model';
import { ServiceLogsService } from '@services/service-logs.service';
import { ServicesPageService } from '@services/services-page.service';
import { Observable } from 'rxjs';

@Component({
    selector: 'kl-service-logs',
    templateUrl: './service-logs.component.html',
    styleUrls: ['./service-logs.component.scss'],
})
export class ServiceLogsComponent implements OnInit {
    serviceLogs$!: Observable<ServiceLogModel | undefined>;
    currentService$!: Observable<ServiceModel | undefined>;

    constructor(
        private serviceLogsService: ServiceLogsService,
        private servicesPageService: ServicesPageService,
    ) { }

    ngOnInit(): void {
        this.currentService$ = this.servicesPageService.getCurrentService();
    }

    onFormSubmit(serviceInstance: any): void {
        this.serviceLogs$ = this.serviceLogsService.get(
            this.servicesPageService.getCurrentServiceId(), serviceInstance
        );
    }

    renderType(type: string | undefined): string | undefined {
        if (typeof type === 'undefined') {
            return type;
        }
        const types: {[key: string]: string} = {
            postgresql: 'PostgreSQL',
            mysql: 'MySQL',
        };
        return types[type] !== undefined ? types[type] : type;
    }
}
