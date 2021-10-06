import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { RouterModule } from '@angular/router';
import { ServicesListHeaderComponent } from './services-list-header.component';

@NgModule({
    declarations: [ServicesListHeaderComponent],
    exports: [
        ServicesListHeaderComponent
    ],
    imports: [
        CommonModule,
        MatButtonModule,
        MatIconModule,
        RouterModule
    ]
})
export class ServicesListHeaderModule { }
