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

import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { FormBuilder } from '@angular/forms';

import { AddAdvancedSettingFormComponent } from './add-advanced-setting-form.component';

describe('AddAdvancedSettingFormComponent', () => {
    let component: AddAdvancedSettingFormComponent;
    let fixture: ComponentFixture<AddAdvancedSettingFormComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [AddAdvancedSettingFormComponent],
            providers: [
                FormBuilder,
            ],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(AddAdvancedSettingFormComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should not emit "successfulSubmit" when form is invalid', () => {
        const spy = spyOn(component.successfulSubmit, 'emit');
        component.formGroup.patchValue({ key: '', value: ''});
        fixture.detectChanges();

        component.onSubmit();

        expect(spy).not.toHaveBeenCalled();
    });

    it('should emit "successfulSubmit" when form is valid', () => {
        const spy = spyOn(component.successfulSubmit, 'emit');
        component.formGroup.patchValue({ key: 'key', value: 'value'});
        fixture.detectChanges();

        component.onSubmit();

        expect(spy).toHaveBeenCalledWith({ key: 'key', value: 'value'});
    });
});
