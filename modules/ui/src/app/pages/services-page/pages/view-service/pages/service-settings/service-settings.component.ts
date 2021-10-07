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

import { ChangeDetectionStrategy, ChangeDetectorRef, Component, OnInit } from '@angular/core';
import { ServiceModel } from '@models/service.model';
import { ServiceSettingsFormResult } from '@pages/services-page/pages/view-service/pages/service-settings/components/service-settings-form/service-settings-form.component';
import { MessagesService } from '@services/messages.service';
import { ServicesPageService } from '@services/services-page.service';
import { throwError } from 'rxjs';
import { catchError, finalize, tap } from 'rxjs/operators';

@Component({
    selector: 'kl-service-settings',
    templateUrl: './service-settings.component.html',
    styleUrls: ['./service-settings.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ServiceSettingsComponent implements OnInit {
    currentService: ServiceModel | undefined;
    isSaving = false;

    private currentServiceId = '';

    constructor(
        private servicesPageService: ServicesPageService,
        private messagesService: MessagesService,
        private cdRef: ChangeDetectorRef,
    ) {
    }

    ngOnInit(): void {
        this.currentServiceId = this.servicesPageService.getCurrentServiceId();

        this.servicesPageService.getCurrentService().subscribe((serviceModel) => {
            this.currentService = serviceModel;
            this.cdRef.detectChanges();
        });
    }

    onSave(result: ServiceSettingsFormResult): void {
        this.isSaving = true;
        this.cdRef.detectChanges();

        const saveService$ = this.servicesPageService.editService(
            this.currentServiceId,
            {...this.currentService, ...result.service}
        );

        saveService$
            .pipe(
                catchError((err) => {
                    this.messagesService.error('An error occurred, please try again later');
                    return throwError(err);
                }),
                tap(() => {
                    this.messagesService.success('Service was successfully updated');
                }),
                finalize(() => {
                    this.isSaving = false;
                    this.cdRef.detectChanges();
                }),
            )
            .subscribe();
    }

}
