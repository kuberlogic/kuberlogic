import { ChangeDetectionStrategy, Component, Inject } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { FormContainerMixin } from '@app/mixins/form-container.mixin';
import { BaseObject } from '@app/mixins/mixins';
import { MessagesService } from '@services/messages.service';
import { ServiceUsersService } from '@services/service-users.service';
import { ServicesPageService } from '@services/services-page.service';

@Component({
    selector: 'kl-reset-service-user-password-dialog',
    templateUrl: './reset-service-user-password-dialog.component.html',
    styleUrls: ['./reset-service-user-password-dialog.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ResetServiceUserPasswordDialogComponent extends FormContainerMixin(BaseObject) {
    formGroup: FormGroup;

    constructor(
        private fb: FormBuilder,
        private messages: MessagesService,
        private usersService: ServiceUsersService,
        private servicesPageService: ServicesPageService,
        private dialogRef: MatDialogRef<ResetServiceUserPasswordDialogComponent>,
        @Inject(MAT_DIALOG_DATA) public name: string,
    ) {
        super();
        this.formGroup = this.fb.group({
            password: ['', [Validators.required, Validators.minLength(8)]],
        });
    }

    onSave(): void {
        if (this.checkForm()) {
            const user = { name: this.name, password: this.formGroup.value.password };
            this.usersService.editUser(this.servicesPageService.getCurrentServiceId(), user).subscribe(
                () => {
                    this.messages.success('Password was successfully changed');
                    this.dialogRef.close();
                }
            );
        }
    }

}
