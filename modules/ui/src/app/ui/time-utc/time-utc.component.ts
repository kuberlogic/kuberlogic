import { Component, Input } from '@angular/core';

@Component({
    selector: 'kl-time-utc',
    templateUrl: './time-utc.component.html',
})
export class TimeUtcComponent {
    @Input() timestamp!: string;
}
