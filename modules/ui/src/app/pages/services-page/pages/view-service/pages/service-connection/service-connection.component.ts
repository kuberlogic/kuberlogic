import { ChangeDetectionStrategy, Component, OnInit } from '@angular/core';
import { ServiceModel } from '@models/service.model';
import { ServicesPageService } from '@services/services-page.service';
import { Observable } from 'rxjs';

@Component({
    selector: 'kl-service-connection',
    templateUrl: './service-connection.component.html',
    styleUrls: ['./service-connection.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ServiceConnectionComponent implements OnInit {
    showInnerConnection = false;
    noExternalConnection = false;
    currentService$!: Observable<ServiceModel | undefined>;

    constructor(
        private servicesPageService: ServicesPageService,
    ) { }

    ngOnInit(): void {
        this.currentService$ = this.servicesPageService.getCurrentService();
        this.currentService$.subscribe((service) => {
            if (service && service.externalConnection?.master?.host === undefined) {
                this.showInnerConnection = true;
                this.noExternalConnection = true;
            }
        });
    }

}
