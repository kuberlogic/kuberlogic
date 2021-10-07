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
import { ServiceBackupConfigModel } from '@models/service-backup-config.model';

@Component({
    selector: 'kl-backup-storage-dialog',
    templateUrl: './backup-storage-dialog.component.html',
    styleUrls: ['./backup-storage-dialog.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class BackupStorageDialogComponent extends FormContainerMixin(BaseObject) {
    formGroup: FormGroup;
    constructor(
        private fb: FormBuilder,
        private dialogRef: MatDialogRef<BackupStorageDialogComponent>,
        @Inject(MAT_DIALOG_DATA) data: Partial<ServiceBackupConfigModel>,
    ) {
        super();
        this.formGroup = this.fb.group({
            aws_access_key_id: [data.aws_access_key_id, [Validators.required]],
            aws_secret_access_key: [data.aws_secret_access_key, [Validators.required]],
            bucket: [data.bucket, [Validators.required]],
            endpoint: [data.endpoint, [Validators.required]],
        });
    }

    onSave(): void {
        if (this.checkForm()) {
            this.dialogRef.close(this.formGroup.value);
        }
    }

}
