import { ChangeDetectionStrategy, Component } from '@angular/core';
import { ServiceModel } from '@models/service.model';
import { ServicesPageService } from '@services/services-page.service';
import { Observable } from 'rxjs';

@Component({
    selector: 'kl-services-list',
    templateUrl: './services-list.component.html',
    styleUrls: ['./services-list.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ServicesListComponent {
    services$: Observable<ServiceModel[] | undefined>;

    constructor(private service: ServicesPageService) {
        this.services$ = this.service.getServicesList();
    }

}
