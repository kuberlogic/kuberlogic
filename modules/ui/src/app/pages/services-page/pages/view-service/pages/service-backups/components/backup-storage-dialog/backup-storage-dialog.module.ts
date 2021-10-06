import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { BackupStorageDialogComponent } from '@pages/services-page/pages/view-service/pages/service-backups/components/backup-storage-dialog/backup-storage-dialog.component';

@NgModule({
    declarations: [BackupStorageDialogComponent],
    imports: [
        CommonModule,
        MatDialogModule,
        MatButtonModule,
        ReactiveFormsModule,
        MatFormFieldModule,
        MatInputModule,
        MatSelectModule,
    ],
    exports: [
        BackupStorageDialogComponent
    ]
})
export class BackupStorageDialogModule { }
