import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { TimeagoModule } from 'ngx-timeago';
import { TimeUtcComponent } from './time-utc.component';

@NgModule({
    declarations: [TimeUtcComponent],
    imports: [
        CommonModule,
        TimeagoModule,
    ],
    exports: [
        TimeUtcComponent,
    ],
})
export class TimeUtcModule { }
