import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { MatRippleModule } from '@angular/material/core';
import { MatIconModule } from '@angular/material/icon';
import { RadioGroupComponent } from './radio-group.component';

@NgModule({
    declarations: [RadioGroupComponent],
    exports: [
        RadioGroupComponent
    ],
    imports: [
        CommonModule,
        MatIconModule,
        MatRippleModule
    ]
})
export class RadioGroupModule { }
