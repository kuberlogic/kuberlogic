import { Component, EventEmitter, OnInit, Output } from '@angular/core';
import { FormBuilder, FormGroup } from '@angular/forms';
import { FormContainerMixin } from '@app/mixins/form-container.mixin';
import { BaseObject } from '@app/mixins/mixins';
import { MessagesService } from '@services/messages.service';
import { ServiceInstancesNames, ServicesPageService } from '@services/services-page.service';
import { BehaviorSubject, throwError } from 'rxjs';
import { catchError, tap } from 'rxjs/operators';

@Component({
    selector: 'kl-service-logs-form',
    templateUrl: './service-logs-form.component.html',
    styleUrls: ['./service-logs-form.component.scss'],
})
export class ServiceLogsFormComponent extends FormContainerMixin(BaseObject) implements OnInit {
    dataSource = new BehaviorSubject<ServiceInstancesNames | undefined>(new Map<string, string>());

    formGroup: FormGroup;
    @Output() successfulSubmit = new EventEmitter<string>();

    constructor(
        private fb: FormBuilder,
        private servicesPageService: ServicesPageService,
        private messagesService: MessagesService,
    ) {
        super();
        this.formGroup = this.fb.group({
            serviceInstance: [''],
        });
    }

    ngOnInit(): void {
        this.servicesPageService.getCurrentServiceInstancesNames()
            .pipe(
                catchError((err) => {
                    this.messagesService.error('An error occurred, please try again later');
                    return throwError(err);
                }),
                tap((data) => {
                    this.dataSource.next(data);
                    this.selectFirstInstance(data);
                }),
            )
            .subscribe();
    }

    selectFirstInstance(data: ServiceInstancesNames | undefined): void {
        if (!data) {
            return;
        }
        const value = data.keys().next().value;
        this.formGroup.controls.serviceInstance.setValue(value);
        this.successfulSubmit.emit(value);
    }

    onSubmit(): void {
        this.successfulSubmit.emit(this.formGroup.controls.serviceInstance.value);
    }
}
