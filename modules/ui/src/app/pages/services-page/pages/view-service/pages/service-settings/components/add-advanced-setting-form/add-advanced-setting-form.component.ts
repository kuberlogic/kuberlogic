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
