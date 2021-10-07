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

import { ChangeDetectionStrategy, Component, ViewChild } from '@angular/core';
import { Router } from '@angular/router';
import { ServiceModel, ServiceModelLimit } from '@models/service.model';
import { CreateServiceFormComponent } from '@pages/services-page/pages/create-service/components/create-service-form/create-service-form.component';
import { MessagesService } from '@services/messages.service';
import { ServicesPageService } from '@services/services-page.service';
import { catchError, tap } from 'rxjs/operators';

@Component({
    selector: 'kl-create-service',
    templateUrl: './create-service.component.html',
    styleUrls: ['./create-service.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class CreateServiceComponent {
    @ViewChild('createServiceForm', { static: false }) createServiceForm!: CreateServiceFormComponent;

    constructor(
        private messages: MessagesService,
        private servicesPageService: ServicesPageService,
        private router: Router,
    ) {
    }

    submitForm(): void {
        this.createServiceForm.onSave();
    }

    createService(serviceModel: Partial<ServiceModel>): void {
        this.servicesPageService.createService(serviceModel)
            .pipe(
                catchError((err) => {
                    this.messages.error(err);
                    throw err;
                }),
                tap(() => {
                    this.messages.success('Service was successfully created');
                    this.router.navigate(['/services']);
                })
            ).subscribe();
    }

}
