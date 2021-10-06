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
