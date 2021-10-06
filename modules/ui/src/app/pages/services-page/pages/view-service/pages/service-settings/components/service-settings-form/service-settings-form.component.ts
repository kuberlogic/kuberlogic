import {
    ChangeDetectionStrategy, ChangeDetectorRef, Component,
    EventEmitter, Input, OnChanges, Output, SimpleChanges,
} from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { FormContainerMixin } from '@app/mixins/form-container.mixin';
import { BaseObject } from '@app/mixins/mixins';
import { ServiceBackupConfigModel } from '@models/service-backup-config.model';
import { ServiceModel, ServiceModelStatus } from '@models/service.model';
import { AddAdvancedSettingFormResult } from '@pages/services-page/pages/view-service/pages/service-settings/components/add-advanced-setting-form/add-advanced-setting-form.component';
import { limitFormGroup } from '@services/services-page.service';

export interface ServiceSettingsFormResult {
    service?: ServiceModel;
    backup?: ServiceBackupConfigModel;
}

@Component({
    selector: 'kl-service-settings-form',
    templateUrl: './service-settings-form.component.html',
    styleUrls: ['./service-settings-form.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ServiceSettingsFormComponent extends FormContainerMixin(BaseObject) implements OnChanges {
    @Input() serviceModel!: ServiceModel;
    @Input() isSaving = false;
    @Output() successfulSubmit = new EventEmitter<ServiceSettingsFormResult>();

    formGroup: FormGroup;
    advancedFormGroup: FormGroup;

    limitsFormGroup: FormGroup;
    maintenanceWindowFormGroup: FormGroup;

    canBeEdited = true;

    readonly hours = [...Array(24).keys()];
    readonly weekdays = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];

    constructor(
        private fb: FormBuilder,
        private cdRef: ChangeDetectorRef
    ) {
        super();
        this.advancedFormGroup = this.fb.group({});

        this.limitsFormGroup = this.fb.group(limitFormGroup);
        this.maintenanceWindowFormGroup = this.fb.group({
            day: ['Sunday'],
            startHour: [''],
        });

        this.formGroup = this.fb.group({
            automatic_upgrades: [false, Validators.required],
            maintenanceWindow: this.maintenanceWindowFormGroup,
            limits: this.limitsFormGroup,
            advancedConf: this.advancedFormGroup,
        });
    }

    ngOnChanges(changes: SimpleChanges): void {
        if (changes.serviceModel && this.serviceModel) {
            this.setupServiceModel(this.serviceModel);
        }
    }

    addAdvancedSetting(setting: AddAdvancedSettingFormResult): void {
        this.advancedFormGroup.addControl(setting.key, this.fb.control(setting.value, [Validators.required]));
    }

    removeAdvancedSetting(settingKey: string): void {
        this.advancedFormGroup.removeControl(settingKey);
    }

    get advancedFormGroupControls(): string[] {
        return Object.keys(this.advancedFormGroup.controls);
    }

    getHourLabel(h: number): string {
        return `${h < 10 ? '0' : ''}${h}:00`;
    }

    onSubmit(): void {
        if (this.checkForm()) {
            this.successfulSubmit.emit({
                service: this.formGroup.value,
            });
        }
    }

    private setupServiceModel(serviceModel: ServiceModel): void {
        if (serviceModel?.maintenanceWindow?.day === '') {
            serviceModel.maintenanceWindow.day = 'Sunday';
        }

        this.formGroup.patchValue(serviceModel);

        if (!!serviceModel.advancedConf) {
            const settingKeys = Object.keys(serviceModel.advancedConf);
            settingKeys.forEach((key) => {
                // @ts-ignore
                this.addAdvancedSetting({ key, value: serviceModel.advancedConf[key] });
            });
        }

        this.canBeEdited = [ServiceModelStatus.RUNNING, ServiceModelStatus.READY].includes(serviceModel.status);
        this.cdRef.detectChanges();
    }

}
