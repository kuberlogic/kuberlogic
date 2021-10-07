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

import { ChangeDetectionStrategy, Component, Input } from '@angular/core';
import { ServiceModel, ServiceModelConnection } from '@models/service.model';
import { MatDialog } from '@angular/material/dialog';
import { ChangePasswordDialogComponent } from '@pages/services-page/pages/view-service/pages/service-connection/components/change-password-dialog/change-password-dialog.component';
import { ServicesPageService } from '@services/services-page.service';
import { ServiceUsersService } from '@services/service-users.service';
import { catchError, tap } from 'rxjs/operators';
import { throwError } from 'rxjs';
import { MessagesService } from '@services/messages.service';

@Component({
    selector: 'kl-service-connection-table',
    templateUrl: './service-connection-table.component.html',
    styleUrls: ['./service-connection-table.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ServiceConnectionTableComponent {
    @Input() serviceModel!: ServiceModel;
    @Input() showInnerConnection = false;

    constructor(
        private dialog: MatDialog,
        private servicesPageService: ServicesPageService,
        private messagesService: MessagesService,
        private servicesUsersService: ServiceUsersService,
    ) {
    }

    get connection(): ServiceModelConnection | undefined {
        return this.showInnerConnection ? this.serviceModel.internalConnection : this.serviceModel.externalConnection;
    }

    changePassword(): void{
        const dialogRef = this.dialog.open(ChangePasswordDialogComponent, {
            closeOnNavigation: true,
        });

        dialogRef.afterClosed().subscribe((password) => {
            if (password !== false) {
                this.servicesUsersService.changePassword(
                        this.servicesPageService.getCurrentServiceId(),
                        this.connection?.master.user,
                        password
                    ).pipe(
                        catchError((err) => {
                            this.messagesService.error('An error occurred, please try again later');
                            return throwError(err);
                        }),
                        tap(() => {
                            this.messagesService.success('Password was successfully updated');
                        }),
                    ).subscribe();
            }
        });
    }

}
