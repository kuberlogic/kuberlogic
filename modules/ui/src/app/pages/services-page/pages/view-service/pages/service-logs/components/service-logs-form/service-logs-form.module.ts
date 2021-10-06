import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';

import { ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { ServiceLogsFormComponent } from '@pages/services-page/pages/view-service/pages/service-logs/components/service-logs-form/service-logs-form.component';

@NgModule({
    declarations: [ServiceLogsFormComponent],
    exports: [
        ServiceLogsFormComponent,
    ],
    imports: [
        CommonModule,
        MatIconModule,
        ReactiveFormsModule,
        MatFormFieldModule,
        MatInputModule,
        MatButtonModule,
        MatSelectModule,
    ]
})
export class ServiceLogsFormModule { }
