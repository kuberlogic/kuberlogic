import { Component, Input } from '@angular/core';
import { ServiceRestoreModel } from '@models/service-restore.model';

@Component({
    selector: 'kl-service-restores-table',
    templateUrl: './service-restores-table.component.html',
    styleUrls: ['./service-restores-table.component.scss'],
})
export class ServiceRestoresTableComponent {
    @Input() restores!: ServiceRestoreModel[];
    readonly displayedColumns: string[] = ['file', 'time', 'status'];
}
