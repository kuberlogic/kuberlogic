import { HttpClientTestingModule } from '@angular/common/http/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, fakeAsync, TestBed, tick } from '@angular/core/testing';
import { FormBuilder } from '@angular/forms';
import { ServiceLogsFormComponent } from '@pages/services-page/pages/view-service/pages/service-logs/components/service-logs-form/service-logs-form.component';
import { MessagesService } from '@services/messages.service';
import { ServicesPageService } from '@services/services-page.service';
import { MockMessageService } from '@testing/mock-messages-service';
import { MockServicesPageService } from '@testing/mock-services-page-service';

describe('ServiceLogsFormComponent', () => {
    let component: ServiceLogsFormComponent;
    let fixture: ComponentFixture<ServiceLogsFormComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ServiceLogsFormComponent],
            providers: [
                FormBuilder,
                { provide: ServicesPageService, useClass: MockServicesPageService },
                { provide: MessagesService, useClass: MockMessageService },
            ],
            schemas: [NO_ERRORS_SCHEMA],
            imports: [HttpClientTestingModule],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ServiceLogsFormComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
        fixture.detectChanges();
    });

    it('should emit "successfulSubmit" on submit', () => {
        component.formGroup.patchValue({serviceInstance: 'serviceInstance'});
        fixture.detectChanges();
        const spy = spyOn(component.successfulSubmit, 'emit');
        component.onSubmit();
        fixture.detectChanges();

        expect(spy).toHaveBeenCalled();
    });

    it('should get new dataSource on init and call form submit', () => {
        component.formGroup.patchValue({serviceInstance: 'serviceInstance'});
        fixture.detectChanges();
        const spy = spyOn(component.successfulSubmit, 'emit');
        component.ngOnInit();
        fixture.detectChanges();
        expect(spy).toHaveBeenCalled();
    });

    it('should not fail if datasource is undefined', () => {
        component.formGroup.patchValue({serviceInstance: 'serviceInstance'});
        fixture.detectChanges();
        const spy = spyOn(component.successfulSubmit, 'emit');
        component.selectFirstInstance(undefined);
        fixture.detectChanges();
        expect(spy).not.toHaveBeenCalled();
    });
});
