import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';

import { RadioGroupComponent } from './radio-group.component';

describe('RadioGroupComponent', () => {
    let component: RadioGroupComponent;
    let fixture: ComponentFixture<RadioGroupComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [RadioGroupComponent],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(RadioGroupComponent);
        component = fixture.componentInstance;
        component.selectors = [
            { title: 'Val1', value: 'val1' },
            { title: 'Val2', value: 'val2' },
        ];
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should select value on click', () => {
        const firstItem = fixture.debugElement.query(By.css('.radio-group__item'));
        firstItem.nativeElement.click();
        fixture.detectChanges();

        expect(component.value).toEqual('val1');
    });
});
