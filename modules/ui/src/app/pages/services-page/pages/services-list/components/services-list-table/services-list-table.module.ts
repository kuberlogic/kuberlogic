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
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { RouterModule } from '@angular/router';
import { ReadReplicasSelectModule } from '@pages/services-page/pages/services-list/components/read-replicas-select/read-replicas-select.module';
import { PipesModule } from '@pipes/pipes.module';
import { TableSkeletonModule } from '@ui/table-skeleton/table-skeleton.module';
import { NgxSkeletonLoaderModule } from 'ngx-skeleton-loader';
import { TimeagoModule } from 'ngx-timeago';
import { ServicesListTableComponent } from './services-list-table.component';

@NgModule({
    declarations: [ServicesListTableComponent],
    exports: [
        ServicesListTableComponent
    ],
    imports: [
        CommonModule,
        MatTableModule,
        MatIconModule,
        PipesModule,
        TimeagoModule,
        RouterModule,
        NgxSkeletonLoaderModule,
        TableSkeletonModule,
        ReadReplicasSelectModule,
    ]
})
export class ServicesListTableModule { }
