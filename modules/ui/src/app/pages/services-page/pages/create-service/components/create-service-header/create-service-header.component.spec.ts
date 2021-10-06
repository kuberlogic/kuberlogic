import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';

import { HarnessLoader } from '@angular/cdk/testing';
import { TestbedHarnessEnvironment } from '@angular/cdk/testing/testbed';
import { MatButtonHarness } from '@angular/material/button/testing';
import { CreateServiceHeaderComponent } from './create-service-header.component';

describe('CreateServiceHeaderComponent', () => {
    let component: CreateServiceHeaderComponent;
    let fixture: ComponentFixture<CreateServiceHeaderComponent>;
    let loader: HarnessLoader;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [RouterTestingModule],
            declarations: [CreateServiceHeaderComponent],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(CreateServiceHeaderComponent);
        component = fixture.componentInstance;
        loader = TestbedHarnessEnvironment.loader(fixture);
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should emit "createPressed" when Create button is clicked', async () => {
        const spy = spyOn(component.createPressed, 'emit');

        const createButton = await loader.getHarness(
            MatButtonHarness.with({selector: '.create-service-header__create-button'})
        );

        await createButton.click();

        expect(spy).toHaveBeenCalled();
    });
});
