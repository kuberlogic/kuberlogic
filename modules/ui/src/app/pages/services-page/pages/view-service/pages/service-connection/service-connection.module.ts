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
import { FormsModule } from '@angular/forms';
import { MatIconModule } from '@angular/material/icon';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatTooltipModule } from '@angular/material/tooltip';
import { ServiceConnectionTableModule } from '@pages/services-page/pages/view-service/pages/service-connection/components/service-connection-table/service-connection-table.module';
import { ServiceConnectionRoutingModule } from '@pages/services-page/pages/view-service/pages/service-connection/service-connection-routing.module';
import { ServiceConnectionComponent } from './service-connection.component';

@NgModule({
    declarations: [ServiceConnectionComponent],
    imports: [
        CommonModule,
        ServiceConnectionRoutingModule,
        MatIconModule,
        MatSlideToggleModule,
        MatTooltipModule,
        ServiceConnectionTableModule,
        FormsModule
    ]
})
export class ServiceConnectionModule { }
