import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatTableModule } from '@angular/material/table';
import { ServiceRestoresTableComponent } from '@pages/services-page/pages/view-service/pages/service-backups/components/service-restores-table/service-restores-table.component';
import { ResetServiceUserPasswordDialogModule } from '@ui/reset-service-user-password-dialog/reset-service-user-password-dialog.module';
import { TableSkeletonModule } from '@ui/table-skeleton/table-skeleton.module';
import { TimeUtcModule } from '@ui/time-utc/time-utc.module';
import { NgxSkeletonLoaderModule } from 'ngx-skeleton-loader';
import { TimeagoModule } from 'ngx-timeago';

@NgModule({
    declarations: [ServiceRestoresTableComponent],
    exports: [
        ServiceRestoresTableComponent,
    ],
    imports: [
        CommonModule,
        MatTableModule,
        MatButtonModule,
        ResetServiceUserPasswordDialogModule,
        NgxSkeletonLoaderModule,
        TableSkeletonModule,
        TimeagoModule,
        TimeUtcModule,
    ]
})
export class ServiceRestoresTableModule { }
