import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ServicesPageService } from '@services/services-page.service';
import { MockServicesPageService } from '@testing/mock-services-page-service';
import { ServicesListComponent } from './services-list.component';

describe('ServicesListComponent', () => {
    let component: ServicesListComponent;
    let fixture: ComponentFixture<ServicesListComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ServicesListComponent],
            providers: [
                { provide: ServicesPageService, useClass: MockServicesPageService }
            ],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ServicesListComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
