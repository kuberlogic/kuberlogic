import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatSelectModule } from '@angular/material/select';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { BackupFormModule } from '@pages/services-page/pages/view-service/pages/service-backups/components/backup-form/backup-form.module';
import { BackupStorageDialogModule } from '@pages/services-page/pages/view-service/pages/service-backups/components/backup-storage-dialog/backup-storage-dialog.module';
import { RestoreServiceBackupsDialogModule } from '@pages/services-page/pages/view-service/pages/service-backups/components/restore-service-backups-dialog/restore-service-backups-dialog.module';
import { ServiceBackupsTableModule } from '@pages/services-page/pages/view-service/pages/service-backups/components/service-backups-table/service-backups-table.module';
import { ServiceRestoresTableModule } from '@pages/services-page/pages/view-service/pages/service-backups/components/service-restores-table/service-restores-table.module';
import { ServiceBackupsRoutingModule } from '@pages/services-page/pages/view-service/pages/service-backups/service-backups-routing.module';
import { ServiceBackupsComponent } from './service-backups.component';

@NgModule({
    declarations: [
        ServiceBackupsComponent,
    ],
    imports: [
        CommonModule,
        MatIconModule,
        ServiceBackupsRoutingModule,
        ServiceBackupsTableModule,
        ServiceRestoresTableModule,
        RestoreServiceBackupsDialogModule,
        BackupStorageDialogModule,
        ReactiveFormsModule,
        MatFormFieldModule,
        MatSelectModule,
        MatSlideToggleModule,
        BackupFormModule,
    ]
})
export class ServiceBackupsModule { }
