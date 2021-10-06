import { ChangeDetectionStrategy, Component, Input } from '@angular/core';

@Component({
    selector: 'kl-table-skeleton',
    templateUrl: './table-skeleton.component.html',
    styleUrls: ['./table-skeleton.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class TableSkeletonComponent {
    @Input() rows = 1;
    @Input() columns = 3;
}
