import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatDialogModule } from '@angular/material/dialog';
import { ConfirmDialogComponent } from './confirm-dialog.component';

@NgModule({
    declarations: [ConfirmDialogComponent],
    imports: [
        CommonModule,
        MatButtonModule,
        MatDialogModule,
    ]
})
export class ConfirmDialogModule { }
