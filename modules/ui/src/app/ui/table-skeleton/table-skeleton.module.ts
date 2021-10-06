import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { NgxSkeletonLoaderModule } from 'ngx-skeleton-loader';
import { TableSkeletonComponent } from './table-skeleton.component';

@NgModule({
    declarations: [TableSkeletonComponent],
    exports: [
        TableSkeletonComponent
    ],
    imports: [
        CommonModule,
        NgxSkeletonLoaderModule
    ]
})
export class TableSkeletonModule { }
