import { Component, Input, OnInit } from '@angular/core';
import { ServiceModel, ServiceModelStatus } from '@models/service.model';
import { MessagesService } from '@services/messages.service';
import { ServicesPageService } from '@services/services-page.service';
import { throwError } from 'rxjs';
import { catchError, tap } from 'rxjs/operators';

@Component({
    selector: 'kl-read-replicas-select',
    templateUrl: './read-replicas-select.component.html',
    styleUrls: ['./read-replicas-select.component.scss'],
})
export class ReadReplicasSelectComponent implements OnInit {
    @Input() service: ServiceModel | undefined;
    selected: number | undefined;
    readonly replicas = [...Array(11).keys()];
    readonly serviceStatus = ServiceModelStatus;
    constructor(
        private messagesService: MessagesService,
        private pageService: ServicesPageService
    ) {}

    ngOnInit(): void {
        this.selected = this.service?.replicas;
    }

    onSubmit(value: any): void {
        if (this.service) {
            this.pageService.editService(`${this.service.ns}:${this.service.name}`, {
                replicas: value,
                name: this.service.name,
                ns: this.service.ns,
                type: this.service.type,
            }).pipe(
                catchError((err) => {
                    this.messagesService.error('An error occurred, please try again later');
                    return throwError(err);
                }),
                tap(() => {
                    this.messagesService.success('Service was successfully updated');
                }),
            ).subscribe();
        }
    }
}
