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
import { MatButtonModule } from '@angular/material/button';
import { MatTableModule } from '@angular/material/table';
import { ServiceBackupsTableComponent } from '@pages/services-page/pages/view-service/pages/service-backups/components/service-backups-table/service-backups-table.component';
import { PipesModule } from '@pipes/pipes.module';
import { ResetServiceUserPasswordDialogModule } from '@ui/reset-service-user-password-dialog/reset-service-user-password-dialog.module';
import { TableSkeletonModule } from '@ui/table-skeleton/table-skeleton.module';
import { TimeUtcModule } from '@ui/time-utc/time-utc.module';
import { NgxFilesizeModule } from 'ngx-filesize';
import { NgxSkeletonLoaderModule } from 'ngx-skeleton-loader';
import { TimeagoModule } from 'ngx-timeago';

@NgModule({
    declarations: [ServiceBackupsTableComponent],
    exports: [
        ServiceBackupsTableComponent
    ],
    imports: [
        CommonModule,
        MatTableModule,
        MatButtonModule,
        ResetServiceUserPasswordDialogModule,
        NgxSkeletonLoaderModule,
        TableSkeletonModule,
        PipesModule,
        TimeagoModule,
        NgxFilesizeModule,
        TimeUtcModule,
    ]
})
export class ServiceBackupsTableModule { }
