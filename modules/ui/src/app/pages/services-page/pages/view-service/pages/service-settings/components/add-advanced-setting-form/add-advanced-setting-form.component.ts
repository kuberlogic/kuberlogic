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

import { ChangeDetectionStrategy, Component, EventEmitter, Output } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { FormContainerMixin } from '@app/mixins/form-container.mixin';
import { BaseObject } from '@app/mixins/mixins';

export interface AddAdvancedSettingFormResult {
    key: string;
    value: string;
}

@Component({
    selector: 'kl-add-advanced-setting-form',
    templateUrl: './add-advanced-setting-form.component.html',
    styleUrls: ['./add-advanced-setting-form.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class AddAdvancedSettingFormComponent extends FormContainerMixin(BaseObject) {
    formGroup: FormGroup;

    @Output() successfulSubmit = new EventEmitter<AddAdvancedSettingFormResult>();

    constructor(
        private fb: FormBuilder,
    ) {
        super();
        this.formGroup = this.fb.group({
            key: ['', [Validators.required, Validators.pattern(/^\S+$/)]],
            value: ['', [Validators.required]],
        });
    }

    onSubmit(): void {
        if (this.checkForm()) {
            this.successfulSubmit.emit(this.formGroup.value);
            this.formGroup.patchValue({ key: '', value: ''});
            this.formGroup.get('key')?.setErrors(null);
            this.formGroup.get('value')?.setErrors(null);
        }
    }

}
