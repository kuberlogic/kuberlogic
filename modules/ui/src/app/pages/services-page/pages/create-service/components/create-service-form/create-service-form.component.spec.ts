import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { FormBuilder } from '@angular/forms';
import { ServiceModelType } from '@models/service.model';
import { MessagesService } from '@services/messages.service';
import { MockMessageService } from '@testing/mock-messages-service';
import { CreateServiceFormComponent } from './create-service-form.component';

const MockData: any = {
    type: ServiceModelType.POSTGRES,
    name: 'postgres',
    ns: 'default',
    version: '13',
    cpu: 1,
    memory: '10',
    volumeSize: '10'
};

describe('CreateServiceFormComponent', () => {
    let component: CreateServiceFormComponent;
    let fixture: ComponentFixture<CreateServiceFormComponent>;
    let messagesService: MockMessageService;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [CreateServiceFormComponent],
            providers: [
                FormBuilder,
                { provide: MessagesService, useClass: MockMessageService }
            ],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(CreateServiceFormComponent);
        component = fixture.componentInstance;
        messagesService = TestBed.inject(MessagesService);
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should show error when form is invalid', () => {
        const spy = spyOn(messagesService, 'error');

        component.formGroup.patchValue({
            type: ServiceModelType.MYSQL
        });
        component.onSave();
        fixture.detectChanges();

        expect(spy).toHaveBeenCalled();
    });

    it('should change versions list if type changed', () => {
        // component.formGroup.patchValue(MockData);
        fixture.detectChanges();
        component.ngOnInit();
        component.formGroup.controls.type.setValue(ServiceModelType.POSTGRES);
        fixture.detectChanges();
        expect(component.formGroup.controls.version.value).toEqual(undefined);
        component.formGroup.controls.version.setValue('11');
        fixture.detectChanges();
        component.formGroup.controls.type.setValue(ServiceModelType.MYSQL);
        fixture.detectChanges();
        expect(component.formGroup.controls.version.value).toEqual('5.7.31');
        fixture.detectChanges();
        component.formGroup.controls.type.setValue(ServiceModelType.POSTGRES);
        fixture.detectChanges();
        expect(component.formGroup.controls.version.value).toEqual('11');
    });

    it('should emit successfulSubmit when form is valid', () => {
        const spy = spyOn(component.successfulSubmit, 'emit').and.callThrough();

        component.formGroup.patchValue(MockData);
        fixture.detectChanges();
        component.ngOnInit();
        component.onSave();
        fixture.detectChanges();

        expect(spy).toHaveBeenCalled();
    });
});
