import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { AddAdvancedSettingFormModule } from '@pages/services-page/pages/view-service/pages/service-settings/components/add-advanced-setting-form/add-advanced-setting-form.module';
import { ServiceSettingsFormComponent } from './service-settings-form.component';

@NgModule({
    declarations: [ServiceSettingsFormComponent],
    exports: [
        ServiceSettingsFormComponent
    ],
    imports: [
        CommonModule,
        MatButtonModule,
        ReactiveFormsModule,
        MatInputModule,
        MatSlideToggleModule,
        MatSelectModule,
        AddAdvancedSettingFormModule,
        MatIconModule
    ]
})
export class ServiceSettingsFormModule { }
