import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { RouterModule } from '@angular/router';
import { CreateServiceHeaderComponent } from './create-service-header.component';

@NgModule({
    declarations: [CreateServiceHeaderComponent],
    exports: [
        CreateServiceHeaderComponent
    ],
    imports: [
        CommonModule,
        MatButtonModule,
        MatIconModule,
        RouterModule
    ]
})
export class CreateServiceHeaderModule { }
