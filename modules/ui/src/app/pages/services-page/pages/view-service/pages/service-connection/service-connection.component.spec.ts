import { ComponentFixture, TestBed } from '@angular/core/testing';

import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ServicesPageService } from '@services/services-page.service';
import { MockServicesPageService } from '@testing/mock-services-page-service';
import { ServiceConnectionComponent } from './service-connection.component';

describe('ServiceConnectionComponent', () => {
    let component: ServiceConnectionComponent;
    let fixture: ComponentFixture<ServiceConnectionComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ServiceConnectionComponent],
            providers: [
                { provide: ServicesPageService, useClass: MockServicesPageService }
            ],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ServiceConnectionComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
