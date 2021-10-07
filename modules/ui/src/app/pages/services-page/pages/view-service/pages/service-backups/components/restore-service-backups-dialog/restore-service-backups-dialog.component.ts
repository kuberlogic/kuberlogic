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

import { ChangeDetectionStrategy, Component, Inject, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { FormContainerMixin } from '@app/mixins/form-container.mixin';
import { BaseObject } from '@app/mixins/mixins';
import { ServiceDatabaseModel } from '@models/service-database.model';
import { MessagesService } from '@services/messages.service';
import { ServiceDatabasesService } from '@services/service-databases.service';
import { ServiceRestoresService } from '@services/service-restores.service';
import { ServicesPageService } from '@services/services-page.service';
import { Observable } from 'rxjs';

@Component({
    selector: 'kl-restore-service-backups-dialog',
    templateUrl: './restore-service-backups-dialog.component.html',
    styleUrls: ['./restore-service-backups-dialog.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class RestoreServiceBackupsDialogComponent
    extends FormContainerMixin(BaseObject)
    implements OnInit {
    formGroup: FormGroup;
    serviceDatabases$!: Observable<ServiceDatabaseModel[] | undefined>;

    constructor(
        private fb: FormBuilder,
        private messages: MessagesService,
        private servicesPageService: ServicesPageService,
        private serviceDatabasesService: ServiceDatabasesService,
        private serviceRestoresService: ServiceRestoresService,
        private dialogRef: MatDialogRef<RestoreServiceBackupsDialogComponent>,
        @Inject(MAT_DIALOG_DATA) public name: string,
    ) {
        super();
        this.formGroup = this.fb.group({
            database: ['', [Validators.required]],
        });
    }

    ngOnInit(): void {
        this.serviceDatabases$ = this.serviceDatabasesService.getDatabases(
            this.servicesPageService.getCurrentServiceId()
        );
    }

    onSave(): void {
        if (this.checkForm()) {
            this.serviceRestoresService.restore(this.servicesPageService.getCurrentServiceId(),
                this.name, this.formGroup.value.database).subscribe(
                () => {
                    this.messages.success('Database was successfully restored');
                    this.dialogRef.close();
                }
            );
        }
    }

}
