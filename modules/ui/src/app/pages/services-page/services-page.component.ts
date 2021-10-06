import { ChangeDetectionStrategy, Component } from '@angular/core';

@Component({
    selector: 'kl-services-page',
    templateUrl: './services-page.component.html',
    styleUrls: ['./services-page.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ServicesPageComponent {
}
