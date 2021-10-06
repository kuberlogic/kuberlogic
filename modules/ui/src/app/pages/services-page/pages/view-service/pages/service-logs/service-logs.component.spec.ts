import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';

import { HttpClientTestingModule } from '@angular/common/http/testing';
import { MessagesService } from '@services/messages.service';
import { ServiceLogsService } from '@services/service-logs.service';
import { MockMessageService } from '@testing/mock-messages-service';
import { MockServiceLogsService } from '@testing/mock-service-logs-service';
import { ServiceLogsComponent } from './service-logs.component';

const typesMock: {[key: string]: string} = {
    '': '',
    postgresql: 'PostgreSQL',
    mysql: 'MySQL',
    other: 'other',
};

describe('ServiceLogsComponent', () => {
    let component: ServiceLogsComponent;
    let fixture: ComponentFixture<ServiceLogsComponent>;
    let logsService: MockServiceLogsService;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ServiceLogsComponent],
            schemas: [NO_ERRORS_SCHEMA],
            imports: [HttpClientTestingModule],
            providers: [
                { provide: ServiceLogsService, useClass: MockServiceLogsService },
                { provide: MessagesService, useClass: MockMessageService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ServiceLogsComponent);
        component = fixture.componentInstance;
        // @ts-ignore
        logsService = TestBed.inject(ServiceLogsService);
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    Object.keys(typesMock).map((type) => {
        it(`should render type for string "${typesMock[type]}"`, () => {
            expect(component.renderType(type)).toEqual(typesMock[type]);
        });
    });

    it(`should render type for undefined`, () => {
        expect(component.renderType(undefined)).toEqual(undefined);
    });

    it('should get new logs on form submit', () => {
        const spy = spyOn(logsService, 'get').and.callThrough();
        component.onFormSubmit('some_username');
        fixture.detectChanges();

        expect(spy).toHaveBeenCalled();
    });
});
