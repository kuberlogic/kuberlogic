import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { ResetServiceUserPasswordDialogComponent } from '@ui/reset-service-user-password-dialog/reset-service-user-password-dialog.component';

@NgModule({
    declarations: [ResetServiceUserPasswordDialogComponent],
    imports: [
        CommonModule,
        MatDialogModule,
        MatButtonModule,
        ReactiveFormsModule,
        MatFormFieldModule,
        MatInputModule
    ],
    exports: [
        ResetServiceUserPasswordDialogComponent,
    ]
})
export class ResetServiceUserPasswordDialogModule { }
