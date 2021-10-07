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

import { ChangeDetectionStrategy, Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { ServiceModel } from '@models/service.model';
import { ServicesPageService } from '@services/services-page.service';
import { Observable } from 'rxjs';

export interface NavLink {
    label: string;
    link: string;
}

@Component({
    selector: 'kl-edit-service',
    templateUrl: './view-service.component.html',
    styleUrls: ['./view-service.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ViewServiceComponent implements OnInit {
    navLinks: NavLink[];
    currentService$!: Observable<ServiceModel | undefined>;

    constructor(
        private route: ActivatedRoute,
        private servicesPageService: ServicesPageService,
    ) {
        this.navLinks = [
            {
                label: 'Connection',
                link: 'connection',
            },
            {
                label: 'Settings',
                link: 'settings',
            },
            {
                label: 'Logs',
                link: 'logs',
            },
            {
                label: 'Backups',
                link: 'backups',
            },
        ];
    }

    ngOnInit(): void {
        this.currentService$ = this.servicesPageService.getService(this.route.snapshot.params.id);
    }

}
