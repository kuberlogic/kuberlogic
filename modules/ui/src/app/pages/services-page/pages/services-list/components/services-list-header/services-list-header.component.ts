import { ChangeDetectionStrategy, Component } from '@angular/core';

@Component({
    selector: 'kl-services-list-header',
    templateUrl: './services-list-header.component.html',
    styleUrls: ['./services-list-header.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ServicesListHeaderComponent { }
