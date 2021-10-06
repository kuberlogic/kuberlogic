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
