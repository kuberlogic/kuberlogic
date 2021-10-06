import { ChangeDetectionStrategy, Component, Inject } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { ConfirmDialogModel } from '@ui/confirm-dialog/confirm-dialog.model';

@Component({
    selector: 'kl-confirm-dialog',
    templateUrl: './confirm-dialog.component.html',
    styleUrls: ['./confirm-dialog.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ConfirmDialogComponent {
    title = '';
    message = '';
    buttonConfirmText = '';
    buttonCancelText = '';

    constructor(
        private dialogRef: MatDialogRef<ConfirmDialogComponent>,
        @Inject(MAT_DIALOG_DATA) private data: ConfirmDialogModel
    ) {
        this.title = data.title;
        this.message = data.message;
        this.buttonConfirmText = data.buttonConfirmText;
        this.buttonCancelText = data.buttonCancelText;
    }

    onConfirm(): void {
        this.dialogRef.close(true);
    }

    onDismiss(): void {
        this.dialogRef.close(false);
    }

}
