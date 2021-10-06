import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ServiceRestoresTableComponent } from './service-restores-table.component';

describe('ServiceRestoresTableComponent', () => {
    let component: ServiceRestoresTableComponent;
    let fixture: ComponentFixture<ServiceRestoresTableComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [],
            declarations: [ServiceRestoresTableComponent],
            providers: [],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ServiceRestoresTableComponent);
        component = fixture.componentInstance;
        component.restores = [
            {
                file: 's3://test/postgresql/kuberlogic-kl-pg/logical_backups/1622729871.sql.gz',
                database: 'db1',
                time: '2021-06-04T08:00:14.000Z',
                status: 'Failed',
            },
        ];
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
