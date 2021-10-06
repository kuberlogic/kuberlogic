import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TimeagoModule } from 'ngx-timeago';
import { TimeUtcComponent } from './time-utc.component';

describe('TimeUtcComponent', () => {
    let component: TimeUtcComponent;
    let fixture: ComponentFixture<TimeUtcComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                TimeagoModule.forRoot(),
            ],
            declarations: [
                TimeUtcComponent,
            ],
        })
            .compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(TimeUtcComponent);
        component = fixture.componentInstance;
        component.timestamp = '2021-06-04T00:00:13.794Z';
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
