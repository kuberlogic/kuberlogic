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

import { Component, EventEmitter, OnInit, Output } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { FormContainerMixin } from '@app/mixins/form-container.mixin';
import { BaseObject } from '@app/mixins/mixins';
import { ServiceModel, ServiceModelType } from '@models/service.model';
import { MessagesService } from '@services/messages.service';
import { limitFormGroup } from '@services/services-page.service';
import { RadioGroupModel } from '@ui/radio-group/radio-group.model';

@Component({
    selector: 'kl-create-service-form',
    styleUrls: ['./create-service-form.component.scss'],
    templateUrl: './create-service-form.component.html',
})
export class CreateServiceFormComponent extends FormContainerMixin(BaseObject) implements OnInit {
    formGroup: FormGroup;
    @Output() successfulSubmit = new EventEmitter<Partial<ServiceModel>>();

    readonly serviceTypes: RadioGroupModel[] = [
        { title: 'MySQL', value: ServiceModelType.MYSQL, svgIcon: 'mysqlIcon' },
        { title: 'PostgreSQL', value: ServiceModelType.POSTGRES, svgIcon: 'postgresqlIcon' },
    ];

    readonly serviceVersions: {[key: string]: string[]} = {
        [ServiceModelType.MYSQL]: [
            '5.7.31',
        ],
        [ServiceModelType.POSTGRES]: [
            '9.5',
            '9.6',
            '10',
            '11',
            '12',
            '13',
        ],
    };
    previousVersions: {[key: string]: string} = {};

    constructor(
        private fb: FormBuilder,
        private messages: MessagesService,
    ) {
        super();
        this.formGroup = this.fb.group({
            type: ['', [Validators.required]],
            ns: ['default'],
            replicas: [1],
            name: ['', [Validators.required]],
            version: ['', [Validators.required]],
            ...limitFormGroup,
        });
    }

    ngOnInit(): void {
        this.formGroup.get('type')?.valueChanges.subscribe((value) => {
            const version = this.previousVersions[value]
                ? this.previousVersions[value]
                : value === ServiceModelType.MYSQL
                    ? this.serviceVersions[ServiceModelType.MYSQL][0]
                    : undefined;
            this.formGroup.get('version')?.patchValue(version);
        });
        this.formGroup.get('version')?.valueChanges.subscribe((value) => {
            this.previousVersions[this.formGroup.controls.type.value] = value;
        });
    }

    onSave(): void {
        if (this.checkForm()) {
            const values = this.formGroup.value;

            const serviceModel: Partial<ServiceModel> = {
                type: values.type,
                name: values.name,
                replicas: values.replicas,
                ns: values.ns,
                version: values.version,
                limits: {
                    cpu: values.cpu,
                    memory: values.memory,
                    volumeSize: values.volumeSize,
                }
            } as Partial<ServiceModel>;

            this.successfulSubmit.emit(serviceModel);
        } else {
            this.messages.error('Please fill out the form correctly');
        }
    }

}
