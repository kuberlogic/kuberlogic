import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TableSkeletonComponent } from './table-skeleton.component';

describe('TableSkeletonComponent', () => {
    let component: TableSkeletonComponent;
    let fixture: ComponentFixture<TableSkeletonComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [TableSkeletonComponent],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(TableSkeletonComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
