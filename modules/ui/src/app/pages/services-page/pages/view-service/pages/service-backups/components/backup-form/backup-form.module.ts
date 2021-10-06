import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { BackupFormComponent } from '@pages/services-page/pages/view-service/pages/service-backups/components/backup-form/backup-form.component';
import { ScheduleSelectModule } from '@pages/services-page/pages/view-service/pages/service-backups/components/schedule-select/schedule-select.module';
import { AddAdvancedSettingFormModule } from '@pages/services-page/pages/view-service/pages/service-settings/components/add-advanced-setting-form/add-advanced-setting-form.module';

@NgModule({
    declarations: [
        BackupFormComponent,
    ],
    exports: [
        BackupFormComponent,
    ],
    imports: [
        CommonModule,
        MatButtonModule,
        ReactiveFormsModule,
        MatInputModule,
        MatSlideToggleModule,
        MatSelectModule,
        AddAdvancedSettingFormModule,
        MatIconModule,
        ScheduleSelectModule
    ]
})
export class BackupFormModule { }
