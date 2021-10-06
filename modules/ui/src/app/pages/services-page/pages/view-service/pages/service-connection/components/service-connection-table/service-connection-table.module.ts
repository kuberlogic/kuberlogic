import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { TableSkeletonModule } from '@ui/table-skeleton/table-skeleton.module';
import { NgxSkeletonLoaderModule } from 'ngx-skeleton-loader';
import { ServiceConnectionTableComponent } from './service-connection-table.component';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { ChangePasswordDialogModule } from '@pages/services-page/pages/view-service/pages/service-connection/components/change-password-dialog/change-password-dialog.module';

@NgModule({
    declarations: [
        ServiceConnectionTableComponent
    ],
    exports: [
        ServiceConnectionTableComponent,
    ],
    imports: [
        MatButtonModule,
        MatIconModule,
        CommonModule,
        NgxSkeletonLoaderModule,
        TableSkeletonModule,
        ChangePasswordDialogModule
    ]
})
export class ServiceConnectionTableModule { }
