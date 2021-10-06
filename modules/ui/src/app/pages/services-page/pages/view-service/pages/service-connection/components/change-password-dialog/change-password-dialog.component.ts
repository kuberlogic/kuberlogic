import { ChangeDetectionStrategy, Component, Inject } from '@angular/core';
import { AbstractControlOptions, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { FormContainerMixin } from '@app/mixins/form-container.mixin';
import { BaseObject } from '@app/mixins/mixins';
import { MustMatch } from '@app/helpers/must-match.validator';

@Component({
    selector: 'kl-change-password-dialog',
    templateUrl: './change-password-dialog.component.html',
    styleUrls: ['./change-password-dialog.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ChangePasswordDialogComponent extends FormContainerMixin(BaseObject) {
    formGroup: FormGroup;
    constructor(
        private fb: FormBuilder,
        private dialogRef: MatDialogRef<any>,
    ) {
        super();
        const formOptions: AbstractControlOptions = {
            validators: MustMatch('password', 'confirmPassword')
        };

        this.formGroup = this.fb.group({
            password: ['', [
                    Validators.required,
                    Validators.pattern(/^(?=.*\d)(?=.*[a-z])(?=.*[A-Z])(?=.*[a-zA-Z]).{8,}$/)
                ]
            ],
            confirmPassword    : ['', [
                    Validators.required,
                    Validators.pattern(/^(?=.*\d)(?=.*[a-z])(?=.*[A-Z])(?=.*[a-zA-Z]).{8,}$/)
                ]
            ],
        }, formOptions);
    }

    onSave(): any {
        if (this.checkForm()) {
            this.dialogRef.close(this.formGroup.value.password);
        }
    }

}
