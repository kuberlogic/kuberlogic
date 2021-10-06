import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { MatInputModule } from '@angular/material/input';
import { MatRadioModule } from '@angular/material/radio';
import { MatSelectModule } from '@angular/material/select';
import { RadioGroupModule } from '@ui/radio-group/radio-group.module';
import { CreateServiceFormComponent } from './create-service-form.component';

@NgModule({
    declarations: [CreateServiceFormComponent],
    exports: [
        CreateServiceFormComponent
    ],
    imports: [
        CommonModule,
        ReactiveFormsModule,
        MatRadioModule,
        MatSelectModule,
        MatInputModule,
        RadioGroupModule
    ]
})
export class CreateServiceFormModule { }
