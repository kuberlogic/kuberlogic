import { AbstractControl, FormGroup } from '@angular/forms';
import { Constructor } from '@app/mixins/mixins';

export interface ServerErrors {
    [id: string]: string;
}

// eslint-disable-next-line
export function FormContainerMixin<TBase extends Constructor>(Base: TBase) {
    abstract class FormContainerMixinClass extends Base {
        protected abstract formGroup: FormGroup;
        private checked = false;

        isInvalid(controlName: string): boolean {
            const control = this.formGroup.controls[controlName];
            if (!control) {
                return false;
            }

            return control.touched && !control.valid;
        }

        get isFormInvalid(): boolean {
            return !this.formGroup.valid && this.checked;
        }

        protected checkForm(): boolean {
            this.checked = true;
            this.formGroup.updateValueAndValidity();
            this.markFormGroupTouched();

            return this.formGroup.valid;
        }

        protected markUntouched(): void {
            const controls = this.getAllControls(this.formGroup);
            controls.forEach((control) => {
                control.markAsUntouched({ onlySelf: true });
            });
        }

        private markFormGroupTouched(): void {
            const controls = this.getAllControls(this.formGroup);
            controls.forEach((control) => {
                control.markAsTouched({ onlySelf: true });
                control.updateValueAndValidity({ onlySelf: true });
            });
            this.formGroup.updateValueAndValidity({ onlySelf: true });
        }

        private getAllControls(formGroup: FormGroup): AbstractControl[] {
            const result = [] as AbstractControl[];
            if (!formGroup.controls) {
                return [];
            }

            const controls: AbstractControl[] = Array.isArray(formGroup.controls) ?
                formGroup.controls :
                (Object as any).values(formGroup.controls);

            controls.forEach((control) => {
                // @ts-ignore
                if (!!control.controls) {
                    // @ts-ignore
                    const childControls = this.getAllControls(control);
                    result.push(...childControls);
                } else {
                    result.push(control);
                }
            });

            return result;
        }
    }

    return FormContainerMixinClass;
}
