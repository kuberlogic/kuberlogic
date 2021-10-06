import { ComponentFixture, TestBed } from '@angular/core/testing';

import { NO_ERRORS_SCHEMA } from '@angular/core';
import { RouterTestingModule } from '@angular/router/testing';
import { ServicesPageService } from '@services/services-page.service';
import { MockServicesPageService } from '@testing/mock-services-page-service';
import { ViewServiceComponent } from './view-service.component';

describe('EditServiceComponent', () => {
    let component: ViewServiceComponent;
    let fixture: ComponentFixture<ViewServiceComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [RouterTestingModule],
            declarations: [ViewServiceComponent],
            providers: [
                { provide: ServicesPageService, useClass: MockServicesPageService }
            ],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ViewServiceComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
