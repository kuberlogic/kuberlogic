import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { AddAdvancedSettingFormComponent } from './add-advanced-setting-form.component';

@NgModule({
    declarations: [AddAdvancedSettingFormComponent],
    exports: [
        AddAdvancedSettingFormComponent
    ],
    imports: [
        CommonModule,
        MatFormFieldModule,
        MatInputModule,
        ReactiveFormsModule,
        MatButtonModule
    ]
})
export class AddAdvancedSettingFormModule { }
