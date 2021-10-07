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

import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatSelectModule } from '@angular/material/select';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { BackupFormModule } from '@pages/services-page/pages/view-service/pages/service-backups/components/backup-form/backup-form.module';
import { BackupStorageDialogModule } from '@pages/services-page/pages/view-service/pages/service-backups/components/backup-storage-dialog/backup-storage-dialog.module';
import { RestoreServiceBackupsDialogModule } from '@pages/services-page/pages/view-service/pages/service-backups/components/restore-service-backups-dialog/restore-service-backups-dialog.module';
import { ServiceBackupsTableModule } from '@pages/services-page/pages/view-service/pages/service-backups/components/service-backups-table/service-backups-table.module';
import { ServiceRestoresTableModule } from '@pages/services-page/pages/view-service/pages/service-backups/components/service-restores-table/service-restores-table.module';
import { ServiceBackupsRoutingModule } from '@pages/services-page/pages/view-service/pages/service-backups/service-backups-routing.module';
import { ServiceBackupsComponent } from './service-backups.component';

@NgModule({
    declarations: [
        ServiceBackupsComponent,
    ],
    imports: [
        CommonModule,
        MatIconModule,
        ServiceBackupsRoutingModule,
        ServiceBackupsTableModule,
        ServiceRestoresTableModule,
        RestoreServiceBackupsDialogModule,
        BackupStorageDialogModule,
        ReactiveFormsModule,
        MatFormFieldModule,
        MatSelectModule,
        MatSlideToggleModule,
        BackupFormModule,
    ]
})
export class ServiceBackupsModule { }
