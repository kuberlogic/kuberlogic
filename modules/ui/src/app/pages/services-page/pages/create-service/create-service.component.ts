import { ChangeDetectionStrategy, Component, ViewChild } from '@angular/core';
import { Router } from '@angular/router';
import { ServiceModel, ServiceModelLimit } from '@models/service.model';
import { CreateServiceFormComponent } from '@pages/services-page/pages/create-service/components/create-service-form/create-service-form.component';
import { MessagesService } from '@services/messages.service';
import { ServicesPageService } from '@services/services-page.service';
import { catchError, tap } from 'rxjs/operators';

@Component({
    selector: 'kl-create-service',
    templateUrl: './create-service.component.html',
    styleUrls: ['./create-service.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class CreateServiceComponent {
    @ViewChild('createServiceForm', { static: false }) createServiceForm!: CreateServiceFormComponent;

    constructor(
        private messages: MessagesService,
        private servicesPageService: ServicesPageService,
        private router: Router,
    ) {
    }

    submitForm(): void {
        this.createServiceForm.onSave();
    }

    createService(serviceModel: Partial<ServiceModel>): void {
        this.servicesPageService.createService(serviceModel)
            .pipe(
                catchError((err) => {
                    this.messages.error(err);
                    throw err;
                }),
                tap(() => {
                    this.messages.success('Service was successfully created');
                    this.router.navigate(['/services']);
                })
            ).subscribe();
    }

}
