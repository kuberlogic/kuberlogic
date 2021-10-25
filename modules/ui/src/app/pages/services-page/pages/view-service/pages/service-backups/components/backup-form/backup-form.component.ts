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

import {
    ChangeDetectionStrategy,
    ChangeDetectorRef,
    Component, EventEmitter, Input, OnChanges, OnInit, Output, SimpleChanges,
} from '@angular/core';
import { FormBuilder, FormGroup } from '@angular/forms';
import { MatDialog } from '@angular/material/dialog';
import { FormContainerMixin } from '@app/mixins/form-container.mixin';
import { BaseObject } from '@app/mixins/mixins';
import { ServiceBackupConfigModel } from '@models/service-backup-config.model';
import { ServiceModel } from '@models/service.model';
import { BackupStorageDialogComponent } from '@pages/services-page/pages/view-service/pages/service-backups/components/backup-storage-dialog/backup-storage-dialog.component';
import { BackupConfigService } from '@services/backup-config.service';
import { MessagesService } from '@services/messages.service';
import { ServicesPageService } from '@services/services-page.service';
import { Observable, throwError } from 'rxjs';
import { catchError, finalize, tap } from 'rxjs/operators';

export interface ServiceSettingsFormResult {
    service?: ServiceModel;
    backup?: ServiceBackupConfigModel;
}

@Component({
    selector: 'kl-backup-form',
    templateUrl: './backup-form.component.html',
    styleUrls: ['./backup-form.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush,
})
export class BackupFormComponent extends FormContainerMixin(BaseObject) implements OnInit, OnChanges {
    @Input() isSaving = false;
    @Output() successfulSubmit = new EventEmitter<ServiceSettingsFormResult>();
    backupConfig: ServiceBackupConfigModel | undefined;
    formGroup: FormGroup;
    private currentServiceId = '';

    constructor(
        private servicesPageService: ServicesPageService,
        private backupConfigService: BackupConfigService,
        private dialog: MatDialog,
        private fb: FormBuilder,
        private cdRef: ChangeDetectorRef,
        private messagesService: MessagesService,
    ) {
        super();
        this.formGroup = this.fb.group({
            enabled: [false],
            aws_access_key_id: ['', []],
            aws_secret_access_key: ['', []],
            region: ['', []],
            bucket: ['', []],
            endpoint: ['', []],
            schedule: ['0 3 * * 4', []],
        });
    }
    ngOnInit(): void {
        this.currentServiceId = this.servicesPageService.getCurrentServiceId();

        this.backupConfigService.getBackupConfig(this.currentServiceId).subscribe((config) => {
            this.backupConfig = config;
            this.setupBackup(config);
            this.cdRef.detectChanges();
        });
    }

    ngOnChanges(changes: SimpleChanges): void {
        if (changes.serviceModelBackup) {
            this.setupBackup(this.backupConfig);
        }
    }

    onOpen(): void {
        const dialogRef = this.dialog.open(BackupStorageDialogComponent, {
            closeOnNavigation: true,
            data: this.formGroup.value,
        });
        dialogRef.afterClosed().subscribe((data) => {
            if (data !== false) {
                this.formGroup.patchValue(data);
                this.save();
            }
        });
    }

    save(): void {
        this.isSaving = true;
        this.cdRef.detectChanges();

        const saveConfig$: Observable<ServiceBackupConfigModel> = this.backupConfig
            ? this.backupConfigService.editBackupConfig(this.currentServiceId, this.formGroup.value)
            : this.backupConfigService.createBackupConfig(this.currentServiceId, this.formGroup.value);

        saveConfig$
            .pipe(
                catchError((err) => {
                    this.messagesService.error('An error occurred, please try again later');
                    return throwError(err);
                }),
                tap((config) => {
                    this.backupConfig = config;
                    this.messagesService.success('Service was successfully updated');
                }),
                finalize(() => {
                    this.isSaving = false;
                    this.cdRef.detectChanges();
                }),
            )
            .subscribe();
    }

    private setupBackup(backup: ServiceBackupConfigModel | undefined): void {
        if (!!backup) {
            this.formGroup.patchValue(backup);
        }
    }
}
