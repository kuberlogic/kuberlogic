import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { ReadReplicasSelectComponent } from '@pages/services-page/pages/services-list/components/read-replicas-select/read-replicas-select.component';

@NgModule({
    declarations: [
        ReadReplicasSelectComponent
    ],
    exports: [
        ReadReplicasSelectComponent
    ],
    imports: [
        CommonModule,
        MatFormFieldModule,
        ReactiveFormsModule,
        MatSelectModule,
    ]
})
export class ReadReplicasSelectModule { }
