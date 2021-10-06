import { Injectable } from '@angular/core';
import { MatDialog, MatDialogRef } from '@angular/material/dialog';
import { ConfirmDialogComponent } from '@ui/confirm-dialog/confirm-dialog.component';
import { Observable } from 'rxjs';

@Injectable({
    providedIn: 'root'
})
export class ConfirmDialogService {

    private dialogRef!: MatDialogRef<ConfirmDialogComponent>;

    constructor(private dialog: MatDialog) {}

    confirm(
        title: string,
        message: string,
        buttonConfirmText = 'Yes',
        buttonCancelText = 'No',
    ): Observable<boolean> {
        const dialogData = {
            title,
            message,
            buttonConfirmText,
            buttonCancelText
        };
        this.dialogRef = this.dialog.open(ConfirmDialogComponent, {
            width: '400px',
            data: dialogData
        });

        return this.dialogRef.afterClosed();
    }
}
