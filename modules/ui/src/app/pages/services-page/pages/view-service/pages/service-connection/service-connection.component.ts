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

import { ChangeDetectionStrategy, Component, OnInit } from '@angular/core';
import { ServiceModel } from '@models/service.model';
import { ServicesPageService } from '@services/services-page.service';
import { Observable } from 'rxjs';

@Component({
    selector: 'kl-service-connection',
    templateUrl: './service-connection.component.html',
    styleUrls: ['./service-connection.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ServiceConnectionComponent implements OnInit {
    showInnerConnection = false;
    noExternalConnection = false;
    currentService$!: Observable<ServiceModel | undefined>;

    constructor(
        private servicesPageService: ServicesPageService,
    ) { }

    ngOnInit(): void {
        this.currentService$ = this.servicesPageService.getCurrentService();
        this.currentService$.subscribe((service) => {
            if (service && service.externalConnection?.master?.host === undefined) {
                this.showInnerConnection = true;
                this.noExternalConnection = true;
            }
        });
    }

}
