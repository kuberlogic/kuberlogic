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

import { Component, OnInit } from '@angular/core';
import { AuthService } from '@services/auth.service';
import { environment } from '@environments/environment';

@Component({
    selector: 'kl-header',
    templateUrl: './header.component.html',
    styleUrls: ['./header.component.scss'],
})
export class HeaderComponent implements OnInit {
    consoleMonitoringUrl: any;
    helpUrl: any;

    constructor(
        private authService: AuthService,
    ) { }

    ngOnInit(): void {
        this.consoleMonitoringUrl = this.getConsoleMonitoringUrl();
        this.helpUrl = this.getHelpUrl();
    }

    onLogout(): void {
        this.authService.logout();
    }

    getConsoleMonitoringUrl() {
        return `${environment.monitoringConsoleUrl}?token=${this.authService.getToken()}`;
    }

    getHelpUrl() {
        return environment.helpUrl;
    }
}
