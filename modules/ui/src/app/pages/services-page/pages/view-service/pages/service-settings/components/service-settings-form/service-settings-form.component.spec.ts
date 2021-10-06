import { ChangeDetectionStrategy, NO_ERRORS_SCHEMA, SimpleChange, SimpleChanges } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FormBuilder } from '@angular/forms';
import { ServiceBackupConfigModel } from '@models/service-backup-config.model';
import { ServiceModel, ServiceModelStatus, ServiceModelType } from '@models/service.model';
import { ServiceSettingsFormComponent } from './service-settings-form.component';

const serviceModel: ServiceModel = {
    type: ServiceModelType.POSTGRES,
    name: 'PostgreSql prod',
    status: ServiceModelStatus.FAILED,
    masters: 2,
    replicas: 5,
    created_at: '2021-02-09T10:56:12.115Z',
    internalConnection: {
        master: {
            host: 'pg-master-cloudlinux-a892.domain.com',
            port: 20990,
            user: 'cloudmanaged_admin',
            password: '*********',
            ssl_mode: 'require',
            cert: '/cert_link'
        },
        replica: {
            host: 'pg-master-cloudlinux-a892.domain.com',
            port: 20990,
            user: 'cloudmanaged_admin',
            password: '*********',
            ssl_mode: 'require',
            cert: '/cert_link'
        },
    },
    externalConnection: {
        master: {
            host: 'external-pg-master-cloudlinux-a892.domain.com',
            port: 20990,
            user: 'cloudmanaged_admin',
            password: '*********',
            ssl_mode: 'require',
            cert: '/cert_link'
        },
        replica: {
            host: 'external-pg-master-cloudlinux-a892.domain.com',
            port: 20990,
            user: 'cloudmanaged_admin',
            password: '*********',
            ssl_mode: 'require',
            cert: '/cert_link'
        },
    },
    maintenanceWindow: {
        day: 'Monday',
        startHour: 0
    },
    limits: {
        cpu: '0.5',
        memory: '1',
        volumeSize: '4'
    }
};

describe('ServiceSettingsFormComponent', () => {
    let component: ServiceSettingsFormComponent;
    let fixture: ComponentFixture<ServiceSettingsFormComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ServiceSettingsFormComponent],
            providers: [
                FormBuilder
            ],
            schemas: [NO_ERRORS_SCHEMA]
        }).overrideComponent(ServiceSettingsFormComponent, {
            set: {  changeDetection: ChangeDetectionStrategy.Default  }
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ServiceSettingsFormComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should not emit "successfulSubmit" when form is not valid', () => {
        const spy = spyOn(component.successfulSubmit, 'emit');
        component.onSubmit();
        expect(spy).not.toHaveBeenCalled();
    });

    it('should set input service to form', () => {
        component.serviceModel = serviceModel;
        fixture.detectChanges();
        component.ngOnChanges({ serviceModel: {} as SimpleChange } as SimpleChanges);
        fixture.detectChanges();
        expect(component.formGroup.value.limits).toEqual(serviceModel.limits);
    });

    it('should add service advanced config to form', () => {
        const spy = spyOn(component.advancedFormGroup, 'addControl');
        component.serviceModel = {...serviceModel, advancedConf: { test1: '1', test2: '2' }};
        fixture.detectChanges();
        component.ngOnChanges({ serviceModel: {} as SimpleChange } as SimpleChanges);
        fixture.detectChanges();
        expect(spy).toHaveBeenCalledTimes(2);
    });

    it('should remove advanced config', () => {
        const spy = spyOn(component.advancedFormGroup, 'removeControl');
        component.serviceModel = {...serviceModel, advancedConf: { test1: '1', test2: '2' }};
        fixture.detectChanges();
        component.ngOnChanges({ serviceModel: {} as SimpleChange } as SimpleChanges);
        fixture.detectChanges();

        component.removeAdvancedSetting('test1');

        expect(spy).toHaveBeenCalled();
    });
});
