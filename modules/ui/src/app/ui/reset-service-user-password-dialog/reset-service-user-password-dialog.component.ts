/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
