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
import { MatTabsModule } from '@angular/material/tabs';
import { ViewServiceHeaderModule } from '@pages/services-page/pages/view-service/components/view-service-header/view-service-header.module';
import { ViewServiceRoutingModule } from '@pages/services-page/pages/view-service/view-service-routing.module';
import { ViewServiceComponent } from './view-service.component';

@NgModule({
    declarations: [ViewServiceComponent],
    imports: [
        CommonModule,
        ViewServiceRoutingModule,
        ViewServiceHeaderModule,
        MatTabsModule,
    ]
})
export class ViewServiceModule { }
