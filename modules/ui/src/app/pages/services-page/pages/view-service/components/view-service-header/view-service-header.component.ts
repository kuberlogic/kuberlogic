import { ChangeDetectionStrategy, Component, Input, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { ServiceModel, ServiceModelStatus, ServiceModelType } from '@models/service.model';
import { MessagesService } from '@services/messages.service';
import { ServicesPageService } from '@services/services-page.service';
import { ConfirmDialogService } from '@ui/confirm-dialog/confirm-dialog.service';
import { catchError, tap } from 'rxjs/operators';

@Component({
    selector: 'kl-edit-service-header',
    templateUrl: './view-service-header.component.html',
    styleUrls: ['./view-service-header.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ViewServiceHeaderComponent implements OnInit {
    @Input() serviceModel!: ServiceModel;
    ServiceModelStatus = ServiceModelStatus;
    ServiceModelType = ServiceModelType;

    constructor(
        private servicesPageService: ServicesPageService,
        private messages: MessagesService,
        private confirmService: ConfirmDialogService,
        private router: Router,
    ) { }

    ngOnInit(): void {
    }

    deleteService(): void {
        this.confirmService.confirm(
            `Delete ${this.serviceModel.name} service`,
            'Are you sure you want to delete this service?',
        )
            .subscribe((result) => {
                if (result) {
                    this.servicesPageService.deleteService(this.serviceModel)
                        .pipe(
                            catchError((err) => {
                                this.messages.success('An error occurred, please try again later');
                                return err;
                            }),
                            tap(() => {
                                this.messages.success('Service was successfully deleted');
                                this.router.navigate(['/services']);
                            }),
                        )
                        .subscribe();
                }
            });
    }

}
