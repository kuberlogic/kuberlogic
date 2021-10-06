import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatTableModule } from '@angular/material/table';
import { ServiceModelStatus, ServiceModelType } from '@models/service.model';
import { ReadReplicasSelectComponent } from '@pages/services-page/pages/services-list/components/read-replicas-select/read-replicas-select.component';
import { MessagesService } from '@services/messages.service';
import { ServicesPageService } from '@services/services-page.service';
import { MockMessageService } from '@testing/mock-messages-service';
import { MockServicesPageService } from '@testing/mock-services-page-service';
import { NgxFilesizeModule } from 'ngx-filesize';
import { By } from '@angular/platform-browser';

describe('ReadReplicasSelectComponent', () => {
    let component: ReadReplicasSelectComponent;
    let fixture: ComponentFixture<ReadReplicasSelectComponent>;
    let servicesPageService: MockServicesPageService;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                MatTableModule,
                NgxFilesizeModule,
            ],
            declarations: [ReadReplicasSelectComponent],
            providers: [
                { provide: ServicesPageService, useClass: MockServicesPageService },
                { provide: MessagesService, useClass: MockMessageService },
            ],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ReadReplicasSelectComponent);
        component = fixture.componentInstance;
        // @ts-ignore
        servicesPageService = TestBed.inject(ServicesPageService);

        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should save service onSave', () => {
        const spy = spyOn(servicesPageService, 'editService').and.callThrough();
        component.service = {
            type: ServiceModelType.POSTGRES,
            name: 'postgres',
            ns: 'default',
            version: '13',
            status: ServiceModelStatus.READY,
            masters: 1,
            replicas: 1,
            created_at: '2021-07-26T14:37:31.000Z',
        };
        component.onSubmit(2);

        expect(spy).toHaveBeenCalled();
    });

    it('should disable replicas select when a service is not in ready status', () => {
       component.service = {
            type: ServiceModelType.POSTGRES,
            name: 'postgres',
            status: ServiceModelStatus.NOT_READY,
            masters: 1,
            replicas: 1,
            created_at: '2021-07-26T14:37:31.000Z',
        };
        fixture.detectChanges();
        const host = fixture.debugElement.query(By.css('mat-select'));
        expect(host.nativeElement.disabled).toBeTrue();
    });
});
