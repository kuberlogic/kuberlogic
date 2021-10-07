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
import { MatDialog } from '@angular/material/dialog';
import { ServiceBackupModel } from '@models/service-backup.model';
import { ServiceRestoreModel } from '@models/service-restore.model';
import { RestoreServiceBackupsDialogComponent } from '@pages/services-page/pages/view-service/pages/service-backups/components/restore-service-backups-dialog/restore-service-backups-dialog.component';
import { ServiceBackupsService } from '@services/service-backups.service';
import { ServiceRestoresService } from '@services/service-restores.service';
import { ServicesPageService } from '@services/services-page.service';
import { Observable } from 'rxjs';

@Component({
    selector: 'kl-service-backups',
    templateUrl: './service-backups.component.html',
    styleUrls: ['./service-backups.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ServiceBackupsComponent implements OnInit {
    serviceBackups$!: Observable<ServiceBackupModel[] | undefined>;
    serviceRestores$!: Observable<ServiceRestoreModel[] | undefined>;

    private currentServiceId = '';

    constructor(
        private serviceBackupsService: ServiceBackupsService,
        private serviceRestoresService: ServiceRestoresService,
        private servicesPageService: ServicesPageService,
        private dialog: MatDialog,
    ) { }

    ngOnInit(): void {
        this.currentServiceId = this.servicesPageService.getCurrentServiceId();
        this.serviceBackups$ = this.serviceBackupsService.getList(this.currentServiceId);
        this.serviceRestores$ = this.serviceRestoresService.getList(this.currentServiceId);
    }

    onRestore(name: string): void {
        this.dialog.open(RestoreServiceBackupsDialogComponent, {
            disableClose: true,
            closeOnNavigation: true,
            data: name,
        });
    }
}
