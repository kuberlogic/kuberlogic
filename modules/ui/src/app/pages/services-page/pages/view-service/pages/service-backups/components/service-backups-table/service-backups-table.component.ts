import { Component, EventEmitter, Input, Output } from '@angular/core';
import { ServiceBackupModel } from '@models/service-backup.model';

@Component({
    selector: 'kl-service-backups-table',
    templateUrl: './service-backups-table.component.html',
    styleUrls: ['./service-backups-table.component.scss'],
})
export class ServiceBackupsTableComponent {
    @Input() backups!: ServiceBackupModel[];
    @Output() restore = new EventEmitter<string>();
    readonly displayedColumns: string[] = ['size', 'lastModified', 'actions'];

    onRestore(name: string): void {
        this.restore.emit(name);
    }
}
