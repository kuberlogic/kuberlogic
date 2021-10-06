import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { RouterModule } from '@angular/router';
import { PipesModule } from '@pipes/pipes.module';
import { ViewServiceHeaderComponent } from './view-service-header.component';

@NgModule({
    declarations: [ViewServiceHeaderComponent],
    imports: [
        CommonModule,
        MatButtonModule,
        MatIconModule,
        RouterModule,
        PipesModule
    ],
    exports: [
        ViewServiceHeaderComponent
    ]
})
export class ViewServiceHeaderModule { }
