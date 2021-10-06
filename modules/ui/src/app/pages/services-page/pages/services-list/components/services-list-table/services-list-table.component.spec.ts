import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ServicesListTableComponent } from './services-list-table.component';

describe('ServicesListTableComponent', () => {
    let component: ServicesListTableComponent;
    let fixture: ComponentFixture<ServicesListTableComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ServicesListTableComponent],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ServicesListTableComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
