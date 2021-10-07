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

import { Injectable } from '@angular/core';
import { MatIconRegistry } from '@angular/material/icon';
import { DomSanitizer } from '@angular/platform-browser';

@Injectable({
    providedIn: 'root'
})
export class IconsService {

    constructor(
        private matIconRegistry: MatIconRegistry,
        private domSanitizer: DomSanitizer,
    ) {
    }

    init(): void {
        this.matIconRegistry.addSvgIcon(
            'mysqlIcon',
            this.domSanitizer.bypassSecurityTrustResourceUrl('assets/svg/service-logos/mysql.svg'),
        );
        this.matIconRegistry.addSvgIcon(
            'postgresqlIcon',
            this.domSanitizer.bypassSecurityTrustResourceUrl('assets/svg/service-logos/postgresql.svg'),
        );
    }
}
