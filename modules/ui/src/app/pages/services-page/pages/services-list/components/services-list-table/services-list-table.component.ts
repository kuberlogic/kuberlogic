import { ChangeDetectionStrategy, Component, Input } from '@angular/core';
import { ServiceModel, ServiceModelType } from '@models/service.model';

@Component({
    selector: 'kl-services-list-table',
    templateUrl: './services-list-table.component.html',
    styleUrls: ['./services-list-table.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ServicesListTableComponent {
    @Input() services: ServiceModel[] = [];
    ServiceModelType = ServiceModelType;

    displayedColumns: string[] = [
        'type',
        'name',
        'status',
        'masters',
        'replicas',
        'created_time',
    ];
}
