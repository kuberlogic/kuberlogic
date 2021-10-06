import { HttpClientTestingModule } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { FormBuilder } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { AuthService } from '@services/auth.service';
import { MessagesService } from '@services/messages.service';
import { MockAuthService } from '@testing/mock-auth-service';
import { MockMessageService } from '@testing/mock-messages-service';
import { LoginComponent } from './login.component';

describe('LoginComponent', () => {
    let component: LoginComponent;
    let fixture: ComponentFixture<LoginComponent>;
    let authService: MockAuthService;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                HttpClientTestingModule,
                RouterTestingModule,
            ],
            declarations: [LoginComponent],
            providers: [
                { provide: MessagesService, useClass: MockMessageService },
                { provide: AuthService, useClass: MockAuthService },
                FormBuilder,
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(LoginComponent);
        component = fixture.componentInstance;
        // @ts-ignore
        authService = TestBed.inject(AuthService);
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should call authService if form valid', () => {
        component.formGroup.patchValue({username: 'name', password: 'password'});
        fixture.detectChanges();
        const spy = spyOn(authService, 'login').and.callThrough();
        component.onLogin();
        fixture.detectChanges();

        expect(spy).toHaveBeenCalled();
    });

    it('should not call authService if form invalid', () => {
        component.formGroup.patchValue({username: 'name', password: ''});
        fixture.detectChanges();
        const spy = spyOn(authService, 'login');
        component.onLogin();
        fixture.detectChanges();

        expect(spy).not.toHaveBeenCalled();
    });
});
