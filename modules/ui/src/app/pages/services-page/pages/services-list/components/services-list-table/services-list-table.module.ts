import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { RouterModule } from '@angular/router';
import { ReadReplicasSelectModule } from '@pages/services-page/pages/services-list/components/read-replicas-select/read-replicas-select.module';
import { PipesModule } from '@pipes/pipes.module';
import { TableSkeletonModule } from '@ui/table-skeleton/table-skeleton.module';
import { NgxSkeletonLoaderModule } from 'ngx-skeleton-loader';
import { TimeagoModule } from 'ngx-timeago';
import { ServicesListTableComponent } from './services-list-table.component';

@NgModule({
    declarations: [ServicesListTableComponent],
    exports: [
        ServicesListTableComponent
    ],
    imports: [
        CommonModule,
        MatTableModule,
        MatIconModule,
        PipesModule,
        TimeagoModule,
        RouterModule,
        NgxSkeletonLoaderModule,
        TableSkeletonModule,
        ReadReplicasSelectModule,
    ]
})
export class ServicesListTableModule { }
